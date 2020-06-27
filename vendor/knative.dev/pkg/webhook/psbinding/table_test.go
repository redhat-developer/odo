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

package psbinding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/client/injection/ducks/duck/v1/podspecable"
	kubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	mwhinformer "knative.dev/pkg/client/injection/kube/informers/admissionregistration/v1/mutatingwebhookconfiguration"
	_ "knative.dev/pkg/client/injection/kube/informers/admissionregistration/v1/mutatingwebhookconfiguration/fake"
	dynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
	secretinformer "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret"
	_ "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret/fake"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/ptr"
	"knative.dev/pkg/system"
	"knative.dev/pkg/webhook"
	certresources "knative.dev/pkg/webhook/certificates/resources"

	. "knative.dev/pkg/reconciler/testing"
	. "knative.dev/pkg/testing/duck"
	. "knative.dev/pkg/webhook/testing"
)

func checkDeploymentIsPatched(t *testing.T, r *TableRow) {
	t.Helper()
	ac := r.Reconciler.(webhook.AdmissionController)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "on-it",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "busybox",
					}},
				},
			},
		},
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Unable to serialize deployment: %v", err)
	}

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		Namespace: d.Namespace,
		Object:    runtime.RawExtension{Raw: b},
	}

	// It is allowed, and patched to include the environment variable.
	resp := ac.Admit(r.Ctx, req)
	ExpectAllowed(t, resp)
	ExpectPatches(t, resp.Patch, []jsonpatch.JsonPatchOperation{{
		Operation: "add",
		Path:      "/spec/template/spec/containers/0/env",
		Value: []interface{}{map[string]interface{}{
			"name":  "FOO",
			"value": "the-value",
		}},
	}})
}

func checkDeploymentIsPatchedBack(t *testing.T, r *TableRow) {
	t.Helper()
	ac := r.Reconciler.(webhook.AdmissionController)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "on-it",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "busybox",
						Env: []corev1.EnvVar{{
							Name:  "FOO",
							Value: "the-value",
						}},
					}},
				},
			},
		},
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Unable to serialize deployment: %v", err)
	}

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		Namespace: d.Namespace,
		Object:    runtime.RawExtension{Raw: b},
	}

	// It is allowed, and patched to REMOVE the environment variable.
	resp := ac.Admit(r.Ctx, req)
	ExpectAllowed(t, resp)
	ExpectPatches(t, resp.Patch, []jsonpatch.JsonPatchOperation{{
		Operation: "remove",
		Path:      "/spec/template/spec/containers/0/env",
	}})
}

func checkDeploymentIsNotPatched(t *testing.T, r *TableRow) {
	t.Helper()
	ac := r.Reconciler.(webhook.AdmissionController)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "off-it",
			Labels: map[string]string{
				"foo": "baz",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "busybox",
					}},
				},
			},
		},
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Unable to serialize deployment: %v", err)
	}

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		Namespace: d.Namespace,
		Object:    runtime.RawExtension{Raw: b},
	}

	// It is allowed, but not patched.
	resp := ac.Admit(r.Ctx, req)
	ExpectAllowed(t, resp)
	if want, got := "", string(resp.Patch); want != got {
		t.Errorf("Admit() = %s, got %s", got, want)
	}
}

func checkDeleteIgnored(t *testing.T, r *TableRow) {
	t.Helper()
	ac := r.Reconciler.(webhook.AdmissionController)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "on-it",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "busybox",
					}},
				},
			},
		},
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Unable to serialize deployment: %v", err)
	}

	req := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Delete,
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		Namespace: d.Namespace,
		Object:    runtime.RawExtension{Raw: b},
	}

	// It is allowed, and patched to include the environment variable.
	resp := ac.Admit(r.Ctx, req)
	ExpectAllowed(t, resp)
	if want, got := "", string(resp.Patch); want != got {
		t.Errorf("Admit() = %s, got %s", got, want)
	}
}

