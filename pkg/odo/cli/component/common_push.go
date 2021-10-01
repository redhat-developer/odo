package component

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/fatih/color"
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

func (cpo *CommonPushOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {
	if !cpo.pushConfig {
		// Not the case of component creation or update (with new config)
		// So nothing to do here and hence return from here
		return nil
	}

	cmpName := cpo.LocalConfigInfo.GetName()

	// Output the "new" section (applying changes)
	log.Info("\nConfiguration changes")

	// If the component does not exist, we will create it for the first time.
	if !cpo.doesComponentExist {

		// Classic case of component creation
		if err := component.CreateComponent(cpo.Context.Client, *cpo.LocalConfigInfo, cpo.componentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %v",
				cmpName,
				err,
			)
			return err
		}
	}
	// Apply config
	err := component.ApplyConfig(cpo.Context.Client, *cpo.LocalConfigInfo, envinfo.EnvSpecificInfo{}, stdout, cpo.doesComponentExist, true)
	if err != nil {
		return errors.Wrapf(err, "Failed to update config to component deployed.")
	}

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

// Push pushes changes as per set options
func (cpo *CommonPushOptions) Push() (err error) {

	deletedFiles := []string{}
	changedFiles := []string{}
	isForcePush := false

	stdout := color.Output
	// Ret from Indexer function
	var ret util.IndexerRet

	cmpName := cpo.LocalConfigInfo.GetName()
	appName := cpo.LocalConfigInfo.GetApplication()
	cpo.sourceType = cpo.LocalConfigInfo.GetSourceType()
	// force write the content to resolvePath
	forceWrite := false

	if cpo.componentContext == "" {
		cpo.componentContext = strings.TrimSuffix(filepath.Dir(cpo.LocalConfigInfo.Filename), ".odo")
	}

	err = cpo.createCmpIfNotExistsAndApplyCmpConfig(stdout)
	if err != nil {
		return
	}

	if !cpo.pushSource {
		// If source is not requested for update, return
		return nil
	}

	log.Infof("\nPushing to component %s of type %s", cmpName, cpo.sourceType)

	if !cpo.forceBuild && cpo.sourceType != config.GIT {
		absIgnoreRules := util.GetAbsGlobExps(cpo.sourcePath, cpo.ignores)

		spinner := log.NewStatus(log.GetStdout())
		defer spinner.End(true)
		if cpo.doesComponentExist {
			spinner.Start("Checking file changes for pushing", false)
		} else {
			// if the component doesn't exist, we don't check for changes in the files
			// thus we show a different message
			spinner.Start("Checking files for pushing", false)
		}

		// run the indexer and find the modified/added/deleted/renamed files
		ret, err = util.RunIndexer(cpo.componentContext, absIgnoreRules)
		spinner.End(true)

		if err != nil {
			return errors.Wrap(err, "unable to run indexer")
		}
		if len(ret.FilesChanged) > 0 || len(ret.FilesDeleted) > 0 {
			forceWrite = true
		}

		if cpo.doesComponentExist {
			// apply the glob rules from the .gitignore/.odoignore file
			// and ignore the files on which the rules apply and filter them out
			filesChangedFiltered, filesDeletedFiltered := filterIgnores(ret.FilesChanged, ret.FilesDeleted, absIgnoreRules)

			// Remove the relative file directory from the list of deleted files
			// in order to make the changes correctly within the OpenShift pod
			deletedFiles, err = util.RemoveRelativePathFromFiles(filesDeletedFiltered, cpo.sourcePath)
			if err != nil {
				return errors.Wrap(err, "unable to remove relative path from list of changed/deleted files")
			}
			klog.V(4).Infof("List of files to be deleted: +%v", deletedFiles)
			changedFiles = filesChangedFiltered

			if len(filesChangedFiltered) == 0 && len(filesDeletedFiltered) == 0 {
				// no file was modified/added/deleted/renamed, thus return without building
				log.Success("No file changes detected, skipping build. Use the '-f' flag to force the build.")
				return nil
			}
		}
	}

	if cpo.forceBuild || !cpo.doesComponentExist {
		isForcePush = true
	}

	// Get SourceLocation here...
	cpo.sourcePath, err = cpo.LocalConfigInfo.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve OS source path to source location")
	}

	switch cpo.sourceType {
	case config.LOCAL:
		klog.V(4).Infof("Copying directory %s to pod", cpo.sourcePath)
		err = component.PushLocal(
			cpo.Context.Client,
			cmpName,
			appName,
			cpo.sourcePath,
			os.Stdout,
			changedFiles,
			deletedFiles,
			isForcePush,
			util.GetAbsGlobExps(cpo.sourcePath, cpo.ignores),
			cpo.show,
		)

		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case config.BINARY:

		// We will pass in the directory, NOT filepath since this is a binary..
		binaryDirectory := filepath.Dir(cpo.sourcePath)

		klog.V(4).Infof("Copying binary file %s to pod", cpo.sourcePath)
		err = component.PushLocal(
			cpo.Context.Client,
			cmpName,
			appName,
			binaryDirectory,
			os.Stdout,
			[]string{cpo.sourcePath},
			deletedFiles,
			isForcePush,
			util.GetAbsGlobExps(cpo.sourcePath, cpo.ignores),
			cpo.show,
		)

		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

		// we don't need a case for building git components
		// the build happens before deployment

		return errors.Wrapf(err, fmt.Sprintf("failed to push component: %v", cmpName))
	}
	if forceWrite {
		err = util.WriteFile(ret.NewFileMap, ret.ResolvedPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to write file")
		}
	}

	log.Success("Changes successfully pushed to component")
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
