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
	"reflect"
	"testing"

	kubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/core/v1/secret/fake"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/ptr"
	"knative.dev/pkg/system"
	"knative.dev/pkg/webhook"
	certresources "knative.dev/pkg/webhook/certificates/resources"

	. "knative.dev/pkg/reconciler/testing"
	. "knative.dev/pkg/webhook/testing"
)

func TestReconcile(t *testing.T) {
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

	ruleScope := admissionregistrationv1beta1.NamespacedScope

	// These are the rules we expect given the context of "validations".
	expectedRules := []admissionregistrationv1beta1.RuleWithOperations{{
		Operations: []admissionregistrationv1beta1.OperationType{"CREATE", "UPDATE"},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{""},
			APIVersions: []string{"v1"},
			Resources:   []string{"configmaps/*"},
			Scope:       &ruleScope,
		},
	}}

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
		Name:    "secret exists, but VWH does not",
		Key:     key,
		Objects: []runtime.Object{secret},
		WantErr: true,
	}, {
		Name: "secret and VWH exist, missing service reference",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
				}},
			},
		},
		WantErr: true,
	}, {
		Name: "secret and VWH exist, missing other stuff",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
						},
					},
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is added.
							Path: ptr.String(path),
						},
						// CABundle is added.
						CABundle: []byte("present"),
					},
					// Rules are added.
					Rules: expectedRules,
				}},
			},
		}},
	}, {
		Name: "secret and VWH exist, added fields are incorrect",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Incorrect
							Path: ptr.String("incorrect"),
						},
						// Incorrect
						CABundle: []byte("incorrect"),
					},
					// Incorrect
					Rules: []admissionregistrationv1beta1.RuleWithOperations{{
						Operations: []admissionregistrationv1beta1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"pkg.knative.dev"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"innerdefaultresources/*"},
						},
					}},
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fixed.
							Path: ptr.String(path),
						},
						// CABundle is fixed.
						CABundle: []byte("present"),
					},
					// Rules are fixed.
					Rules: expectedRules,
				}},
			},
		}},
	}, {
		Name:    "failure updating VWH",
		Key:     key,
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("update", "validatingwebhookconfigurations"),
		},
		Objects: []runtime.Object{secret,
			&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Incorrect
							Path: ptr.String("incorrect"),
						},
						// Incorrect
						CABundle: []byte("incorrect"),
					},
					// Incorrect
					Rules: []admissionregistrationv1beta1.RuleWithOperations{{
						Operations: []admissionregistrationv1beta1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"pkg.knative.dev"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"innerdefaultresources/*"},
						},
					}},
				}},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fixed.
							Path: ptr.String(path),
						},
						// CABundle is fixed.
						CABundle: []byte("present"),
					},
					// Rules are fixed.
					Rules: expectedRules,
				}},
			},
		}},
	}, {
		Name: ":fire: everything is fine :fire:",
		Key:  key,
		Objects: []runtime.Object{secret,
			&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{{
					Name: name,
					ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
						Service: &admissionregistrationv1beta1.ServiceReference{
							Namespace: system.Namespace(),
							Name:      "webhook",
							// Path is fine.
							Path: ptr.String(path),
						},
						// CABundle is fine.
						CABundle: []byte("present"),
					},
					// Rules are fine.
					Rules: expectedRules,
				}},
			},
		},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		wh := &reconciler{
			name: name,
			path: path,

			client:       kubeclient.Get(ctx),
			vwhlister:    listers.GetValidatingWebhookConfigurationLister(),
			secretlister: listers.GetSecretLister(),

			constructors: make(map[string]reflect.Value),
			secretName:   secretName,
		}

		for configName, constructor := range validations {
			wh.registerConfig(configName, constructor)
		}

		return wh
	}))
}

func TestNew(t *testing.T) {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{})

	c := NewAdmissionController(ctx, "foo", "/bar", validations)
	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}
}