func TestWebhookReconcile(t *testing.T) {
	name, path := "foo.bar.baz", "/blah"
	secretName := "webhook-secret"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: system.Namespace(),
		},
		Data: map[string][]byte{
			certresources.ServerKey:  []byte("present"),
			certresources.ServerCert: []byte("present"),
			certresources.CACert:     []byte("present"),
		},
	}

	equivalent := admissionregistrationv1.Equivalent

	// The key to use, which for this singleton reconciler doesn't matter (although the
	// namespace matters for namespace validation).
	key := system.Namespace() + "/does not matter"

	table := TableTest{{
		Name:    "no secret",
		Key:     key,
		WantErr: true,
	}, {
		Name: "secret missing CA Cert",
		Key:  key,
		Objects: []runtime.Object{&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: system.Namespace(),
			},
			Data: map[string][]byte{
				certresources.ServerKey:  []byte("present"),
				certresources.ServerCert: []byte("present"),
				// certresources.CACert:     []byte("missing"),
			},
		}},
		WantErr: true,
	}, {
		Name:    "secret exists, but MWH does not",
		Key:     key,
		Objects: []runtime.Object{secret},
		WantErr: true,
	}, {
		Name: "secret and MWH exist, missing service reference",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
				}},
			},
		},
		WantErr: true,
	}, {
		Name: "secret and MWH exist, missing other stuff",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
						},
					},
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is added.
							Path: ptr.String(path),
						},
						// CABundle is added.
						CABundle: []byte("present"),
					},
					// Rules are added.
					Rules: nil,
					// MatchPolicy is added.
					MatchPolicy: &equivalent,
					// Selectors are added.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		}},
	}, {
		Name: "secret and MWH exist, added fields are incorrect",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Incorrect
							Path: ptr.String("incorrect"),
						},
						// Incorrect
						CABundle: []byte("incorrect"),
					},
					// Incorrect (really just incomplete)
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"pkg.knative.dev"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"innerdefaultresources/*"},
						},
					}},
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fixed.
							Path: ptr.String(path),
						},
						// CABundle is fixed.
						CABundle: []byte("present"),
					},
					// Rules are fixed.
					Rules: nil,
					// MatchPolicy is added.
					MatchPolicy: &equivalent,
					// Selectors are added.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		}},
	}, {
		Name:    "failure updating MWH",
		Key:     key,
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("update", "mutatingwebhookconfigurations"),
		},
		Objects: []runtime.Object{secret,
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Incorrect
							Path: ptr.String("incorrect"),
						},
						// Incorrect
						CABundle: []byte("incorrect"),
					},
					// Incorrect (really just incomplete)
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"pkg.knative.dev"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"innerdefaultresources/*"},
						},
					}},
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fixed.
							Path: ptr.String(path),
						},
						// CABundle is fixed.
						CABundle: []byte("present"),
					},
					// Rules are fixed.
					Rules: nil,
					// MatchPolicy is added.
					MatchPolicy: &equivalent,
					// Selectors are added.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		}},
	}, {
		Name: ":fire: everything is fine :fire:",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fine.
							Path: ptr.String(path),
						},
						// CABundle is fine.
						CABundle: []byte("present"),
					},
					// Rules are fine.
					Rules: nil,
					// MatchPolicy is fine.
					MatchPolicy: &equivalent,
					// Selectors are fine.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
	}, {
		Name: "a new binding has entered the match",
		Key:  key,
		Objects: []runtime.Object{secret,
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "random.knative.dev/v2beta3",
							Kind:       "Knoodle",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
				},
			},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fine.
							Path: ptr.String(path),
						},
						// CABundle is fine.
						CABundle: []byte("present"),
					},
					// Rules are fine.
					Rules: nil,
					// MatchPolicy is fine.
					MatchPolicy: &equivalent,
					// Selectors are fine.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							Path:      ptr.String(path),
						},
						CABundle: []byte("present"),
					},
					// A new rule is added to intercept the new type.
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"random.knative.dev"},
							APIVersions: []string{"v2beta3"},
							Resources:   []string{"knoodles/*"},
						},
					}},
					MatchPolicy:       &equivalent,
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		}},
	}, {
		Name: "steady state direct bindings",
		Key:  key,
		Objects: []runtime.Object{secret,
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar1",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "random.knative.dev/v2beta3",
							Kind:       "Knoodle",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "one-value",
				},
			},
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar2",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "the-value",
				},
			},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							Path:      ptr.String(path),
						},
						CABundle: []byte("present"),
					},
					// A new rule is added to intercept the new type.
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"apps"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments/*"},
						},
					}, {
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"random.knative.dev"},
							APIVersions: []string{"v2beta3"},
							Resources:   []string{"knoodles/*"},
						},
					}},
					MatchPolicy:       &equivalent,
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
		// Verify that Admit properly patches deployments after being programmed
		// with the binding.
		PostConditions: []func(*testing.T, *TableRow){
			checkDeploymentIsPatched,
			checkDeploymentIsNotPatched,
			checkDeleteIgnored,
		},
	}, {
		Name: "steady state selector",
		Key:  key,
		Objects: []runtime.Object{secret,
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar1",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "random.knative.dev/v2beta3",
							Kind:       "Knoodle",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								// Match everything.
								MatchLabels: map[string]string{},
							},
						},
					},
					Foo: "one-value",
				},
			},
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar2",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"foo": "bar",
								},
							},
						},
					},
					Foo: "the-value",
				},
			},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							Path:      ptr.String(path),
						},
						CABundle: []byte("present"),
					},
					// A new rule is added to intercept the new type.
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"apps"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments/*"},
						},
					}, {
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"random.knative.dev"},
							APIVersions: []string{"v2beta3"},
							Resources:   []string{"knoodles/*"},
						},
					}},
					MatchPolicy:       &equivalent,
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
		// Verify that Admit properly patches deployments after being programmed
		// with the binding.
		PostConditions: []func(*testing.T, *TableRow){
			checkDeploymentIsPatched,
			checkDeploymentIsNotPatched,
			checkDeleteIgnored,
		},
	}, {
		Name: "tombstoned binding undoes patch",
		Key:  key,
		Objects: []runtime.Object{secret,
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar2",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"foo": "bar",
								},
							},
						},
					},
					Foo: "the-value",
				},
			},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							Path:      ptr.String(path),
						},
						CABundle: []byte("present"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"apps"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments/*"},
						},
					}},
					MatchPolicy:       &equivalent,
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
		// Verify that Admit properly patches deployments after being programmed
		// with the binding.
		PostConditions: []func(*testing.T, *TableRow){
			checkDeploymentIsPatchedBack,
			checkDeploymentIsNotPatched,
			checkDeleteIgnored,
		},
	}, {
		Name: "multiple new bindings have entered the match",
		Key:  key,
		Objects: []runtime.Object{secret,
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "random.knative.dev/v2beta3",
							Kind:       "Knoodle",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
				},
			},
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "blah",
					Name:      "bazinga",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "pseudorandom.knative.dev/v3beta1",
							Kind:       "Knoogle",
							Namespace:  "blah",
							Name:       "oh-yeah",
						},
					},
				},
			},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fine.
							Path: ptr.String(path),
						},
						// CABundle is fine.
						CABundle: []byte("present"),
					},
					// Rules are fine.
					Rules: nil,
					// MatchPolicy is fine.
					MatchPolicy: &equivalent,
					// Selectors are fine.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							Path:      ptr.String(path),
						},
						CABundle: []byte("present"),
					},
					// New rules are added to intercept the new types.
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"pseudorandom.knative.dev"},
							APIVersions: []string{"v3beta1"},
							Resources:   []string{"knoogles/*"},
						},
					}, {
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"random.knative.dev"},
							APIVersions: []string{"v2beta3"},
							Resources:   []string{"knoodles/*"},
						},
					}},
					MatchPolicy:       &equivalent,
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		}},
	}, {
		Name: "a new selector-based binding has entered the match",
		Key:  key,
		Objects: []runtime.Object{secret,
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "random.knative.dev/v2beta3",
							Kind:       "Knoodle",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
			&admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fine.
							Path: ptr.String(path),
						},
						// CABundle is fine.
						CABundle: []byte("present"),
					},
					// Rules are fine.
					Rules: nil,
					// MatchPolicy is fine.
					MatchPolicy: &equivalent,
					// Selectors are fine.
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							Path:      ptr.String(path),
						},
						CABundle: []byte("present"),
					},
					// A new rule is added to intercept the new type.
					Rules: []admissionregistrationv1.RuleWithOperations{{
						Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"random.knative.dev"},
							APIVersions: []string{"v2beta3"},
							Resources:   []string{"knoodles/*"},
						},
					}},
					MatchPolicy:       &equivalent,
					NamespaceSelector: &ExclusionSelector,
					ObjectSelector:    &ExclusionSelector,
				}},
			},
		}},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := NewReconciler(name, path, secretName, kubeclient.Get(ctx), listers.GetMutatingWebhookConfigurationLister(), listers.GetSecretLister(), nil)
		r.ListAll = func() ([]Bindable, error) {
			bl := make([]Bindable, 0)
			for _, elt := range listers.GetDuckObjects() {
				b, ok := elt.(Bindable)
				if !ok {
					continue
				}
				bl = append(bl, b)
			}
			return bl, nil
		}
		return r
	}))
}

func TestNew(t *testing.T) {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{})

	c := NewAdmissionController(ctx, "foo", "/bar",
		func(context.Context, cache.ResourceEventHandler) ListAll {
			return func() ([]Bindable, error) {
				return nil, nil
			}
		},
		func(ctx context.Context, b Bindable) (context.Context, error) {
			return ctx, nil
		})
	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}

	if want, got := 0, c.WorkQueue.Len(); want != got {
		t.Errorf("WorkQueue.Len() = %d, wanted %d", got, want)
	}

	la, ok := c.Reconciler.(pkgreconciler.LeaderAware)
	if !ok {
		t.Fatalf("%T is not leader aware", c.Reconciler)
	}

	if err := la.Promote(pkgreconciler.UniversalBucket(), c.MaybeEnqueueBucketKey); err != nil {
		t.Errorf("Promote() = %v", err)
	}

	if want, got := 1, c.WorkQueue.Len(); want != got {
		t.Errorf("WorkQueue.Len() = %d, wanted %d", got, want)
	}
}

func TestNewReconciler(t *testing.T) {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{})
	tests := []struct {
		name           string
		selectorOption ReconcilerOption
		wantSelector   metav1.LabelSelector
	}{
		{
			name:         "no selector, use default",
			wantSelector: ExclusionSelector,
		},
		{
			name:           "ExclusionSelector option",
			selectorOption: WithSelector(ExclusionSelector),
			wantSelector:   ExclusionSelector,
		},
		{
			name:           "InclusionSelector option",
			selectorOption: WithSelector(InclusionSelector),
			wantSelector:   InclusionSelector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := kubeclient.Get(ctx)
			mwhInformer := mwhinformer.Get(ctx)
			secretInformer := secretinformer.Get(ctx)
			withContext := func(ctx context.Context, b Bindable) (context.Context, error) {
				return ctx, nil
			}
			var r *Reconciler
			if tt.selectorOption == nil {
				r = NewReconciler("foo", "/bar", "foosec", client, mwhInformer.Lister(), secretInformer.Lister(), withContext)
			} else {
				r = NewReconciler("foo", "/bar", "foosec", client, mwhInformer.Lister(), secretInformer.Lister(), withContext, tt.selectorOption)
			}
			if diff := cmp.Diff(r.selector, tt.wantSelector); diff != "" {
				t.Errorf("Wrong selector configured. Got: %+v, want: %+v, diff: %v", r.selector, tt.wantSelector, diff)
			}
		})
	}
}

func TestBaseReconcile(t *testing.T) {
	table := TableTest{{
		Name: "bad key",
		Key:  "this/is/a/bad/key",
	}, {
		Name: "not found",
		Key:  "its/missing",
	}, {
		Name: "add finalizer, add env var",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: mustTU(t, &TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			}),
		}},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddFinalizer("foo", "bar", "" /* resource version */),
			patchAddLabel("foo"),
			patchAddEnv("foo", "on-it", "asdfasdfasdfasdf"),
		},
	}, {
		Name:    "failure adding finalizer",
		Key:     "foo/bar",
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("patch", "testbindables"),
		},
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddFinalizer("foo", "bar", "" /* resource version */),
		},
	}, {
		Name:    "failure patching deployment",
		Key:     "foo/bar",
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("patch", "deployments"),
		},
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "foo",
					Name:       "bar",
					Finalizers: []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "foo",
					Labels: map[string]string{duck.BindingIncludeLabel: "true"},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
							}},
						},
					},
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddEnv("foo", "on-it", "asdfasdfasdfasdf"),
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: mustTU(t, &TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "foo",
					Name:       "bar",
					Finalizers: []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:    "Ready",
							Status:  "False",
							Reason:  "BindingFailed",
							Message: "failed binding subject on-it: inducing failure for patch deployments",
						}},
					},
				},
			}),
		}},
	}, {
		Name: "steady state",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "foo",
					Name:       "bar",
					Finalizers: []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
								Env: []corev1.EnvVar{{
									Name:  "FOO",
									Value: "asdfasdfasdfasdf",
								}},
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
		},
	}, {
		Name: "finalizing, but not our turn.",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers: []string{
						"slow.your.role",
						"testbindables.duck.knative.dev",
					},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "new value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
		},
	}, {
		Name: "finalizing, missing subject (remove the finalizer).",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers: []string{
						"testbindables.duck.knative.dev",
					},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "new value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:    "Ready",
							Status:  "False",
							Reason:  "SubjectMissing",
							Message: `deployments.apps "on-it" not found`,
						}},
					},
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchRemoveFinalizer("foo", "bar", "" /* resource version */),
		},
	}, {
		Name: "finalizing forbidden subject",
		Key:  "foo/bar",
		WithReactors: []clientgotesting.ReactionFunc{
			// This will cause the duck informer factory to return a Forbidden error on Get(gvr)
			// The informer calls list to ensure the type exists - this will
			func(a clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				if a.Matches("list", "deployments") {
					return true, nil, apierrs.NewForbidden(schema.GroupResource{}, "", errors.New("some-error"))
				}
				return false, nil, nil
			},
		},
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "new value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchRemoveFinalizer("foo", "bar", "" /* resource version */),
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: mustTU(t, &TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "new value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "False",
							Reason: "SubjectUnavailable",
							// prefix comes from apiserrs.NewForbidden
							Message: "forbidden: some-error",
						}},
					},
				},
			}),
		}},
	}, {
		Name: "finalizing (unbind, and remove the finalizer)",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers: []string{
						"testbindables.duck.knative.dev",
					},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
								Env: []corev1.EnvVar{{
									Name:  "FOO",
									Value: "value",
								}},
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
			patchRemoveEnv("foo", "on-it"),
			patchRemoveFinalizer("foo", "bar", "" /* resource version */),
		},
	}, {
		Name: "add finalizer, add env var (via selector)",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddFinalizer("foo", "bar", "" /* resource version */),
			patchAddLabel("foo"),
			patchAddEnv("foo", "on-it", "asdfasdfasdfasdf"),
		},
	}, {
		Name: "steady state (via selector)",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "foo",
					Name:       "bar",
					Finalizers: []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
								Env: []corev1.EnvVar{{
									Name:  "FOO",
									Value: "asdfasdfasdfasdf",
								}},
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
		},
	}, {
		Name: "finalizing, missing subject (remove the finalizer, via selector)",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers: []string{
						"testbindables.duck.knative.dev",
					},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
					Foo: "new value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
			patchRemoveFinalizer("foo", "bar", "" /* resource version */),
		},
	}, {
		Name: "finalizing (unbind, and remove the finalizer, via selector)",
		Key:  "foo/bar",
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers: []string{
						"testbindables.duck.knative.dev",
					},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
					Foo: "value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
								Env: []corev1.EnvVar{{
									Name:  "FOO",
									Value: "value",
								}},
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
			patchRemoveEnv("foo", "on-it"),
			patchRemoveFinalizer("foo", "bar", "" /* resource version */),
		},
	}, {
		Name:    "failure updating status",
		Key:     "foo/bar",
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("update", "testbindables"),
		},
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "foo",
					Name:       "bar",
					Finalizers: []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
								Env: []corev1.EnvVar{{
									Name:  "FOO",
									Value: "asdfasdfasdfasdf",
								}},
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: mustTU(t, &TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "foo",
					Name:       "bar",
					Finalizers: []string{"testbindables.duck.knative.dev"},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Name:       "on-it",
						},
					},
					Foo: "asdfasdfasdfasdf",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:   "Ready",
							Status: "True",
						}},
					},
				},
			}),
		}},
	}, {
		Name:    "finalizing (error during unbind)",
		Key:     "foo/bar",
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("patch", "deployments"),
		},
		Objects: []runtime.Object{
			&TestBindable{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "foo",
					Name:              "bar",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers: []string{
						"testbindables.duck.knative.dev",
					},
				},
				Spec: TestBindableSpec{
					BindingSpec: duckv1alpha1.BindingSpec{
						Subject: tracker.Reference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Namespace:  "foo",
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
					Foo: "value",
				},
				Status: TestBindableStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{{
							Type:    "Ready",
							Status:  "False",
							Reason:  "BindingFailed",
							Message: "failed binding subject on-it: inducing failure for patch deployments",
						}},
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "on-it",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "foo",
								Image: "busybox",
								Env: []corev1.EnvVar{{
									Name:  "FOO",
									Value: "value",
								}},
							}},
						},
					},
				},
			},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		WantPatches: []clientgotesting.PatchActionImpl{
			patchAddLabel("foo"),
			patchRemoveEnv("foo", "on-it"),
		},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		gvr := SchemeGroupVersion.WithResource("testbindables")
		ctx = podspecable.WithDuck(ctx)

		dc := dynamicclient.Get(ctx)

		return &BaseReconciler{
			GVR: gvr,

			DynamicClient: dc,
			Factory:       podspecable.Get(ctx),

			Tracker: &FakeTracker{},

			Recorder: record.NewFakeRecorder(20),

			Get: func(namespace, name string) (Bindable, error) {
				for _, elt := range listers.GetDuckObjects() {
					b, ok := elt.(*TestBindable)
					if !ok {
						continue
					}
					if b.Namespace != namespace || b.Name != name {
						continue
					}
					return b, nil
				}
				return nil, apierrs.NewNotFound(gvr.GroupResource(), name)
			},
			NamespaceLister: listers.GetNamespaceLister(),
		}
	}))
}

