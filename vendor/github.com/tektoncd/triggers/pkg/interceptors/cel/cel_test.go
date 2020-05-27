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

package cel

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/tektoncd/pipeline/pkg/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	rtesting "knative.dev/pkg/reconciler/testing"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const testNS = "testing-ns"

func TestInterceptor_ExecuteTrigger(t *testing.T) {
	tests := []struct {
		name    string
		CEL     *triggersv1.CELInterceptor
		payload io.ReadCloser
		want    []byte
	}{
		{
			name: "simple body check with matching body",
			CEL: &triggersv1.CELInterceptor{
				Filter: "body.value == 'testing'",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"value":"testing"}`)),
			want:    []byte(`{"value":"testing"}`),
		},
		{
			name: "simple header check with matching header",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header['X-Test'][0] == 'test-value'",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{}`)),
			want:    []byte(`{}`),
		},
		{
			name: "overloaded header check with case insensitive matching",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header.match('x-test', 'test-value')",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{}`)),
			want:    []byte(`{}`),
		},
		{
			name: "body and header check",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header.match('x-test', 'test-value') && body.value == 'test'",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"value":"test"}`)),
			want:    []byte(`{"value":"test"}`),
		},
		{
			name: "body and header check",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header.canonical('x-test') == 'test-value' && body.value == 'test'",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"value":"test"}`)),
			want:    []byte(`{"value":"test"}`),
		},
		{
			name: "single overlay",
			CEL: &triggersv1.CELInterceptor{
				Filter: "body.value == 'test'",
				Overlays: []triggersv1.CELOverlay{
					{Key: "new", Expression: "body.value"},
				},
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"value":"test"}`)),
			want:    []byte(`{"new":"test","value":"test"}`),
		},
		{
			name: "single overlay with no filter",
			CEL: &triggersv1.CELInterceptor{
				Overlays: []triggersv1.CELOverlay{
					{Key: "new", Expression: "body.ref.split('/')[2]"},
				},
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"ref":"refs/head/master","name":"testing"}`)),
			want:    []byte(`{"new":"master","ref":"refs/head/master","name":"testing"}`),
		},
		{
			name: "overlay with string library functions",
			CEL: &triggersv1.CELInterceptor{
				Overlays: []triggersv1.CELOverlay{
					{Key: "new", Expression: "body.ref.split('/')[2]"},
					{Key: "replaced", Expression: "body.name.replace('ing','ed',0)"},
				},
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"ref":"refs/head/master","name":"testing"}`)),
			want:    []byte(`{"replaced":"testing","new":"master","ref":"refs/head/master","name":"testing"}`),
		},
		{
			name: "update with base64 decoding",
			CEL: &triggersv1.CELInterceptor{
				Overlays: []triggersv1.CELOverlay{
					{Key: "value", Expression: "body.value.decodeb64()"},
				},
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"value":"eyJ0ZXN0IjoiZGVjb2RlIn0="}`)),
			want:    []byte(`{"value":{"test":"decode"}}`),
		},
		{
			name: "multiple overlays",
			CEL: &triggersv1.CELInterceptor{
				Filter: "body.value == 'test'",
				Overlays: []triggersv1.CELOverlay{
					{Key: "test.one", Expression: "body.value"},
					{Key: "test.two", Expression: "body.value"},
				},
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"value":"test"}`)),
			want:    []byte(`{"test":{"two":"test","one":"test"},"value":"test"}`),
		},
		{
			name:    "nil body does not panic",
			CEL:     &triggersv1.CELInterceptor{Filter: "header.match('x-test', 'test-value')"},
			payload: nil,
			want:    []byte(`{}`),
		},
		{
			name: "incrementing an integer value",
			CEL: &triggersv1.CELInterceptor{
				Overlays: []triggersv1.CELOverlay{
					{Key: "val1", Expression: "body.count + 1.0"},
					{Key: "val2", Expression: "int(body.count) + 3"},
					{Key: "val3", Expression: "body.count + 3.5"},
					{Key: "val4", Expression: "body.measure * 3.0"},
				},
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"count":1,"measure":1.7}`)),
			want:    []byte(`{"val4":5.1,"val3":4.5,"val2":4,"val1":2,"count":1,"measure":1.7}`),
		},
		{
			name: "validating a secret",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header.canonical('X-Secret-Token').compareSecret('token', 'test-secret', 'testing-ns')",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"count":1,"measure":1.7}`)),
			want:    []byte(`{"count":1,"measure":1.7}`),
		},
		{
			name: "validating a secret in the default namespace",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header.canonical('X-Secret-Token').compareSecret('token', 'test-secret')",
			},
			payload: ioutil.NopCloser(bytes.NewBufferString(`{"count":1,"measure":1.7}`)),
			want:    []byte(`{"count":1,"measure":1.7}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(rt *testing.T) {
			logger, _ := logging.NewLogger("", "")
			ctx, _ := rtesting.SetupFakeContext(t)
			kubeClient := fakekubeclient.Get(ctx)
			if _, err := kubeClient.CoreV1().Secrets(testNS).Create(makeSecret()); err != nil {
				rt.Error(err)
			}
			w := NewInterceptor(tt.CEL, kubeClient, "testing-ns", logger)
			request := &http.Request{
				Body: tt.payload,
				Header: http.Header{
					"Content-Type":   []string{"application/json"},
					"X-Test":         []string{"test-value"},
					"X-Secret-Token": []string{"secrettoken"},
				},
			}
			resp, err := w.ExecuteTrigger(request)
			if err != nil {
				rt.Errorf("Interceptor.ExecuteTrigger() error = %v", err)
				return
			}
			got, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				rt.Fatalf("error reading response body: %v", err)
			}
			defer resp.Body.Close()
			if !reflect.DeepEqual(got, tt.want) {
				rt.Errorf("Interceptor.ExecuteTrigger() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestInterceptor_ExecuteTrigger_Errors(t *testing.T) {
	tests := []struct {
		name    string
		CEL     *triggersv1.CELInterceptor
		payload []byte
		want    string
	}{
		{
			name: "simple body check with non-matching body",
			CEL: &triggersv1.CELInterceptor{
				Filter: "body.value == 'test'",
			},
			payload: []byte(`{"value":"testing"}`),
			want:    "expression body.value == 'test' did not return true",
		},
		{
			name: "simple header check with non matching header",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header['X-Test'][0] == 'unknown'",
			},
			payload: []byte(`{}`),
			want:    "expression header.*'unknown' did not return true",
		},
		{
			name: "overloaded header check with case insensitive failed match",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header.match('x-test', 'no-match')",
			},
			payload: []byte(`{}`),
			want:    "expression header.match\\('x-test', 'no-match'\\) did not return true",
		},
		{
			name: "unable to parse the expression",
			CEL: &triggersv1.CELInterceptor{
				Filter: "header['X-Test",
			},
			payload: []byte(`{"value":"test"}`),
			want:    "Syntax error: token recognition error at: ''X-Test'",
		},
		{
			name: "unable to parse the JSON body",
			CEL: &triggersv1.CELInterceptor{
				Filter: "body.value == 'test'",
			},
			payload: []byte(`{]`),
			want:    "invalid character ']' looking for beginning of object key string",
		},
		{
			name: "bad overlay",
			CEL: &triggersv1.CELInterceptor{
				Filter: "body.value == 'test'",
				Overlays: []triggersv1.CELOverlay{
					{Key: "new", Expression: "test.value"},
				},
			},
			payload: []byte(`{"value":"test"}`),
			want:    "failed to evaluate overlay expression 'test.value'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := logging.NewLogger("", "")
			w := &Interceptor{
				CEL:    tt.CEL,
				Logger: logger,
			}
			request := &http.Request{
				Body: ioutil.NopCloser(bytes.NewReader(tt.payload)),
				GetBody: func() (io.ReadCloser, error) {
					return ioutil.NopCloser(bytes.NewReader(tt.payload)), nil
				},
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"X-Test":       []string{"test-value"},
				},
			}
			_, err := w.ExecuteTrigger(request)
			if !matchError(t, tt.want, err) {
				t.Errorf("evaluate() got %s, wanted %s", err, tt.want)
				return
			}
		})
	}
}

func TestExpressionEvaluation(t *testing.T) {
	testSHA := "ec26c3e57ca3a959ca5aad62de7213c562f8c821"
	testRef := "refs/heads/master"
	jsonMap := map[string]interface{}{
		"value": "testing",
		"sha":   testSHA,
		"ref":   testRef,
		"pull_request": map[string]interface{}{
			"commits": 2,
		},
		"b64value":  "ZXhhbXBsZQ==",
		"json_body": `{"testing": "value"}`,
	}
	refParts := strings.Split(testRef, "/")
	header := http.Header{}
	header.Add("X-Test-Header", "value")
	evalEnv := map[string]interface{}{"body": jsonMap, "header": header}
	tests := []struct {
		name   string
		expr   string
		secret *corev1.Secret
		want   ref.Val
	}{
		{
			name: "simple body value",
			expr: "body.value",
			want: types.String("testing"),
		},
		{
			name: "boolean body value",
			expr: "body.value == 'testing'",
			want: types.Bool(true),
		},
		{
			name: "truncate a long string",
			expr: "body.sha.truncate(7)",
			want: types.String("ec26c3e"),
		},
		{
			name: "truncate a string to its own length",
			expr: "body.value.truncate(7)",
			want: types.String("testing"),
		},
		{
			name: "truncate a string to fewer characters than it has",
			expr: "body.sha.truncate(45)",
			want: types.String(testSHA),
		},
		{
			name: "split a string on a character",
			expr: "body.ref.split('/')",
			want: types.NewStringList(types.NewRegistry(), refParts),
		},
		{
			name: "extract a branch from a non refs string",
			expr: "body.value.split('/')",
			want: types.NewStringList(types.NewRegistry(), []string{"testing"}),
		},
		{
			name: "combine split and truncate",
			expr: "body.value.split('/')[0].truncate(2)",
			want: types.String("te"),
		},
		{
			name: "exact header lookup",
			expr: "header.canonical('X-Test-Header')",
			want: types.String("value"),
		},
		{
			name: "canonical header lookup",
			expr: "header.canonical('x-test-header')",
			want: types.String("value"),
		},
		{
			name: "decode a base64 value",
			expr: "body.b64value.decodeb64()",
			want: types.Bytes("example"),
		},
		{
			name: "increment an integer",
			expr: "body.pull_request.commits + 1",
			want: types.Int(3),
		},
		{
			name:   "compare string against secret",
			expr:   "'secrettoken'.compareSecret('token', 'test-secret', 'testing-ns') ",
			want:   types.Bool(true),
			secret: makeSecret(),
		},
		{
			name:   "compare string against secret with no match",
			expr:   "'nomatch'.compareSecret('token', 'test-secret', 'testing-ns') ",
			want:   types.Bool(false),
			secret: makeSecret(),
		},
		{
			name:   "compare string against secret in the default namespace",
			expr:   "'secrettoken'.compareSecret('token', 'test-secret') ",
			want:   types.Bool(true),
			secret: makeSecret(),
		},
		{
			name: "parse JSON body in a string",
			expr: "body.json_body.parseJSON().testing == 'value'",
			want: types.Bool(true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(rt *testing.T) {
			ctx, _ := rtesting.SetupFakeContext(rt)
			kubeClient := fakekubeclient.Get(ctx)
			if tt.secret != nil {
				if _, err := kubeClient.CoreV1().Secrets(tt.secret.ObjectMeta.Namespace).Create(tt.secret); err != nil {
					rt.Error(err)
				}
			}
			env, err := makeCelEnv(testNS, kubeClient)
			if err != nil {
				t.Fatal(err)
			}

			got, err := evaluate(tt.expr, env, evalEnv)
			if err != nil {
				rt.Errorf("evaluate() got an error %s", err)
				return
			}
			_, ok := got.(*types.Err)
			if ok {
				rt.Errorf("error evaluating expression: %s", got)
				return
			}

			if !got.Equal(tt.want).(types.Bool) {
				rt.Errorf("evaluate() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluation_Error(t *testing.T) {
	testSHA := "ec26c3e57ca3a959ca5aad62de7213c562f8c821"
	jsonMap := map[string]interface{}{
		"value": "testing",
		"sha":   testSHA,
		"pull_request": map[string]interface{}{
			"commits": []string{},
		},
	}
	header := http.Header{}
	evalEnv := map[string]interface{}{"body": jsonMap, "header": header}
	tests := []struct {
		name     string
		expr     string
		secretNS string
		want     string
	}{
		{
			name: "unknown value",
			expr: "body.val",
			want: "no such key: val",
		},
		{
			name: "invalid syntax",
			expr: "body.value = 'testing'",
			want: "Syntax error: token recognition error",
		},
		{
			name: "unknown function",
			expr: "trunca(body.sha, 7)",
			want: "undeclared reference to 'trunca'",
		},
		{
			name: "invalid function overloading with match",
			expr: "body.match('testing', 'test')",
			want: "failed to convert to http.Header",
		},
		{
			name: "invalid function overloading with canonical",
			expr: "body.canonical('testing')",
			want: "failed to convert to http.Header",
		},
		{
			name: "invalid base64 decoding",
			expr: "\"AA=A\".decodeb64()",
			want: "failed to decode 'AA=A' in decodeB64.*illegal base64 data",
		},
		{
			name: "missing secret",
			expr: "'testing'.compareSecret('testing', 'testSecret', 'mytoken')",
			want: "failed to find secret.*testing.*",
		},
		{
			name:     "secret not in default ns",
			expr:     "'testing'.compareSecret('testSecret', 'mytoken')",
			secretNS: "another-ns",
			want:     "failed to find secret.*another-ns.*",
		},
		{
			name: "invalid parseJSON body",
			expr: "body.value.parseJSON().test == 'test'",
			want: "invalid character 'e' in literal",
		},
		{
			name: "base64 decoding non-string",
			expr: "body.pull_request.decodeb64()",
			want: "unexpected type 'map' passed to decodeB64",
		},
		{
			name: "parseJSON decoding non-string",
			expr: "body.pull_request.parseJSON().test == 'test'",
			want: "unexpected type 'map' passed to parseJSON",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(rt *testing.T) {
			ctx, _ := rtesting.SetupFakeContext(t)
			kubeClient := fakekubeclient.Get(ctx)
			ns := testNS
			if tt.secretNS != "" {
				secret := makeSecret()
				if _, err := kubeClient.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(secret); err != nil {
					rt.Error(err)
				}
				ns = tt.secretNS
			}
			env, err := makeCelEnv(ns, kubeClient)
			if err != nil {
				t.Fatal(err)
			}
			_, err = evaluate(tt.expr, env, evalEnv)
			if !matchError(t, tt.want, err) {
				rt.Errorf("evaluate() got %s, wanted %s", err, tt.want)
			}
		})
	}
}

func matchError(t *testing.T, s string, e error) bool {
	t.Helper()
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}

func makeSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      "test-secret",
		},
		Data: map[string][]byte{
			"token": []byte("secrettoken"),
		},
	}
}
