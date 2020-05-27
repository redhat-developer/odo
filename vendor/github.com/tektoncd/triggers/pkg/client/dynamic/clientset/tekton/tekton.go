package tekton

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	"github.com/tektoncd/triggers/pkg/client/dynamic/clientset"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var allowedTektonTypes = map[string][]string{
	"v1alpha1": {"pipelineresources", "pipelineruns", "taskruns", "pipelines", "clustertasks", "tasks", "conditions"},
	"v1beta1":  {"pipelineruns", "taskruns", "pipelines", "clustertasks", "tasks"},
}

// WithClient adds Tekton related clients to the Dynamic client.
func WithClient(client dynamic.Interface) clientset.Option {
	return func(cs *clientset.Clientset) {
		for version, resources := range allowedTektonTypes {
			for _, resource := range resources {
				r := schema.GroupVersionResource{
					Group:    pipeline.GroupName,
					Version:  version,
					Resource: resource,
				}
				cs.Add(r, client)
			}
		}
	}
}