func mustTU(t *testing.T, ro duck.OneOfOurs) *unstructured.Unstructured {
	u, err := duck.ToUnstructured(ro)
	if err != nil {
		t.Fatalf("ToUnstructured(%+v) = %v", ro, err)
	}
	return u
}

func patchAddLabel(namespace string) clientgotesting.PatchActionImpl {
	action := clientgotesting.PatchActionImpl{}
	resource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}
	actionImpl := clientgotesting.ActionImpl{
		Namespace:   namespace,
		Verb:        "patch",
		Resource:    resource,
		Subresource: "",
	}
	action.ActionImpl = actionImpl
	action.Name = namespace
	action.PatchType = types.MergePatchType

	patch, _ := json.Marshal(jsonLabelPatch)
	action.Patch = patch
	return action
}
func patchAddFinalizer(namespace, name, resourceVersion string) clientgotesting.PatchActionImpl {
	action := clientgotesting.PatchActionImpl{}
	action.Name = name
	action.Namespace = namespace

	patch := fmt.Sprintf(`{"metadata":{"finalizers":["testbindables.duck.knative.dev"],"resourceVersion":%q}}`, resourceVersion)

	action.Patch = []byte(patch)
	return action
}

func patchRemoveFinalizer(namespace, name, resourceVersion string) clientgotesting.PatchActionImpl {
	action := clientgotesting.PatchActionImpl{}
	action.Name = name
	action.Namespace = namespace

	patch := fmt.Sprintf(`{"metadata":{"finalizers":[],"resourceVersion":%q}}`, resourceVersion)

	action.Patch = []byte(patch)
	return action
}

func patchAddEnv(namespace, name, value string) clientgotesting.PatchActionImpl {
	action := clientgotesting.PatchActionImpl{}
	action.Name = name
	action.Namespace = namespace

	patch := fmt.Sprintf(`[{"op":"add","path":"/spec/template/spec/containers/0/env","value":[{"name":"FOO","value":%q}]}]`, value)

	action.Patch = []byte(patch)
	return action
}

func patchRemoveEnv(namespace, name string) clientgotesting.PatchActionImpl {
	action := clientgotesting.PatchActionImpl{}
	action.Name = name
	action.Namespace = namespace

	patch := `[{"op":"remove","path":"/spec/template/spec/containers/0/env"}]`

	action.Patch = []byte(patch)
	return action
}
