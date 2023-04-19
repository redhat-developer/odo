package delete

import (
	"context"

	"github.com/devfile/library/v2/pkg/devfile/parser"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Client interface {
	// ListClusterResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component.
	// The mode indicates which component to list, either Dev, Deploy or Any (using constant labels.Component*Mode).
	ListClusterResourcesToDelete(ctx context.Context, componentName string, namespace string, mode string) ([]unstructured.Unstructured, error)
	// DeleteResources deletes the unstructured resources and return the resources that failed to be deleted
	// set wait to true to wait for all the dependencies to be deleted
	DeleteResources(resources []unstructured.Unstructured, wait bool) []unstructured.Unstructured
	// ExecutePreStopEvents executes preStop events if any, as a precondition to deleting a devfile component deployment
	ExecutePreStopEvents(ctx context.Context, devfileObj parser.DevfileObj, appName string, componentName string) error
	// ListClusterResourcesToDeleteFromDevfile parses all the devfile components and returns a list of resources that are present on the cluster that can be deleted,
	// and a bool that indicates if the devfile component has been pushed to the innerloop.
	// The mode indicates which component to list, either Dev, Deploy or Any (using constant labels.Component*Mode).
	ListClusterResourcesToDeleteFromDevfile(devfileObj parser.DevfileObj, appName string, componentName string, mode string) (bool, []unstructured.Unstructured, error)
	// ListPodmanResourcesToDelete returns a list of resources that are present on podman in Dev mode that can be deleted for the given component/app,
	// and a bool that indicates if the devfile component has been pushed to the innerloop.
	// The mode indicates which component to list, either Dev, Deploy or Any (using constant labels.Component*Mode).
	ListPodmanResourcesToDelete(appName string, componentName string, mode string) (isInnerLoopDeployed bool, pods []*corev1.Pod, err error)
}
