package component

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/project"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

// CommonPushOptions has data needed for all pushes
type CommonPushOptions struct {
	ignores []string
	show    bool

	sourceType       config.SrcType
	sourcePath       string
	componentContext string
	client           *occlient.Client
	localConfigInfo  *config.LocalConfigInfo

	pushConfig  bool
	pushSource  bool
	forceBuild  bool
	isCmpExists bool

	*genericclioptions.Context
}

// NewCommonPushOptions instantiates a commonPushOptions object
func NewCommonPushOptions() *CommonPushOptions {
	return &CommonPushOptions{
		show: false,
	}
}

// ResolveSrcAndConfigFlags sets all pushes if none is asked
func (cpo *CommonPushOptions) ResolveSrcAndConfigFlags() {
	// If neither config nor source flag is passed, update both config and source to the component
	if !cpo.pushConfig && !cpo.pushSource {
		cpo.pushConfig = true
		cpo.pushSource = true
	}
}

func (cpo *CommonPushOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {
	if !cpo.pushConfig {
		// Not the case of component creation or updation(with new config)
		// So nothing to do here and hence return from here
		return nil
	}

	cmpName := cpo.localConfigInfo.GetName()

	// Output the "new" section (applying changes)
	log.Info("\nConfiguration changes")

	// If the component does not exist, we will create it for the first time.
	if !cpo.isCmpExists {

		// Classic case of component creation
		if err := component.CreateComponent(cpo.Context.Client, *cpo.localConfigInfo, cpo.componentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %+v",
				cmpName,
				err,
			)
			os.Exit(1)
		}
	}

	// Apply config
	err := component.ApplyConfig(cpo.Context.Client, *cpo.localConfigInfo, stdout, cpo.isCmpExists)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed")
	}

	return nil
}

// ResolveProject completes the push options as needed
func (cpo *CommonPushOptions) ResolveProject(prjName string) (err error) {

	// check if project exist
	isPrjExists, err := project.Exists(cpo.Context.Client, prjName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if project with name %s exists", prjName)
	}
	if !isPrjExists {
		err = project.Create(cpo.Context.Client, prjName, true)
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
	cpo.sourceType = cpo.localConfigInfo.GetSourceType()

	glog.V(4).Infof("SourceLocation: %s", cpo.localConfigInfo.GetSourceLocation())

	// Get SourceLocation here...
	cpo.sourcePath, err = cpo.localConfigInfo.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve absolute path to source location")
	}

	glog.V(4).Infof("Source Path: %s", cpo.sourcePath)
	return
}

// Push pushes changes as per set options
func (cpo *CommonPushOptions) Push() (err error) {

	stdout := color.Output

	cmpName := cpo.localConfigInfo.GetName()
	appName := cpo.localConfigInfo.GetApplication()

	if cpo.componentContext == "" {
		cpo.componentContext = strings.Trim(filepath.Dir(cpo.localConfigInfo.Filename), ".odo")
	}

	if err := cpo.createCmpIfNotExistsAndApplyCmpConfig(stdout); err != nil {
		return err
	}
	// Force the build if the component doesn't exist, we may have an out-of-date index.
	if !cpo.isCmpExists {
		cpo.forceBuild = true
	}

	if !cpo.pushSource {
		// If source is not requested for update, return
		return nil
	}

	log.Infof("\nPushing to component %s of type %s, force=%v",
		cmpName, cpo.sourceType, cpo.forceBuild)

	// Get SourceLocation here...
	cpo.sourcePath, err = cpo.localConfigInfo.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve OS source path to source location")
	}

	var (
		path  string
		files []string
	)
	switch cpo.sourceType {
	case config.LOCAL:
		glog.V(4).Infof("Copying directory %s to pod", cpo.sourcePath)
		path = cpo.sourcePath
		files = nil
	case config.BINARY:
		// We will pass in the directory, NOT filepath since this is a binary.
		glog.V(4).Infof("Copying binary file %s to pod", cpo.sourcePath)
		path = filepath.Dir(cpo.sourcePath)
		files = []string{cpo.sourcePath}
	default:
		// we don't need a case for building git components
		// the build happens before deployment
		return nil
	}
	err = component.PushLocal(
		cpo.Context.Client,
		cmpName,
		appName,
		path,
		os.Stdout,
		files, // files to copy
		nil,   // deleted files, computed by PushLocal
		cpo.forceBuild,
		cpo.ignores,
		cpo.show,
	)
	return nil
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
