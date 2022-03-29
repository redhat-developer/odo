package delete

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Client interface {
	// ListClusterResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component
	ListClusterResourcesToDelete(componentName string, namespace string) ([]unstructured.Unstructured, error)
	// DeleteResources deletes the unstuctured resources and return the resources that failed to be deleted
	DeleteResources([]unstructured.Unstructured) []unstructured.Unstructured
	// ExecutePreStopEvents executes preStop events if any, as a precondition to deleting a devfile component deployment
	ExecutePreStopEvents(devfileObj parser.DevfileObj, appName string) error
	// ListResourcesToDeleteFromDevfile parses all the devfile components and returns a list of resources that are present on the cluster that can be deleted,
	// and a bool that indicates if the devfile component has been pushed to the innerloop
	ListResourcesToDeleteFromDevfile(devfileObj parser.DevfileObj, appName string) (bool, []unstructured.Unstructured, error)
}
