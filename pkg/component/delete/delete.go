package delete

import (
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type DeleteComponentClient struct {
	kubeClient kclient.ClientInterface
}

func NewDeleteComponentClient(kubeClient kclient.ClientInterface) *DeleteComponentClient {
	return &DeleteComponentClient{
		kubeClient: kubeClient,
	}
}

// ListResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component
// It only returns resources not owned by another resource of the component, letting the garbage collector do its job
func (do *DeleteComponentClient) ListResourcesToDelete(componentName string, namespace string) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured
	labels := componentlabels.GetLabels(componentName, "app", false)
	labels[applabels.ManagedBy] = "odo"
	selector := util.ConvertLabelsToSelector(labels)
	list, err := do.kubeClient.GetAllResourcesFromSelector(selector, namespace)
	if err != nil {
		return nil, err
	}
	for _, resource := range list {
		referenced := false
		for _, ownerRef := range resource.GetOwnerReferences() {
			if references(list, ownerRef) {
				referenced = true
				break
			}
		}
		if !referenced {
			result = append(result, resource)
		}
	}

	return result, nil
}

func (do *DeleteComponentClient) DeleteResources(resources []unstructured.Unstructured) []unstructured.Unstructured {
	var failed []unstructured.Unstructured
	for _, resource := range resources {
		gvr, err := do.kubeClient.GetRestMappingFromUnstructured(resource)
		if err != nil {
			failed = append(failed, resource)
			continue
		}
		err = do.kubeClient.DeleteDynamicResource(resource.GetName(), gvr.Resource.Group, gvr.Resource.Version, gvr.Resource.Resource)
		if err != nil {
			failed = append(failed, resource)
		}
	}
	return failed
}

// references returns true if ownerRef references a resource in the list
func references(list []unstructured.Unstructured, ownerRef metav1.OwnerReference) bool {
	for _, resource := range list {
		if ownerRef.APIVersion == resource.GetAPIVersion() && ownerRef.Kind == resource.GetKind() && ownerRef.Name == resource.GetName() {
			return true
		}
	}
	return false
}
