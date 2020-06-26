/*
Copyright 2020 The Knative Authors

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

package conversion

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	apixclient "knative.dev/pkg/client/injection/apiextensions/client/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/ptr"
	"knative.dev/pkg/system"
	certresources "knative.dev/pkg/webhook/certificates/resources"

	. "knative.dev/pkg/reconciler/testing"
	. "knative.dev/pkg/webhook/testing"
)

func TestReconcile(t *testing.T) {
	key := "some.crd.group.dev"
	path := "/some/path"
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
		Name:    "secret exists, but CRD does not",
		Key:     key,
		Objects: []runtime.Object{secret},
		WantErr: true,
	}, {
		Name: "secret and CRD exist, missing service reference",
		Key:  key,
		Objects: []runtime.Object{
			secret,
			&apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{},
				},
			},
		},
		WantErr: true,
	}, {
		Name: "secret and CRD exist, missing other stuff",
		Key:  key,
		Objects: []runtime.Object{
			secret,
			&apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
								},
							},
						},
					},
				},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
									// Path is added.
									Path: ptr.String(path),
								},
								// CABundle is added.
								CABundle: []byte("present"),
							},
						},
					},
				},
			},
		}},
	}, {
		Name: "secret and CRD exist, incorrect fields",
		Key:  key,
		Objects: []runtime.Object{
			secret,
			&apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
									// Incorrect path
									Path: ptr.String("/incorrect"),
								},
								// CABundle is added.
								CABundle: []byte("incorrect"),
							},
						},
					},
				},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
									// Path is added.
									Path: ptr.String(path),
								},
								// CABundle is added.
								CABundle: []byte("present"),
							},
						},
					},
				},
			},
		}},
	}, {
		Name:    "failed to update custom resource definition",
		Key:     key,
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("update", "customresourcedefinitions"),
		},
		Objects: []runtime.Object{
			secret,
			&apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
									// Incorrect path
									Path: ptr.String("/incorrect"),
								},
								// CABundle is added.
								CABundle: []byte("incorrect"),
							},
						},
					},
				},
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: &apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
									// Path is added.
									Path: ptr.String(path),
								},
								// CABundle is added.
								CABundle: []byte("present"),
							},
						},
					},
				},
			},
		}},
	}, {
		Name: "stable",
		Key:  key,
		Objects: []runtime.Object{
			secret,
			&apixv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: key,
				},
				Spec: apixv1.CustomResourceDefinitionSpec{
					Conversion: &apixv1.CustomResourceConversion{
						Strategy: apixv1.WebhookConverter,
						Webhook: &apixv1.WebhookConversion{
							ClientConfig: &apixv1.WebhookClientConfig{
								Service: &apixv1.ServiceReference{
									Namespace: system.Namespace(),
									Name:      "webhook",
									Path:      ptr.String(path),
								},
								// CABundle is added.
								CABundle: []byte("present"),
							},
						},
					},
				},
			},
		},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &reconciler{
			kinds:        kinds,
			path:         path,
			secretName:   secretName,
			secretLister: listers.GetSecretLister(),
			crdLister:    listers.GetCustomResourceDefinitionLister(),
			client:       apixclient.Get(ctx),
		}
	}))
}
