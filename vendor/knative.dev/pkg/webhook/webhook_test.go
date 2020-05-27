/*
Copyright 2019 The Knative Authors

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

package webhook

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"knative.dev/pkg/controller"

	// Make system.Namespace() work in tests.
	_ "knative.dev/pkg/system/testing"

	. "knative.dev/pkg/reconciler/testing"
)

func newDefaultOptions() Options {
	return Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	}
}

const (
	testResourceName = "test-resource"
	user1            = "brutto@knative.dev"
)

func init() {
	// Don't hang forever when running tests.
	GracePeriod = 100 * time.Millisecond
}

func newNonRunningTestWebhook(t *testing.T, options Options, acs ...interface{}) (
	ctx context.Context, ac *Webhook, cancel context.CancelFunc) {
	t.Helper()

	// Create fake clients
	ctx, ctxCancel, informers := SetupFakeContextWithCancel(t)
	ctx = WithOptions(ctx, options)

	stopCb, err := controller.RunInformers(ctx.Done(), informers...)
	if err != nil {
		t.Fatalf("StartInformers() = %v", err)
	}
	cancel = func() {
		ctxCancel()
		stopCb()
	}

	ac, err = New(ctx, acs)
	if err != nil {
		t.Fatalf("Failed to create new admission controller: %v", err)
	}
	return
}

func TestRegistrationStopChanFire(t *testing.T) {
	opts := newDefaultOptions()
	_, ac, cancel := newNonRunningTestWebhook(t, opts)
	defer cancel()

	stopCh := make(chan struct{})

	var g errgroup.Group
	g.Go(func() error {
		return ac.Run(stopCh)
	})
	close(stopCh)

	if err := g.Wait(); err != nil {
		t.Fatal("Error during run: ", err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", opts.Port))
	if err == nil {
		conn.Close()
		t.Errorf("Unexpected success to dial to port %d", opts.Port)
	}
}

func TestWebhookKubeletProbe(t *testing.T) {
	opts := newDefaultOptions()
	ctx, webhook, cancel := newNonRunningTestWebhook(t, opts)
	defer cancel()

	recorder := bombRecorder{ResponseRecorder: httptest.NewRecorder()}
	probeReq := httptest.NewRequest("GET", "/", nil)
	probeReq.Header.Set("User-Agent", "kube-probe/1.16")

	webhook.ServeHTTP(&recorder, probeReq)

	if got, want := recorder.Code, http.StatusOK; got != want {
		t.Fatalf("Probe got HTTP status %d - expected %d", got, want)
	}

	if got, want := recorder.writeCount, 1; got != want {
		t.Errorf("HTTP status was written %d times - expected only one write", got)
	}

	// Stop the webhook - which means probes should fail
	//
	// The steps below aren't obvious and requires you to
	// know the implementation details
	cancel()
	webhook.Run(ctx.Done())

	recorder = bombRecorder{ResponseRecorder: httptest.NewRecorder()}
	webhook.ServeHTTP(&recorder, probeReq)

	if got, want := recorder.Code, http.StatusInternalServerError; got != want {
		t.Fatalf("Probe got HTTP status %d - expected %d", got, want)
	}

	if got, want := recorder.writeCount, 1; got != want {
		t.Errorf("HTTP status was written %d times - expected only one write", got)
	}
}

type bombRecorder struct {
	*httptest.ResponseRecorder
	writeCount int
}

func (rw *bombRecorder) WriteHeader(code int) {
	rw.writeCount += 1
	rw.ResponseRecorder.WriteHeader(code)
}
