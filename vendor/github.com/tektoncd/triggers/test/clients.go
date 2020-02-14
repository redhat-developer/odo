/*
Copyright 2019 Tekton Authors LLC
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

/*
Get access to client objects

To initialize client objects you can use the newClients function. It returns a clients struct
that contains initialized clients for accessing:

	- Kubernetes resources
	- Pipelines resources (https://github.com/tektoncd/pipeline)
	- Triggers resources (https://github.com/tektoncd/triggers)

*/

package test

import (
	"testing"

	pipelineclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	triggersclientset "github.com/tektoncd/triggers/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	knativetest "knative.dev/pkg/test"
)

// clients holds instances of interfaces for making requests to the Pipeline controllers.
type clients struct {
	KubeClient kubernetes.Interface

	PipelineClient pipelineclientset.Interface
	TriggersClient triggersclientset.Interface
}

// newClients instantiates and returns several clientsets required for making requests to the
// cluster specified by the combination of clusterName and configPath.
func newClients(t *testing.T, configPath, clusterName string) *clients {
	t.Helper()
	var err error
	c := &clients{}

	cfg, err := knativetest.BuildClientConfig(configPath, clusterName)
	if err != nil {
		t.Fatalf("Failed to create configuration obj from %s for cluster %s: %s", configPath, clusterName, err)
	}

	c.KubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create kubernetes clientset from config file at %s: %s", configPath, err)
	}

	c.PipelineClient, err = pipelineclientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create pipeline clientset from config file at %s: %s", configPath, err)
	}

	c.TriggersClient, err = triggersclientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create triggers clientset from config file at %s: %s", configPath, err)
	}
	return c
}
