package delete

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Client interface {
	// ListClusterResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component
	ListClusterResourcesToDelete(componentName string, namespace string) ([]unstructured.Unstructured, error)
	// DeleteResources deletes the unstructured resources and return the resources that failed to be deleted
	// set wait to true to wait for all the dependencies to be deleted
	DeleteResources(resources []unstructured.Unstructured, wait bool) []unstructured.Unstructured
	// ExecutePreStopEvents executes preStop events if any, as a precondition to deleting a devfile component deployment
	ExecutePreStopEvents(devfileObj parser.DevfileObj, appName string) error
	// ListResourcesToDeleteFromDevfile parses all the devfile components and returns a list of resources that are present on the cluster that can be deleted,
	// and a bool that indicates if the devfile component has been pushed to the innerloop
	// the mode indicates which component to list, either Dev, Deploy or Any (using constant labels.Component*Mode)
	ListResourcesToDeleteFromDevfile(devfileObj parser.DevfileObj, appName string, mode string) (bool, []unstructured.Unstructured, error)
}
