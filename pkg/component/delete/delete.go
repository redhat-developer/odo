package delete

import (
	"fmt"
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
)

type DeleteComponentClient struct {
	kubeClient kclient.ClientInterface
}

func NewDeleteComponentClient(kubeClient kclient.ClientInterface) *DeleteComponentClient {
	return &DeleteComponentClient{
		kubeClient: kubeClient,
	}
}

func (o *DeleteComponentClient) ListKubernetesComponents(devfileObj parser.DevfileObj, path string) (list []unstructured.Unstructured, err error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: devfile.KubernetesComponentType},
	})
	if err != nil {
		return
	}
	var u unstructured.Unstructured
	for _, component := range components {
		if component.Kubernetes != nil {
			u, err = libdevfile.GetK8sComponentAsUnstructured(component.Kubernetes, path, devfilefs.DefaultFs{})
			if err != nil {
				return
			}
			list = append(list, u)
		}
	}
	return
}

func (o *DeleteComponentClient) UnDeploy(devfileObj parser.DevfileObj, path string) error {
	undeployHandler := newUndeployHandler(path, o.kubeClient)
	return libdevfile.Deploy(devfileObj, undeployHandler)
}

// DevfileComponentDelete deletes the devfile component
func (do *DeleteComponentClient) DeleteComponent(devfileObj parser.DevfileObj, componentName string) error {
	appName := "app"
	labels := componentlabels.GetLabels(componentName, appName, false)
	if labels == nil {
		return fmt.Errorf("cannot delete with labels being nil")
	}
	klog.V(4).Infof("Gathering information for component: %q", componentName)

	pod, err := do.getPod(componentName, appName)
	if err != nil {
		return err
	}

	// if there are preStop events, execute them before deleting the deployment
	if libdevfile.HasPreStopEvents(devfileObj) {
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("unable to execute preStop events, pod for component %s is not running", componentName)
		}
		klog.V(4).Infof("Executing %s event commands for component %s", libdevfile.PreStop, componentName)
		err = libdevfile.ExecPreStopEvents(devfileObj, componentName, component.NewExecHandler(do.kubeClient, pod.Name, false))
		if err != nil {
			return err
		}
	}

	klog.V(4).Infof("Deleting component %s", componentName)
	err = do.kubeClient.Delete(labels, false)
	if err != nil {
		return err
	}

	return nil
}

// getPod returns a pod, and an error based on the component name and app name;
// it returns a nil error if the resource is forbidden or if the pod is not found
func (o DeleteComponentClient) getPod(componentName, appName string) (pod *corev1.Pod, err error) {
	klog.V(3).Infof("Checking component status for %q", componentName)
	selector := componentlabels.GetSelector(componentName, appName)
	pod, err = o.kubeClient.GetOnePodFromSelector(selector)
	if err != nil {
		klog.V(1).Info("Component not found on the cluster.")
		if kerrors.IsForbidden(err) {
			klog.V(3).Infof("Resource for %s forbidden", componentName)
			return pod, nil
		} else if e, ok := err.(*kclient.PodNotFoundError); ok {
			klog.V(3).Infof("Resource for %s not found; cause: %v", componentName, e)
			return pod, nil
		}
		return pod, errors.Wrapf(err, "unable to determine if component %s exists", componentName)
	}

	return
}
