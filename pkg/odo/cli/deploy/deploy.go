package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	img "github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/service"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/utils/pointer"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "deploy"

// DeployOptions encapsulates the options for the odo command
type DeployOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Flags
	contextFlag string
}

var deployExample = templates.Examples(`
  # Deploy components defined in the devfile
  %[1]s
`)

// NewDeployOptions creates a new DeployOptions instance
func NewDeployOptions() *DeployOptions {
	return &DeployOptions{}
}

func (o *DeployOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete DeployOptions after they've been created
func (o *DeployOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	containsDevfile, err := location.DirectoryContainsDevfile(filesystem.DefaultFs{}, cwd)
	if err != nil {
		return err
	}
	if !containsDevfile {
		devfileLocation, err2 := o.clientset.InitClient.SelectDevfile(map[string]string{}, o.clientset.FS, cwd)
		if err2 != nil {
			return err2
		}

		devfilePath, err2 := o.clientset.InitClient.DownloadDevfile(devfileLocation, cwd)
		if err2 != nil {
			return fmt.Errorf("unable to download devfile: %w", err2)
		}

		devfileObj, _, err2 := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: pointer.BoolPtr(false)})
		if err2 != nil {
			return fmt.Errorf("unable to download devfile: %w", err2)
		}

		// Set the name in the devfile and writes the devfile back to the disk
		err = o.clientset.InitClient.PersonalizeName(devfileObj, map[string]string{})
		if err != nil {
			return fmt.Errorf("failed to update the devfile's name: %w", err)
		}

	}
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag))
	if err != nil {
		return err
	}

	envFileInfo, err := envinfo.NewEnvSpecificInfo(o.contextFlag)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve configuration information")
	}
	if !envFileInfo.Exists() {
		var cmpName string
		cmpName, err = component.GatherName(o.EnvSpecificInfo.GetDevfileObj(), o.GetDevfilePath())
		if err != nil {
			return errors.Wrap(err, "unable to retrieve component name")
		}
		err = envFileInfo.SetComponentSettings(envinfo.ComponentSettings{Name: cmpName, Project: o.GetProject(), AppName: "app"})
		if err != nil {
			return errors.Wrap(err, "failed to write new env.yaml file")
		}

	} else if envFileInfo.GetComponentSettings().Project != o.GetProject() {
		err = envFileInfo.SetConfiguration("project", o.GetProject())
		if err != nil {
			return errors.Wrap(err, "failed to update project in env.yaml file")
		}
	}
	return
}

// Validate validates the DeployOptions based on completed values
func (o *DeployOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *DeployOptions) Run() error {
	devfileObj := o.EnvSpecificInfo.GetDevfileObj()
	deployHandler := newDeployHandler(devfileObj, filepath.Dir(o.EnvSpecificInfo.GetDevfilePath()), o.KClient, o.GetApplication())
	return libdevfile.Deploy(devfileObj, deployHandler)
}

// NewCmdDeploy implements the odo command
func NewCmdDeploy(name, fullName string) *cobra.Command {
	o := NewDeployOptions()
	deployCmd := &cobra.Command{
		Use:     name,
		Short:   "Deploy components",
		Long:    "Deploy the components defined in the devfile",
		Example: fmt.Sprintf(deployExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(deployCmd, clientset.INIT)

	// Add a defined annotation in order to appear in the help menu
	deployCmd.Annotations["command"] = "utility"
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return deployCmd
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

func (o *deployHandler) ApplyImage(image v1alpha2.Component) error {
	fmt.Printf("Apply image %s\n", image.Name)
	return img.BuildPushSpecificImage(o.devfileObj, o.path, image, true)
}

func (o *deployHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	fmt.Printf("Apply kubernetes %s\n", kubernetes.Name)
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
