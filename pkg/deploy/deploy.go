package deploy

import (
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/pkg/errors"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/service"
)

type DeployClient struct {
	kubeClient kclient.ClientInterface
}

func NewDeployClient(kubeClient kclient.ClientInterface) *DeployClient {
	return &DeployClient{
		kubeClient: kubeClient,
	}
}

func (o *DeployClient) Deploy(devfileObj parser.DevfileObj, path string, appName string) error {
	deployHandler := newDeployHandler(devfileObj, path, o.kubeClient, appName)
	return libdevfile.Deploy(devfileObj, deployHandler)
}

type deployHandler struct {
	devfileObj parser.DevfileObj
	path       string
	kubeClient kclient.ClientInterface
	appName    string
}

func newDeployHandler(devfileObj parser.DevfileObj, path string, kubeClient kclient.ClientInterface, appName string) *deployHandler {
	return &deployHandler{
		devfileObj: devfileObj,
		path:       path,
		kubeClient: kubeClient,
		appName:    appName,
	}
}

func (o *deployHandler) ApplyImage(img v1alpha2.Component) error {
	return image.BuildPushSpecificImage(o.devfileObj, o.path, img, true)
}

func (o *deployHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	// validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	_, err := service.ValidateResourceExist(o.kubeClient, kubernetes, o.path)
	if err != nil {
		return err
	}

	labels := componentlabels.GetLabels(kubernetes.Name, o.appName, true)
	u, err := service.GetK8sComponentAsUnstructured(kubernetes.Kubernetes, o.path, devfilefs.DefaultFs{})
	if err != nil {
		return err
	}

	log.Infof("\nDeploying Kubernetes %s: %s", u.GetKind(), u.GetName())
	isOperatorBackedService, err := service.PushKubernetesResource(o.kubeClient, u, labels)
	if err != nil {
		return errors.Wrap(err, "failed to create service(s) associated with the component")
	}
	if isOperatorBackedService {
		log.Successf("Kubernetes resource %q on the cluster; refer %q to know how to link it to the component", strings.Join([]string{u.GetKind(), u.GetName()}, "/"), "odo link -h")

	}
	return nil
}

func (o *deployHandler) Execute(command v1alpha2.Command) error {
	return errors.New("Exec command is not implemented for Deploy")
}
