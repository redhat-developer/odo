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

package main

import (
	"context"
	"os"

	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("ClusterTriggerBinding"): &v1alpha1.ClusterTriggerBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("EventListener"):         &v1alpha1.EventListener{},
	v1alpha1.SchemeGroupVersion.WithKind("TriggerBinding"):        &v1alpha1.TriggerBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("TriggerTemplate"):       &v1alpha1.TriggerTemplate{},
}

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"webhook.triggers.tekton.dev",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		v1alpha1.WithUpgradeViaDefaulting,

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"validation.webhook.triggers.tekton.dev",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.triggers.tekton.dev",

		// The path on which to serve the webhook.
		"/config-validation",

		configmap.Constructors{
			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
		},
	)
}

func main() {
	serviceName := os.Getenv("WEBHOOK_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "tekton-triggers-webhook"
	}

	secretName := os.Getenv("WEBHOOK_SECRET_NAME")
	if secretName == "" {
		secretName = "triggers-webhook-certs"
	}

	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: serviceName,
		Port:        8443,
		SecretName:  secretName,
	})

	sharedmain.MainWithContext(ctx, "webhook",
		certificates.NewController,
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
		NewConfigValidationController,
	)
}
