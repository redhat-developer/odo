package delete

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/devfile/library/pkg/devfile/parser"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/util"
)

type DeleteComponentClient struct {
	kubeClient   kclient.ClientInterface
	podmanClient podman.Client
	execClient   exec.Client
}

var _ Client = (*DeleteComponentClient)(nil)

func NewDeleteComponentClient(
	kubeClient kclient.ClientInterface,
	podmanClient podman.Client,
	execClient exec.Client,
) *DeleteComponentClient {
	return &DeleteComponentClient{
		kubeClient:   kubeClient,
		podmanClient: podmanClient,
		execClient:   execClient,
	}
}

// ListClusterResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component
// It only returns resources not owned by another resource of the component, letting the garbage collector do its job
func (do *DeleteComponentClient) ListClusterResourcesToDelete(ctx context.Context, componentName string, namespace string) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured
	selector := odolabels.GetSelector(componentName, odocontext.GetApplication(ctx), odolabels.ComponentAnyMode, false)
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
		if err != nil && !kerrors.IsNotFound(err) {
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
// Returns a Warning if an error happens communicating with the cluster
func (do DeleteComponentClient) ListClusterResourcesToDeleteFromDevfile(devfileObj parser.DevfileObj, appName string, componentName string, mode string) (isInnerLoopDeployed bool, resources []unstructured.Unstructured, err error) {
	var deployment *v1.Deployment
	if mode == odolabels.ComponentDevMode || mode == odolabels.ComponentAnyMode {
		// Inner Loop
		var deploymentName string
		deploymentName, err = util.NamespaceKubernetesObject(componentName, appName)
		if err != nil {
			return isInnerLoopDeployed, resources, fmt.Errorf("failed to get the resource %q name for component %q; cause: %w", kclient.DeploymentKind, componentName, err)
		}

		deployment, err = do.kubeClient.GetDeploymentByName(deploymentName)
		if err != nil && !kerrors.IsNotFound(err) {
			// Kubernetes cluster access fails, return with a warning only
			err = clierrors.NewWarning(fmt.Sprintf("failed to get deployment %q", deploymentName), err)
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
				return isInnerLoopDeployed, resources, fmt.Errorf("failed to parse the resource %q: %q; cause: %w", kclient.DeploymentKind, deploymentName, err)
			}
			resources = append(resources, unstructuredDeploy)
		}
	}

	// Parse the devfile for K8s resources; these may belong to either innerloop or outerloop
	localResources, err := libdevfile.ListKubernetesComponents(devfileObj, filepath.Dir(devfileObj.Ctx.GetAbsPath()))
	if err != nil {
		return isInnerLoopDeployed, resources, fmt.Errorf("failed to gather resources for deletion: %w", err)
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
		// If a specific mode is asked for, then make sure it matches with the cr's mode.
		if err != nil || (mode != odolabels.ComponentAnyMode && odolabels.GetMode(cr.GetLabels()) != mode) {
			if cr != nil {
				klog.V(4).Infof("Ignoring resource: %s/%s; its mode(%s) does not match with the given mode(%s)", gvr.Resource.Resource, lr.GetName(), odolabels.GetMode(cr.GetLabels()), mode)
			} else {
				klog.V(4).Infof("Ignoring resource: %s/%s; it does not exist on the cluster", gvr.Resource.Resource, lr.GetName())
			}
			continue
		}
		resources = append(resources, *cr)
	}

	return isInnerLoopDeployed, resources, nil
}

// ExecutePreStopEvents executes preStop events if any, as a precondition to deleting a devfile component deployment
func (do *DeleteComponentClient) ExecutePreStopEvents(devfileObj parser.DevfileObj, appName string, componentName string) error {
	if !libdevfile.HasPreStopEvents(devfileObj) {
		return nil
	}

	klog.V(4).Infof("Gathering information for component: %q", componentName)

	klog.V(3).Infof("Checking component status for %q", componentName)
	selector := odolabels.GetSelector(componentName, appName, odolabels.ComponentDevMode, false)
	pod, err := do.kubeClient.GetRunningPodFromSelector(selector)
	if err != nil {
		klog.V(1).Info("Component not found on the cluster.")

		if kerrors.IsForbidden(err) {
			klog.V(3).Infof("Resource for %q forbidden", componentName)
			log.Warningf("You are forbidden from accessing the resource. Please check if you the right permissions and try again.")
			return nil
		}

		if e, ok := err.(*platform.PodNotFoundError); ok {
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
	err = libdevfile.ExecPreStopEvents(devfileObj, component.NewExecHandler(do.kubeClient, do.execClient, appName, componentName, pod.Name, "", false))
	if err != nil {
		klog.V(4).Infof("Failed to execute %q event commands for component %q, cause: %v", libdevfile.PreStop, componentName, err.Error())
	}

	return nil
}

func (do *DeleteComponentClient) ListPodmanResourcesToDelete(appName string, componentName string) (isInnerLoopDeployed bool, pods []*corev1.Pod, err error) {
	// Inner Loop
	var podName string
	podName, err = util.NamespaceKubernetesObject(componentName, appName)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get the resource %q name for component %q; cause: %w", kclient.DeploymentKind, componentName, err)
	}

	allPods, err := do.podmanClient.PodLs()
	if err != nil {
		err = clierrors.NewWarning("failed to get pods on podman", err)
		return false, nil, err
	}

	if _, isInnerLoopDeployed = allPods[podName]; isInnerLoopDeployed {
		podDef, err := do.podmanClient.KubeGenerate(podName)
		if err != nil {
			return false, nil, err
		}
		pods = append(pods, podDef)
	}
	return isInnerLoopDeployed, pods, nil
}
