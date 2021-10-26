package component

import (
	"path/filepath"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/v2/pkg/envinfo"
	"github.com/openshift/odo/v2/pkg/kclient"
	"github.com/openshift/odo/v2/pkg/odo/genericclioptions"
	"github.com/openshift/odo/v2/pkg/project"
	"github.com/openshift/odo/v2/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// CommonPushOptions has data needed for all pushes
type CommonPushOptions struct {
	show             bool
	componentContext string
	pushConfig       bool
	pushSource       bool
	EnvSpecificInfo  *envinfo.EnvSpecificInfo
	*genericclioptions.Context
}

// NewCommonPushOptions instantiates a commonPushOptions object
func NewCommonPushOptions() *CommonPushOptions {
	return &CommonPushOptions{
		show: false,
	}
}

//InitEnvInfoFromContext initializes envinfo from the context
func (cpo *CommonPushOptions) InitEnvInfoFromContext() (err error) {
	cpo.EnvSpecificInfo, err = envinfo.NewEnvSpecificInfo(cpo.componentContext)
	if err != nil {
		return err
	}
	return nil
}

//AddContextFlag adds the context flag to specified command storing value of flag in options.componentContext
func (cpo *CommonPushOptions) AddContextFlag(cmd *cobra.Command) {
	genericclioptions.AddContextFlag(cmd, &cpo.componentContext)
}

// ResolveSrcAndConfigFlags sets all pushes if none is asked
func (cpo *CommonPushOptions) ResolveSrcAndConfigFlags() {
	// If neither config nor source flag is passed, update both config and source to the component
	if !cpo.pushConfig && !cpo.pushSource {
		cpo.pushConfig = true
		cpo.pushSource = true
	}
}

// ResolveProject completes the push options as needed
func (cpo *CommonPushOptions) ResolveProject(prjName string) (err error) {

	// check if project exist
	isPrjExists, err := project.Exists(cpo.Context, prjName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if project with name %s exists", prjName)
	}
	if !isPrjExists {
		err = project.Create(cpo.Context, prjName, true)
		if err != nil {
			return errors.Wrapf(
				err,
				"project %s does not exist. Failed creating it. Please try after creating project using `odo project create <project_name>`",
				prjName,
			)
		}
	}
	cpo.Context.Client.Namespace = prjName
	return
}

// retrieveKubernetesDefaultNamespace tries to retrieve the current active namespace
// to set as a default namespace
func retrieveKubernetesDefaultNamespace() (string, error) {
	// Get current active namespace
	client, err := kclient.New()
	if err != nil {
		return "", err
	}
	return client.Namespace, nil
}

// retrieveCmdNamespace retrieves the namespace from project flag, if unset
// we revert to the default namespace available from Kubernetes
func retrieveCmdNamespace(cmd *cobra.Command) (componentNamespace string, err error) {
	// For "odo create" check to see if --project has been passed.
	if cmd.Flags().Changed("project") {
		componentNamespace, err = cmd.Flags().GetString("project")
		if err != nil {
			return "", err
		}
	} else {
		componentNamespace, err = retrieveKubernetesDefaultNamespace()
		if err != nil {
			return "", err
		}
	}

	return componentNamespace, nil
}

// gatherName parses the Devfile and retrieves an appropriate name in two ways.
// 1. If metadata.name exists, we use it
// 2. If metadata.name does NOT exist, we use the folder name where the devfile.yaml is located
func gatherName(devObj parser.DevfileObj, devfilePath string) (string, error) {

	metadata := devObj.Data.GetMetadata()

	klog.V(4).Infof("metadata.Name: %s", metadata.Name)

	// 1. Use metadata.name if it exists
	if metadata.Name != "" {

		// Remove any suffix's that end with `-`. This is because many Devfile's use the original v1 Devfile pattern of
		// having names such as "foo-bar-" in order to prepend container names such as "foo-bar-container1"
		return strings.TrimSuffix(metadata.Name, "-"), nil
	}

	// 2. Use the folder name as a last resort if nothing else exists
	sourcePath, err := util.GetAbsPath(devfilePath)
	if err != nil {
		return "", errors.Wrap(err, "unable to get source path")
	}
	klog.V(4).Infof("Source path: %s", sourcePath)
	klog.V(4).Infof("devfile dir: %s", filepath.Dir(sourcePath))

	return filepath.Base(filepath.Dir(sourcePath)), nil
}
