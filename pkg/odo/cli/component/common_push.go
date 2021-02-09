package component

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/project"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// CommonPushOptions has data needed for all pushes
type CommonPushOptions struct {
	ignores []string
	show    bool

	sourceType       config.SrcType
	sourcePath       string
	componentContext string

	EnvSpecificInfo *envinfo.EnvSpecificInfo

	pushConfig         bool
	pushSource         bool
	forceBuild         bool
	doesComponentExist bool

	*genericclioptions.Context
}

// NewCommonPushOptions instantiates a commonPushOptions object
func NewCommonPushOptions() *CommonPushOptions {
	return &CommonPushOptions{
		show: false,
	}
}

//InitConfigFromContext initializes localconfiginfo from the context
func (cpo *CommonPushOptions) InitConfigFromContext() error {
	var err error
	cpo.LocalConfigInfo, err = config.NewLocalConfigInfo(cpo.componentContext)
	if err != nil {
		return err
	}
	return nil
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

//ValidateComponentCreate validates if the request to create component is valid
func (cpo *CommonPushOptions) ValidateComponentCreate() error {
	var err error
	s := log.Spinner("Checking component")
	defer s.End(false)
	cpo.doesComponentExist, err = component.Exists(cpo.Context.Client, cpo.LocalConfigInfo.GetName(), cpo.LocalConfigInfo.GetApplication())
	if err != nil {
		return errors.Wrapf(err, "failed to check if component of name %s exists in application %s", cpo.LocalConfigInfo.GetName(), cpo.LocalConfigInfo.GetApplication())
	}

	if err = component.ValidateComponentCreateRequest(cpo.Context.Client, cpo.LocalConfigInfo.GetComponentSettings(), cpo.componentContext); err != nil {
		s.End(false)
		log.Italic("\nRun 'odo catalog list components' for a list of supported component types")
		return fmt.Errorf("Invalid component type %s, %v", *cpo.LocalConfigInfo.GetComponentSettings().Type, errors.Cause(err))
	}
	s.End(true)
	return nil
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

// SetSourceInfo sets up source information
func (cpo *CommonPushOptions) SetSourceInfo() (err error) {
	cpo.sourceType = cpo.LocalConfigInfo.GetSourceType()

	klog.V(4).Infof("SourceLocation: %s", cpo.LocalConfigInfo.GetSourceLocation())

	// Get SourceLocation here...
	cpo.sourcePath, err = cpo.LocalConfigInfo.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve absolute path to source location")
	}

	klog.V(4).Infof("Source Path: %s", cpo.sourcePath)
	return
}

// filterIgnores applies the glob rules on the filesChanged and filesDeleted and filters them
// returns the filtered results which match any of the glob rules
func filterIgnores(filesChanged, filesDeleted, absIgnoreRules []string) (filesChangedFiltered, filesDeletedFiltered []string) {
	for _, file := range filesChanged {
		match, err := util.IsGlobExpMatch(file, absIgnoreRules)
		if err != nil {
			continue
		}
		if !match {
			filesChangedFiltered = append(filesChangedFiltered, file)
		}
	}

	for _, file := range filesDeleted {
		match, err := util.IsGlobExpMatch(file, absIgnoreRules)
		if err != nil {
			continue
		}
		if !match {
			filesDeletedFiltered = append(filesDeletedFiltered, file)
		}
	}
	return filesChangedFiltered, filesDeletedFiltered
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
func retrieveCmdNamespace(cmd *cobra.Command) (string, error) {
	var componentNamespace string
	var err error

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

// gatherName parses the Devfile retrieves an appropriate name in two ways.
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
