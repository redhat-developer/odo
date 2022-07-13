package deploy

import (
	"errors"
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	component2 "github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/service"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type DeployClient struct {
	kubeClient kclient.ClientInterface
}

var _ Client = (*DeployClient)(nil)

func NewDeployClient(kubeClient kclient.ClientInterface) *DeployClient {
	return &DeployClient{
		kubeClient: kubeClient,
	}
}

func (o *DeployClient) Deploy(fs filesystem.Filesystem, devfileObj parser.DevfileObj, path string, appName string) error {
	deployHandler := newDeployHandler(fs, devfileObj, path, o.kubeClient, appName)
	return libdevfile.Deploy(devfileObj, deployHandler)
}

type deployHandler struct {
	fs         filesystem.Filesystem
	devfileObj parser.DevfileObj
	path       string
	kubeClient kclient.ClientInterface
	appName    string
}

var _ libdevfile.Handler = (*deployHandler)(nil)

func newDeployHandler(fs filesystem.Filesystem, devfileObj parser.DevfileObj, path string, kubeClient kclient.ClientInterface, appName string) *deployHandler {
	return &deployHandler{
		fs:         fs,
		devfileObj: devfileObj,
		path:       path,
		kubeClient: kubeClient,
		appName:    appName,
	}
}

// ApplyImage builds and pushes the OCI image to be used on Kubernetes
func (o *deployHandler) ApplyImage(img v1alpha2.Component) error {
	return image.BuildPushSpecificImage(o.fs, o.path, img, true)
}

// ApplyKubernetes applies inline Kubernetes YAML from the devfile.yaml file
func (o *deployHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	// Validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	_, err := component2.ValidateResourceExist(o.kubeClient, o.devfileObj, kubernetes, o.path)
	if err != nil {
		return err
	}

	// Get the most common labels that's applicable to all resources being deployed.
	// Set the mode to DEPLOY. Regardless of what Kubernetes resource we are deploying.
	labels := odolabels.GetLabels(o.devfileObj.Data.GetMetadata().Name, o.appName, odolabels.ComponentDeployMode)
	klog.V(4).Infof("Injecting labels: %+v into k8s artifact", labels)

	// Create the annotations
	// Retrieve the component type from the devfile and also inject it into the list of annotations
	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(o.devfileObj.Data.GetMetadata()))

	// Get the Kubernetes component
	u, err := libdevfile.GetK8sComponentAsUnstructured(o.devfileObj, kubernetes.Name, o.path, devfilefs.DefaultFs{})
	if err != nil {
		return err
	}

	// Deploy the actual Kubernetes component and error out if there's an issue.
	log.Sectionf("Deploying Kubernetes Component: %s", u.GetName())
	isOperatorBackedService, err := service.PushKubernetesResource(o.kubeClient, u, labels, annotations)
	if err != nil {
		return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}

	if isOperatorBackedService {
		log.Successf("Kubernetes resource %q on the cluster; refer %q to know how to link it to the component", strings.Join([]string{u.GetKind(), u.GetName()}, "/"), "odo link -h")

	}
	return nil
}

// Execute will deploy the listed information in the `exec` section of devfile.yaml
// We currently do NOT support this in `odo deploy`.
func (o *deployHandler) Execute(command v1alpha2.Command) error {
	// TODO:
	// * Make sure we inject the "deploy" mode label once we implement exec in `odo deploy`
	// * Make sure you inject the "component type" label once we implement exec.
	return errors.New("exec command is not implemented for Deploy")
}
