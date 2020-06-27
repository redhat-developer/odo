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

package configmaps

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	// Injection stuff
	_ "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/admissionregistration/v1/validatingwebhookconfiguration/fake"
	_ "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret/fake"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/system"
	"knative.dev/pkg/webhook"

	_ "knative.dev/pkg/system/testing"

	. "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/reconciler/testing"
	. "knative.dev/pkg/webhook/testing"
)

const (
	testConfigValidationName = "configmap.webhook.knative.dev"
	testConfigValidationPath = "/cm"
	testConfigName           = "test-config"
)

var (
	validations = configmap.Constructors{
		testConfigName: newConfigFromConfigMap,
	}
	initialConfigWebhook = &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: testConfigValidationName,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: testConfigValidationName,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Namespace: system.Namespace(),
					Name:      "webhook",
				},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "pkg.knative.dev/release",
					Operator: metav1.LabelSelectorOpExists,
				}},
			},
		}},
	}
)

func newNonRunningTestConfigValidationController(t *testing.T) (
	kubeClient *fakekubeclientset.Clientset,
	ac *reconciler) {
	t.Helper()
	// Create fake clients
	kubeClient = fakekubeclientset.NewSimpleClientset(initialConfigWebhook)

	ac = newTestConfigValidationController(t)
	return
}

func newTestConfigValidationController(t *testing.T) *reconciler {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{
		SecretName: "webhook-secret",
	})
	return NewAdmissionController(ctx, testConfigValidationName, testConfigValidationPath,
		validations).Reconciler.(*reconciler)
}

func TestDeleteAllowedForConfigMap(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Delete,
	}

	if resp := ac.Admit(TestContextWithLogger(t), req); !resp.Allowed {
		t.Fatal("Unexpected denial of delete")
	}
}

func TestConnectAllowedForConfigMap(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Connect,
	}

	resp := ac.Admit(TestContextWithLogger(t), req)
	if !resp.Allowed {
		t.Fatalf("Unexpected denial of connect")
	}
}

func TestNonConfigMapKindFails(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "pkg.knative.dev",
			Version: "v1alpha1",
			Kind:    "Garbage",
		},
	}

	ExpectFailsWith(t, ac.Admit(TestContextWithLogger(t), req), "unhandled kind")
}

func TestAdmitCreateValidConfigMap(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createValidConfigMap()
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectAllowed(t, resp)
}

func TestDenyInvalidCreateConfigMapWithWrongType(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createWrongTypeConfigMap()
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectFailsWith(t, resp, "invalid syntax")
}

func TestDenyInvalidCreateConfigMapOutOfRange(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createWrongValueConfigMap()
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectFailsWith(t, resp, "out of range")
}

func TestAdmitUpdateValidConfigMap(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createValidConfigMap()
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, updateCreateConfigMapRequest(ctx, t, r))

	ExpectAllowed(t, resp)
}

func TestDenyInvalidUpdateConfigMapWithWrongType(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createWrongTypeConfigMap()
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectFailsWith(t, resp, "invalid syntax")
}

func TestDenyInvalidUpdateConfigMapOutOfRange(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createWrongValueConfigMap()
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectFailsWith(t, resp, "out of range")
}

func TestAllowConfigMapExample(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := createValidConfigMap()
	// Add an _example field but add no annotation.
	r.Data[configmap.ExampleKey] = "bar"
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectAllowed(t, resp)
}

func TestAllowUnknownConfigMapExample(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-other-config",
			Annotations: map[string]string{
				configmap.ExampleChecksumAnnotation: "foo",
			},
		},
		Data: map[string]string{
			configmap.ExampleKey: "bar",
		},
	}
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectAllowed(t, resp)
}

func TestDenyInvalidUpdateConfigMapExample(t *testing.T) {
	_, ac := newNonRunningTestConfigValidationController(t)

	r := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: testConfigName,
			Annotations: map[string]string{
				configmap.ExampleChecksumAnnotation: "foo",
			},
		},
		Data: map[string]string{
			configmap.ExampleKey: "bar",
		},
	}
	ctx := TestContextWithLogger(t)

	resp := ac.Admit(ctx, createCreateConfigMapRequest(ctx, t, r))

	ExpectFailsWith(t, resp, fmt.Sprintf("%q modified", configmap.ExampleKey))
}

type config struct {
	value float64
}

func newConfigFromConfigMap(configMap *corev1.ConfigMap) (*config, error) {
	data := configMap.Data
	cfg := &config{}
	for _, b := range []struct {
		key   string
		field *float64
	}{{
		key:   "value",
		field: &cfg.value,
	}} {
		if raw, ok := data[b.key]; !ok {
			return nil, fmt.Errorf("not found")
		} else if val, err := strconv.ParseFloat(raw, 64); err != nil {
			return nil, err
		} else {
			*b.field = val
		}
	}

	// some sample validation on the value
	if cfg.value > 2.0 || cfg.value < 0.0 {
		return nil, fmt.Errorf("out of range")
	}

	return cfg, nil
}

func createValidConfigMap() *corev1.ConfigMap {
	return createConfigMap("1.5")
}

func createWrongTypeConfigMap() *corev1.ConfigMap {
	return createConfigMap("bad")
}

func createWrongValueConfigMap() *corev1.ConfigMap {
	return createConfigMap("2.5")
}

func createConfigMap(value string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: system.Namespace(),
			Name:      testConfigName,
		},
		Data: map[string]string{
			"value": value,
		},
	}
}

func createCreateConfigMapRequest(ctx context.Context, t *testing.T, r *corev1.ConfigMap) *admissionv1.AdmissionRequest {
	return configMapRequest(t, r, admissionv1.Create)
}

func updateCreateConfigMapRequest(ctx context.Context, t *testing.T, r *corev1.ConfigMap) *admissionv1.AdmissionRequest {
	return configMapRequest(t, r, admissionv1.Update)
}

func configMapRequest(
	t *testing.T,
	r *corev1.ConfigMap,
	o admissionv1.Operation,
) *admissionv1.AdmissionRequest {
	t.Helper()
	req := &admissionv1.AdmissionRequest{
		Operation: o,
		Kind: metav1.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		},
		UserInfo: authenticationv1.UserInfo{Username: "mattmoor"},
	}
	marshaled, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Failed to marshal resource: %v", err)
	}
	req.Object.Raw = marshaled
	req.Resource.Group = ""
	return req
}
