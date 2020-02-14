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
	"log"

	"github.com/tektoncd/pipeline/pkg/system"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"github.com/tektoncd/triggers/pkg/logging"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
)

// WebhookLogKey is the name of the logger for the webhook cmd
const (
	WebhookLogKey = "webhook"
	// ConfigName is the name of the ConfigMap that the logging config will be stored in
	ConfigName = "config-logging-triggers"
)

func main() {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in cluster config: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("Failed to get the client set: %v", err)
	}

	logger := logging.ConfigureLogging(WebhookLogKey, ConfigName, stopCh, kubeClient)

	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Fatalf("Failed to sync the logger", zap.Error(err))
		}
	}()

	options := webhook.ControllerOptions{
		ServiceName:                     "tekton-triggers-webhook",
		DeploymentName:                  "tekton-triggers-webhook",
		Namespace:                       system.GetNamespace(),
		Port:                            8443,
		SecretName:                      "triggers-webhook-certs",
		WebhookName:                     "triggers-webhook.tekton.dev",
		ResourceAdmissionControllerPath: "/",
	}
	resourceHandlers := map[schema.GroupVersionKind]webhook.GenericCRD{
		v1alpha1.SchemeGroupVersion.WithKind("EventListener"):   &v1alpha1.EventListener{},
		v1alpha1.SchemeGroupVersion.WithKind("TriggerBinding"):  &v1alpha1.TriggerBinding{},
		v1alpha1.SchemeGroupVersion.WithKind("TriggerTemplate"): &v1alpha1.TriggerTemplate{},
	}
	resourceAdmissionController := webhook.NewResourceAdmissionController(resourceHandlers, options, true)
	admissionControllers := map[string]webhook.AdmissionController{
		options.ResourceAdmissionControllerPath: resourceAdmissionController,
	}

	// Decorate contexts with the current state of the config.
	ctxFunc := v1alpha1.WithUpgradeViaDefaulting

	controller, err := webhook.New(kubeClient, options, admissionControllers, logger, ctxFunc)
	if err != nil {
		logger.Fatal("Error creating admission controller", zap.Error(err))
	}

	if err := controller.Run(stopCh); err != nil {
		logger.Fatal("Error running admission controller", zap.Error(err))
	}
}
