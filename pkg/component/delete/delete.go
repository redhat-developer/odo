package delete

import (
	"fmt"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
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
	log.Printf("Gathering information for component: %q", componentName)

	pod, err := do.getPod(componentName, appName)

	// if there are preStop events, execute them before deleting the deployment
	if libdevfile.HasPreStopEvents(devfileObj) {
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("unable to execute preStop events, pod for component %s is not running", componentName)
		}
		log.Infof("\nExecuting %s event commands for component %s", libdevfile.PreStop, componentName)
		err = libdevfile.ExecPreStopEvents(devfileObj, componentName, component.NewExecHandler(do.kubeClient, pod.Name, false))
		if err != nil {
			return err
		}
	}

	log.Infof("\nDeleting component %s", componentName)
	spinner := log.Spinner("Deleting Kubernetes resources for component")
	defer spinner.End(false)

	err = do.kubeClient.Delete(labels, false)
	if err != nil {
		return err
	}

	spinner.End(true)
	log.Successf("Successfully deleted component")
	return nil
}

// getPod returns a pod, and an error based on the component name and app name;
// it returns a nil error if the resource is forbidden or if the pod is not found
func (o DeleteComponentClient) getPod(componentName, appName string) (pod *corev1.Pod, err error) {
	// podSpinner := log.Spinner("Checking status for component")
	// defer podSpinner.End(false)

	selector := componentlabels.GetSelector(componentName, appName)
	pod, err = o.kubeClient.GetOnePodFromSelector(selector)
	if kerrors.IsForbidden(err) {
		klog.V(2).Infof("Resource for %s forbidden", componentName)
		// log the error if it failed to determine if the component exists due to insufficient RBACs
		// podSpinner.End(false)
		log.Warningf("%v", err)
		return pod, nil
	} else if e, ok := err.(*kclient.PodNotFoundError); ok {
		// podSpinner.End(false)
		log.Warningf("%v", e)
		return pod, nil
	} else if err != nil {
		return pod, errors.Wrapf(err, "unable to determine if component %s exists", componentName)
	}

	// podSpinner.End(true)
	return
}
