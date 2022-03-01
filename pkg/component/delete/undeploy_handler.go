package delete

import (
	"errors"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"k8s.io/klog"
)

type undeployHandler struct {
	path       string
	kubeClient kclient.ClientInterface
}

func newUndeployHandler(path string, kubeClient kclient.ClientInterface) *undeployHandler {
	return &undeployHandler{
		path:       path,
		kubeClient: kubeClient,
	}
}

func (o *undeployHandler) ApplyImage(image v1alpha2.Component) error {
	return nil
}

func (o *undeployHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	// Parse the component's Kubernetes manifest
	u, err := libdevfile.GetK8sComponentAsUnstructured(kubernetes.Kubernetes, o.path, devfilefs.DefaultFs{})
	if err != nil {
		return err
	}

	// Get the REST mappings
	gvr, err := o.kubeClient.GetRestMappingFromUnstructured(u)
	if err != nil {
		return err
	}
	klog.V(4).Infof("Un-deploying the Kubernetes %s: %s", u.GetKind(), u.GetName())
	// Un-deploy the K8s manifest
	err = o.kubeClient.DeleteDynamicResource(u.GetName(), gvr.Resource.Group, gvr.Resource.Version, gvr.Resource.Resource)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			klog.V(3).Infof("Resource for %s not found; cause: %v", u.GetName(), err)
			return nil
		}
	}
	return err
}

func (o *undeployHandler) Execute(command v1alpha2.Command) error {
	return errors.New("Exec command is not implemented for Deploy")
}
