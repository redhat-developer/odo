/*
Copyright 2020 The Knative Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package metrics

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"testing"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/stats/view"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	logtesting "knative.dev/pkg/logging/testing"
)

func TestOpenCensusConfig(t *testing.T) {
	cert, err := ioutil.ReadFile(filepath.Join("testdata", "client-cert.pem"))
	if err != nil {
		t.Fatalf("Couldn't find testdata/client-cert.pem: %v", err)
	}
	key, err := ioutil.ReadFile(filepath.Join("testdata", "client-key.pem"))
	if err != nil {
		t.Fatalf("Couldn't find testdata/client-key.pem: %v", err)
	}

	cases := []struct {
		desc     string
		config   metricsConfig
		tls      *tls.Config
		err      error
		wantFunc func(*testing.T, view.Exporter)
	}{{
		desc: "No TLS mostly default",
		config: metricsConfig{
			domain:             "test",
			component:          "test",
			backendDestination: openCensus,
		},
		wantFunc: func(t *testing.T, v view.Exporter) {
			if v == nil {
				t.Error("Expected view to be non-nil")
			}
		},
	}, {
		desc: "With TLS",

		config: metricsConfig{
			domain:             "secure",
			component:          "test",
			backendDestination: openCensus,
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-opencensus",
				},
				Data: map[string][]byte{
					"client-cert.pem": cert,
					"client-key.pem":  key,
				},
			},
			requireSecure: true,
		},
		tls: &tls.Config{},
		wantFunc: func(t *testing.T, v view.Exporter) {
			if v == nil {
				t.Error("Expected view to be non-nil")
			}
			oc, ok := v.(*ocagent.Exporter)
			if !ok {
				t.Errorf("Did not get an OpenCensus exporter: %+v", v)
			}
			oc.Flush()
		},
	}}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var server net.Listener
			var shutdown chan error
			var err error
			if c.err == nil {
				server, shutdown, err = GetServer(c.tls)
				if err != nil {
					t.Fatalf("Failed to start server: %v", err)
				}
				c.config.collectorAddress = server.Addr().String()
			}

			got, gotErr := newOpenCensusExporter(&c.config, logtesting.TestLogger(t))
			if c.err != nil {
				if diff := cmp.Diff(c.err, gotErr); diff != "" {
					t.Errorf("wrong err (-want +got) = %v", diff)
				}
				return
			}
			if gotErr != nil {
				t.Errorf("unexpected err: %v", gotErr)
				return
			}
			if c.wantFunc != nil {
				c.wantFunc(t, got)
			}

			t.Logf("Awaiting channel shutdown at %s", server.Addr().String())
			err = <-shutdown
			if err != nil {
				t.Errorf("Error from server: %v", err)
			}
			err = server.Close()
			if err != nil {
				t.Errorf("Failed to shut down server: %v", err)
			}
		})
	}
}

type fakeSecrets struct {
	secrets []corev1.Secret
}

func fakeSecretList(s ...corev1.Secret) *fakeSecrets {
	return &fakeSecrets{secrets: s}
}

func (f *fakeSecrets) Get(name string) (*corev1.Secret, error) {
	for _, s := range f.secrets {
		if fmt.Sprintf("%s/%s", s.Namespace, s.Name) == name {
			return &s, nil
		}

		if s.Name == name {
			return &s, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{Resource: "secrets"}, name)
}

func GetServer(config *tls.Config) (net.Listener, chan error, error) {
	var server net.Listener
	var err error
	if config == nil {
		server, err = net.Listen("tcp", "localhost:0")
	} else {
		if config.Certificates == nil {
			serverCert, err := tls.LoadX509KeyPair(
				filepath.Join("testdata", "server-cert.pem"), filepath.Join("testdata", "server-key.pem"))
			if err != nil {
				return nil, nil, fmt.Errorf("Unable to load server cert from testadata: %w", err)
			}
			config.Certificates = []tls.Certificate{serverCert}
		}
		server, err = tls.Listen("tcp", "localhost:0", config)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create listen server: %w", err)
	}
	shutdown := make(chan error)
	go func() {
		c, err := server.Accept()
		if err != nil {
			shutdown <- fmt.Errorf("Failed to accept connection: %w", err)
			return
		}
		err = c.Close()
		if err != nil {
			shutdown <- fmt.Errorf("Failed to close server connection: %w", err)
			return
		}
		shutdown <- nil
	}()
	return server, shutdown, err
}
