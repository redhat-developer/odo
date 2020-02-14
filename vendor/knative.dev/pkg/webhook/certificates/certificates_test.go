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

package certificates

import (
	"context"
	"errors"
	"testing"

	kubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/core/v1/secret/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/system"
	"knative.dev/pkg/webhook"
	certresources "knative.dev/pkg/webhook/certificates/resources"

	. "knative.dev/pkg/reconciler/testing"
	. "knative.dev/pkg/webhook/testing"
)

func TestReconcile(t *testing.T) {
	secretName, serviceName := "webhook-secret", "webhook-service"
	secret, err := certresources.MakeSecret(context.Background(),
		secretName, system.Namespace(), serviceName)
	if err != nil {
		t.Fatalf("MakeSecret() = %v", err)
	}

	// Mutate the MakeSecret to return our secret deterministically.
	certresources.MakeSecret = func(ctx context.Context, name, namespace, serviceName string) (*corev1.Secret, error) {
		return secret, nil
	}
	defer func() {
		certresources.MakeSecret = certresources.MakeSecretInternal
	}()

	// The key to use, which for this singleton reconciler doesn't matter (although the
	// namespace matters for namespace validation).
	key := system.Namespace() + "/does not matter"

	table := TableTest{{
		Name:    "well formed secret exists",
		Key:     key,
		Objects: []runtime.Object{secret},
	}, {
		Name:        "secret does not exist",
		Key:         key,
		WantCreates: []runtime.Object{secret},
	}, {
		Name:    "secret does not exist, create error",
		Key:     key,
		WantErr: true,
		WithReactors: []clientgotesting.ReactionFunc{
			InduceFailure("create", "secrets"),
		},
		WantCreates: []runtime.Object{secret},
	}, {
		Name: "missing server key",
		Key:  key,
		Objects: []runtime.Object{&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: system.Namespace(),
			},
			Data: map[string][]byte{
				// certresources.ServerKey:  []byte("missing"),
				certresources.ServerCert: []byte("present"),
				certresources.CACert:     []byte("present"),
			},
		}},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: secret,
		}},
	}, {
		Name: "missing server cert",
		Key:  key,
		Objects: []runtime.Object{&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: system.Namespace(),
			},
			Data: map[string][]byte{
				certresources.ServerKey: []byte("present"),
				// certresources.ServerCert: []byte("missing"),
				certresources.CACert: []byte("present"),
			},
		}},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: secret,
		}},
	}, {
		Name: "missing CA cert",
		Key:  key,
		Objects: []runtime.Object{&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: system.Namespace(),
			},
			Data: map[string][]byte{
				certresources.ServerKey:  []byte("present"),
				certresources.ServerCert: []byte("present"),
				// certresources.CACert: []byte("missing"),
			},
		}},
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: secret,
		}},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &reconciler{
			client:       kubeclient.Get(ctx),
			secretlister: listers.GetSecretLister(),
			secretName:   secretName,
			serviceName:  serviceName,
		}
	}))
}

func TestReconcileMakeSecretFailure(t *testing.T) {
	secretName, serviceName := "webhook-secret", "webhook-service"
	secret, err := certresources.MakeSecret(context.Background(),
		secretName, system.Namespace(), serviceName)
	if err != nil {
		t.Fatalf("MakeSecret() = %v", err)
	}

	// Mutate the MakeSecret to return our secret deterministically.
	certresources.MakeSecret = func(ctx context.Context, name, namespace, serviceName string) (*corev1.Secret, error) {
		return nil, errors.New("this is an error")
	}
	defer func() {
		certresources.MakeSecret = certresources.MakeSecretInternal
	}()

	// The key to use, which for this singleton reconciler doesn't matter (although the
	// namespace matters for namespace validation).
	key := system.Namespace() + "/does not matter"

	table := TableTest{{
		Name:    "would return error, but not called",
		Key:     key,
		Objects: []runtime.Object{secret},
	}, {
		Name:    "secret does not exist",
		Key:     key,
		WantErr: true,
	}, {
		Name: "missing server key",
		Key:  key,
		Objects: []runtime.Object{&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: system.Namespace(),
			},
			Data: map[string][]byte{
				// certresources.ServerKey:  []byte("missing"),
				certresources.ServerCert: []byte("present"),
				certresources.CACert:     []byte("present"),
			},
		}},
		WantErr: true,
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &reconciler{
			client:       kubeclient.Get(ctx),
			secretlister: listers.GetSecretLister(),
			secretName:   secretName,
			serviceName:  serviceName,
		}
	}))
}

func TestNew(t *testing.T) {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{})

	c := NewController(ctx, configmap.NewStaticWatcher())
	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}
}
