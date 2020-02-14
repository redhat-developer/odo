/*
Copyright 2019 The Tekton Authors

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

package gitlab

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/tektoncd/pipeline/pkg/logging"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func TestInterceptor_ExecuteTrigger(t *testing.T) {
	type args struct {
		payload   []byte
		secret    *corev1.Secret
		token     string
		eventType string
	}
	tests := []struct {
		name    string
		GitLab  *triggersv1.GitLabInterceptor
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "no secret",
			GitLab: &triggersv1.GitLabInterceptor{},
			args: args{
				payload: []byte("somepayload"),
				token:   "foo",
			},
			want:    []byte("somepayload"),
			wantErr: false,
		},
		{
			name: "invalid header for secret",
			GitLab: &triggersv1.GitLabInterceptor{
				SecretRef: &triggersv1.SecretRef{
					SecretName: "mysecret",
					SecretKey:  "token",
				},
			},
			args: args{
				token: "foo",
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
					Data: map[string][]byte{
						"token": []byte("secrettoken"),
					},
				},
				payload: []byte("somepayload"),
			},
			wantErr: true,
		},
		{
			name: "valid header for secret",
			GitLab: &triggersv1.GitLabInterceptor{
				SecretRef: &triggersv1.SecretRef{
					SecretName: "mysecret",
					SecretKey:  "token",
				},
			},
			args: args{
				token: "secret",
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
					Data: map[string][]byte{
						"token": []byte("secret"),
					},
				},
				payload: []byte("somepayload"),
			},
			wantErr: false,
			want:    []byte("somepayload"),
		},
		{
			name: "valid event",
			GitLab: &triggersv1.GitLabInterceptor{
				EventTypes: []string{"foo", "bar"},
			},
			args: args{
				eventType: "foo",
				payload:   []byte("somepayload"),
			},
			wantErr: false,
			want:    []byte("somepayload"),
		},
		{
			name: "invalid event",
			GitLab: &triggersv1.GitLabInterceptor{
				EventTypes: []string{"foo", "bar"},
			},
			args: args{
				eventType: "baz",
				payload:   []byte("somepayload"),
			},
			wantErr: true,
		},
		{
			name: "valid event, invalid secret",
			GitLab: &triggersv1.GitLabInterceptor{
				EventTypes: []string{"foo", "bar"},
				SecretRef: &triggersv1.SecretRef{
					SecretName: "mysecret",
					SecretKey:  "token",
				},
			},
			args: args{
				eventType: "bar",
				payload:   []byte("somepayload"),
				token:     "foo",
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
					Data: map[string][]byte{
						"token": []byte("secrettoken"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid event, valid secret",
			GitLab: &triggersv1.GitLabInterceptor{
				EventTypes: []string{"foo", "bar"},
				SecretRef: &triggersv1.SecretRef{
					SecretName: "mysecret",
					SecretKey:  "token",
				},
			},
			args: args{
				eventType: "baz",
				payload:   []byte("somepayload"),
				token:     "secrettoken",
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
					Data: map[string][]byte{
						"token": []byte("secrettoken"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid event, valid secret",
			GitLab: &triggersv1.GitLabInterceptor{
				EventTypes: []string{"foo", "bar"},
				SecretRef: &triggersv1.SecretRef{
					SecretName: "mysecret",
					SecretKey:  "token",
				},
			},
			args: args{
				eventType: "bar",
				payload:   []byte("somepayload"),
				token:     "secrettoken",
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
					Data: map[string][]byte{
						"token": []byte("secrettoken"),
					},
				},
			},
			want: []byte("somepayload"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := rtesting.SetupFakeContext(t)
			logger, _ := logging.NewLogger("", "")
			kubeClient := fakekubeclient.Get(ctx)
			request := &http.Request{
				Body: ioutil.NopCloser(bytes.NewReader(tt.args.payload)),
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			}
			if tt.args.token != "" {
				request.Header.Add("X-GitLab-Token", tt.args.token)
			}
			if tt.args.eventType != "" {
				request.Header.Add("X-GitLab-Event", tt.args.eventType)
			}
			if tt.args.secret != nil {
				ns := tt.GitLab.SecretRef.Namespace
				if ns == "" {
					ns = metav1.NamespaceDefault
				}
				if _, err := kubeClient.CoreV1().Secrets(ns).Create(tt.args.secret); err != nil {
					t.Error(err)
				}
			}
			w := &Interceptor{
				KubeClientSet: kubeClient,
				GitLab:        tt.GitLab,
				Logger:        logger,
			}
			resp, err := w.ExecuteTrigger(request)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Interceptor.ExecuteTrigger() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			got, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("error reading response: %v", err)
			}
			defer resp.Body.Close()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Interceptor.ExecuteTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
