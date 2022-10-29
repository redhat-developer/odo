package deploy

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
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

func (o *DeployClient) Deploy(fs filesystem.Filesystem, devfileObj parser.DevfileObj, path string, appName string, componentName string) error {
	deployHandler := newDeployHandler(fs, devfileObj, path, o.kubeClient, appName, componentName)
	return libdevfile.Deploy(devfileObj, deployHandler)
}

type deployHandler struct {
	fs            filesystem.Filesystem
	devfileObj    parser.DevfileObj
	path          string
	kubeClient    kclient.ClientInterface
	appName       string
	componentName string
}

var _ libdevfile.Handler = (*deployHandler)(nil)

func newDeployHandler(fs filesystem.Filesystem, devfileObj parser.DevfileObj, path string, kubeClient kclient.ClientInterface, appName string, componentName string) *deployHandler {
	return &deployHandler{
		fs:            fs,
		devfileObj:    devfileObj,
		path:          path,
		kubeClient:    kubeClient,
		appName:       appName,
		componentName: componentName,
	}
}

// ApplyImage builds and pushes the OCI image to be used on Kubernetes
func (o *deployHandler) ApplyImage(img v1alpha2.Component) error {
	return image.BuildPushSpecificImage(o.fs, o.path, img, true)
}

// ApplyKubernetes applies inline Kubernetes YAML from the devfile.yaml file
func (o *deployHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	return component.ApplyKubernetes(odolabels.ComponentDeployMode, o.appName, o.componentName, o.devfileObj, kubernetes, o.kubeClient, o.path)
}

// Execute will deploy the listed information in the `exec` section of devfile.yaml
// We currently do NOT support this in `odo deploy`.
func (o *deployHandler) Execute(command v1alpha2.Command) error {
	// TODO:
	// * Make sure we inject the "deploy" mode label once we implement exec in `odo deploy`
	// * Make sure you inject the "component type" label once we implement exec.

	// get the containers matching the command
	containers, err := libdevfile.GetContainerComponentsForCommand(o.devfileObj, command)
	if err != nil {
		return err
	}
	if len(containers) != 1 {
		return libdevfile.NewNotExactlyOneContainer()
	}
	containerName := containers[0]

	// get all the container components from the devfile and find the one we are interested in
	devfileContainers, err := o.devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{v1alpha2.ContainerComponentType},
	})
	if err != nil {
		return err
	}
	var devfileContainer v1alpha2.Component
	var found bool
	for _, dc := range devfileContainers {
		if dc.Name == containerName {
			found = true
			devfileContainer = dc
		}
	}
	if !found {
		return libdevfile.NewNotExactlyOneContainer()
	}

	job, err := libdevfile.GetJobFromContainerWithCommand(devfileContainer, command)
	if err != nil {
		return err
	}
	job.Labels = odolabels.GetLabels(o.devfileObj.GetMetadataName(), "app", "", odolabels.ComponentDeployMode, false)

	return o.kubeClient.CreateJob(job)
}
