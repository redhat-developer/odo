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
	"fmt"
	"log"
	"net/http"

	"go.uber.org/zap"

	dynamicClientset "github.com/tektoncd/triggers/pkg/client/dynamic/clientset"
	"github.com/tektoncd/triggers/pkg/client/dynamic/clientset/tekton"
	"github.com/tektoncd/triggers/pkg/logging"
	"github.com/tektoncd/triggers/pkg/sink"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/signals"
)

const (
	// EventListenerLogKey is the name of the logger for the eventlistener cmd
	EventListenerLogKey = "eventlistener"
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
		log.Fatalf("Failed to get the Kubernetes client set: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("Failed to get the dynamic client: %v", err)
	}
	dynamicCS := dynamicClientset.New(tekton.WithClient(dynamicClient))

	logger := logging.ConfigureLogging(EventListenerLogKey, ConfigName, stopCh, kubeClient)
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Fatalf("Failed to sync the logger", zap.Error(err))
		}
	}()

	logger.Info("EventListener pod started")

	sinkArgs, err := sink.GetArgs()
	if err != nil {
		logger.Fatal(err)
	}

	sinkClients, err := sink.ConfigureClients()
	if err != nil {
		logger.Fatal(err)
	}

	// Create EventListener Sink
	r := sink.Sink{
		KubeClientSet:          kubeClient,
		DiscoveryClient:        sinkClients.DiscoveryClient,
		DynamicClient:          dynamicCS,
		TriggersClient:         sinkClients.TriggersClient,
		PipelineClient:         sinkClients.PipelineClient,
		HTTPClient:             http.DefaultClient,
		EventListenerName:      sinkArgs.ElName,
		EventListenerNamespace: sinkArgs.ElNamespace,
		Logger:                 logger,
	}

	// Listen and serve
	logger.Infof("Listen and serve on port %s", sinkArgs.Port)
	http.HandleFunc("/", r.HandleEvent)
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", sinkArgs.Port), nil))
}
