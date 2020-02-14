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

package sink

import (
	"flag"

	"golang.org/x/xerrors"

	pipelineclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	triggersclientset "github.com/tektoncd/triggers/pkg/client/clientset/versioned"
	discoveryclient "k8s.io/client-go/discovery"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
)

const (
	// Flag definitions
	name      = "el-name"
	namespace = "el-namespace"
	port      = "port"
)

var (
	nameFlag = flag.String("el-name", "",
		"The name of the EventListener resource for this sink.")
	namespaceFlag = flag.String("el-namespace", "",
		"The namespace of the EventListener resource for this sink.")
	portFlag = flag.String("port", "",
		"The port for the EventListener sink to listen on.")
)

// Args define the arguments for Sink.
type Args struct {
	// ElName is the EventListener name.
	ElName string
	// ElNamespace is the EventListener namespace.
	ElNamespace string
	// Port is the port the Sink should listen on.
	Port string
}

// Clients define the set of client dependencies Sink requires.
type Clients struct {
	DiscoveryClient discoveryclient.DiscoveryInterface
	RESTClient      restclient.Interface
	TriggersClient  triggersclientset.Interface
	PipelineClient  pipelineclientset.Interface
}

// GetArgs returns the flagged Args
func GetArgs() (Args, error) {
	flag.Parse()
	if *nameFlag == "" {
		return Args{}, xerrors.Errorf("-%s arg not found", name)
	}
	if *namespaceFlag == "" {
		return Args{}, xerrors.Errorf("-%s arg not found", namespace)
	}
	if *portFlag == "" {
		return Args{}, xerrors.Errorf("-%s arg not found", port)
	}
	return Args{
		ElName:      *nameFlag,
		ElNamespace: *namespaceFlag,
		Port:        *portFlag,
	}, nil
}

// ConfigureClients returns the kubernetes and triggers clientsets
func ConfigureClients() (Clients, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return Clients{}, xerrors.Errorf("Failed to get in cluster config: %s", err)
	}
	kubeClient, err := kubeclientset.NewForConfig(clusterConfig)
	if err != nil {
		return Clients{}, xerrors.Errorf("Failed to create KubeClient: %s", err)
	}
	triggersClient, err := triggersclientset.NewForConfig(clusterConfig)
	if err != nil {
		return Clients{}, xerrors.Errorf("Failed to create TriggersClient: %s", err)
	}
	pipelineclient, err := pipelineclientset.NewForConfig(clusterConfig)
	if err != nil {
		return Clients{}, xerrors.Errorf("Failed to create PipelineClient: %s", err)
	}

	return Clients{
		DiscoveryClient: kubeClient.Discovery(),
		RESTClient:      kubeClient.RESTClient(),
		TriggersClient:  triggersClient,
		PipelineClient:  pipelineclient,
	}, nil
}
