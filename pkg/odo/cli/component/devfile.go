package component

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	dfutil "github.com/devfile/library/pkg/util"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/service"
)

// DevfilePush has the logic to perform the required actions for a given devfile
func (po *PushOptions) DevfilePush() error {

	// Wrap the push so that we can capture the error in JSON-only mode
	err := po.devfilePushInner()

	if err != nil && log.IsJSON() {
		eventLoggingClient := machineoutput.NewConsoleMachineEventLoggingClient()
		eventLoggingClient.ReportError(err, machineoutput.TimestampNow())

		// Suppress the error to prevent it from being output by the generic machine-readable handler (which will produce invalid JSON for our purposes)
		err = nil

		// os.Exit(1) since we are suppressing the generic machine-readable handler's exit code logic
		os.Exit(1)
	}

	if err != nil {
		return err
	}

	// push is successful, save the runMode used
	runMode := envinfo.Run
	if po.debugFlag {
		runMode = envinfo.Debug
	}

	return po.EnvSpecificInfo.SetRunMode(runMode)
}

func (po *PushOptions) devfilePushInner() (err error) {
	devObj, err := devfile.ParseAndValidateFromFile(po.DevfilePath)
	if err != nil {
		return err
	}
	componentName := po.EnvSpecificInfo.GetName()

	// Set the source path to either the context or current working directory (if context not set)
	po.sourcePath, err = dfutil.GetAbsPath(po.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&po.ignoreFlag, po.sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	var platformContext interface{}
	kc := kubernetes.KubernetesContext{
		Namespace: po.KClient.GetCurrentNamespace(),
	}
	platformContext = kc

	devfileHandler, err := adapters.NewComponentAdapter(componentName, po.sourcePath, po.GetApplication(), devObj, platformContext)
	if err != nil {
		return err
	}

	pushParams := common.PushParameters{
		Path:            po.sourcePath,
		IgnoredFiles:    po.ignoreFlag,
		ForceBuild:      po.forceBuildFlag,
		Show:            po.showFlag,
		EnvSpecificInfo: *po.EnvSpecificInfo,
		DevfileBuildCmd: strings.ToLower(po.buildCommandFlag),
		DevfileRunCmd:   strings.ToLower(po.runCommandflag),
		DevfileDebugCmd: strings.ToLower(po.debugCommandFlag),
		Debug:           po.debugFlag,
		DebugPort:       po.EnvSpecificInfo.GetDebugPort(),
	}

	_, err = po.EnvSpecificInfo.ListURLs()
	if err != nil {
		return err
	}

	// Start or update the component
	err = devfileHandler.Push(pushParams)
	if err != nil {
		err = errors.Errorf("Failed to start component with name %q. Error: %v",
			componentName,
			err,
		)
	} else {
		log.Infof("\nPushing devfile component %q", componentName)
		log.Success("Changes successfully pushed to component")
	}

	return
}

// DevfileUnDeploy undeploys the devfile kubernetes components
func (do *DeleteOptions) DevfileUnDeploy() error {
	devfileObj := do.EnvSpecificInfo.GetDevfileObj()
	undeployHandler := newUndeployHandler(filepath.Dir(do.EnvSpecificInfo.GetDevfilePath()), do.KClient)
	return libdevfile.Deploy(devfileObj, undeployHandler)
}

// DevfileComponentDelete deletes the devfile component
func (do *DeleteOptions) DevfileComponentDelete() error {
	devObj, err := devfile.ParseAndValidateFromFile(do.GetDevfilePath())
	if err != nil {
		return err
	}

	componentName := do.EnvSpecificInfo.GetName()

	kc := kubernetes.KubernetesContext{
		Namespace: do.KClient.GetCurrentNamespace(),
	}

	labels := componentlabels.GetLabels(componentName, do.EnvSpecificInfo.GetApplication(), false)
	devfileHandler, err := adapters.NewComponentAdapter(componentName, do.contextFlag, do.GetApplication(), devObj, kc)
	if err != nil {
		return err
	}

	return devfileHandler.Delete(labels, do.showLogFlag, do.waitFlag)
}

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
	u, err := service.GetK8sComponentAsUnstructured(kubernetes.Kubernetes, o.path, devfilefs.DefaultFs{})
	if err != nil {
		return err
	}

	// Get the REST mappings
	gvr, err := o.kubeClient.GetRestMappingFromUnstructured(u)
	if err != nil {
		return err
	}
	log.Printf("Un-deploying the Kubernetes %s: %s", u.GetKind(), u.GetName())
	// Un-deploy the K8s manifest
	return o.kubeClient.DeleteDynamicResource(u.GetName(), gvr.Resource.Group, gvr.Resource.Version, gvr.Resource.Resource)
}
