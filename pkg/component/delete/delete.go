package delete

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/util"
)

type DeleteComponentClient struct {
	kubeClient kclient.ClientInterface
}

func NewDeleteComponentClient(kubeClient kclient.ClientInterface) *DeleteComponentClient {
	return &DeleteComponentClient{
		kubeClient: kubeClient,
	}
}

// ListClusterResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component
// It only returns resources not owned by another resource of the component, letting the garbage collector do its job
func (do *DeleteComponentClient) ListClusterResourcesToDelete(componentName string, namespace string) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured
	labels := componentlabels.GetLabels(componentName, "app", false)
	labels[applabels.ManagedBy] = "odo"
	selector := util.ConvertLabelsToSelector(labels)
	list, err := do.kubeClient.GetAllResourcesFromSelector(selector, namespace)
	if err != nil {
		return nil, err
	}
	for _, resource := range list {
		// If the resource is Terminating, there is no sense in displaying it.
		if resource.GetDeletionTimestamp() != nil {
			continue
		}
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

func (do *DeleteComponentClient) DeleteResources(resources []unstructured.Unstructured, wait bool) []unstructured.Unstructured {
	var failed []unstructured.Unstructured
	for _, resource := range resources {
		gvr, err := do.kubeClient.GetRestMappingFromUnstructured(resource)
		if err != nil {
			failed = append(failed, resource)
			continue
		}
		err = do.kubeClient.DeleteDynamicResource(resource.GetName(), gvr.Resource, wait)
		if err != nil {
			klog.V(3).Infof("failed to delete resource %q (%s.%s.%s): %v", resource.GetName(), gvr.Resource.Group, gvr.Resource.Version, gvr.Resource.Resource, err)
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

// ListResourcesToDeleteFromDevfile parses all the devfile components and returns a list of resources that are present on the cluster and can be deleted
func (do DeleteComponentClient) ListResourcesToDeleteFromDevfile(devfileObj parser.DevfileObj, appName string) (isInnerLoopDeployed bool, resources []unstructured.Unstructured, err error) {
	// Inner Loop
	// Fetch the deployment of the devfile component
	componentName := devfileObj.GetMetadataName()
	var deploymentName string
	deploymentName, err = util.NamespaceKubernetesObject(componentName, appName)
	if err != nil {
		return isInnerLoopDeployed, resources, fmt.Errorf("Failed to get the resource %q name for component %q; cause: %w", kclient.DeploymentKind, deploymentName, err)
	}

	deployment, err := do.kubeClient.GetDeploymentByName(deploymentName)
	if err != nil && !kerrors.IsNotFound(err) {
		return isInnerLoopDeployed, resources, err
	}

	// if the deployment is found on the cluster,
	// then convert it to unstructured.Unstructured object so that it can be appended to resources;
	// else continue to outer loop
	if deployment.Name != "" {
		isInnerLoopDeployed = true
		var unstructuredDeploy unstructured.Unstructured
		unstructuredDeploy, err = kclient.ConvertK8sResourceToUnstructured(deployment)
		if err != nil {
			return isInnerLoopDeployed, resources, fmt.Errorf("Failed to parse the resource %q: %q; cause: %w", kclient.DeploymentKind, deploymentName, err)
		}
		resources = append(resources, unstructuredDeploy)
	}

	// Outer Loop
	// Parse the devfile for outerloop K8s resources
	localResources, err := libdevfile.ListKubernetesComponents(devfileObj, devfileObj.Ctx.GetAbsPath())
	if err != nil {
		return isInnerLoopDeployed, resources, fmt.Errorf("Failed to gather resources for deletion: %w", err)
	}
	for _, lr := range localResources {
		var gvr *meta.RESTMapping
		gvr, err = do.kubeClient.GetRestMappingFromUnstructured(lr)
		if err != nil {
			continue
		}
		// Try to fetch the resource from the cluster; if it exists, append it to the resources list
		var cr *unstructured.Unstructured
		cr, err = do.kubeClient.GetDynamicResource(gvr.Resource, lr.GetName())
		if err != nil {
			continue
		}
		resources = append(resources, *cr)
	}
	return isInnerLoopDeployed, resources, nil
}

// ExecutePreStopEvents executes preStop events if any, as a precondition to deleting a devfile component deployment
func (do *DeleteComponentClient) ExecutePreStopEvents(devfileObj parser.DevfileObj, appName string) error {
	if !libdevfile.HasPreStopEvents(devfileObj) {
		return nil
	}
	componentName := devfileObj.GetMetadataName()
	klog.V(4).Infof("Gathering information for component: %q", componentName)

	klog.V(3).Infof("Checking component status for %q", componentName)
	selector := componentlabels.GetSelector(componentName, appName)
	pod, err := do.kubeClient.GetOnePodFromSelector(selector)
	if err != nil {
		klog.V(1).Info("Component not found on the cluster.")

		if kerrors.IsForbidden(err) {
			klog.V(3).Infof("Resource for %q forbidden", componentName)
			log.Warningf("You are forbidden from accessing the resource. Please check if you the right permissions and try again.")
			return nil
		}

		if e, ok := err.(*kclient.PodNotFoundError); ok {
			klog.V(3).Infof("Resource for %q not found; cause: %v", componentName, e)
			log.Warningf("Resources not found on the cluster. Run `odo delete component -v <DEBUG_LEVEL_0-9>` to know more.")
			return nil
		}

		return fmt.Errorf("unable to determine if component %s exists; cause: %v", componentName, err.Error())
	}

	// do not fail Delete operation if if the pod is not running or if the event execution fails
	if pod.Status.Phase != corev1.PodRunning {
		klog.V(4).Infof("unable to execute preStop events, pod for component %q is not running", componentName)
		return nil
	}

	klog.V(4).Infof("Executing %q event commands for component %q", libdevfile.PreStop, componentName)
	// ignore the failures if any; delete should not fail because preStop events failed to execute
	err = libdevfile.ExecPreStopEvents(devfileObj, component.NewExecHandler(do.kubeClient, pod.Name, false))
	if err != nil {
		klog.V(4).Infof("Failed to execute %q event commands for component %q, cause: %v", libdevfile.PreStop, componentName, err.Error())
	}

	return nil
}
