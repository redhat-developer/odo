package component

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

// CommonPushOptions has data needed for all pushes
type commonPushOptions struct {
	ignores []string
	show    bool

	sourceType       config.SrcType
	sourcePath       string
	componentContext string
	client           *occlient.Client
	localConfig      *config.LocalConfigInfo

	pushConfig bool
	pushSource bool

	*genericclioptions.Context
}

// NewCommonPushOptions instantiates a commonPushOptions object
func NewCommonPushOptions() *commonPushOptions {
	return &commonPushOptions{
		show: false,
	}
}

func (cpo *commonPushOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {
	if !cpo.pushConfig {
		// Not the case of component creation or updation(with new config)
		// So nothing to do here and hence return from here
		return nil
	}

	cmpName := cpo.localConfig.GetName()
	appName := cpo.localConfig.GetApplication()
	cmpType := cpo.localConfig.GetType()

	isCmpExists, err := component.Exists(cpo.Context.Client, cmpName, appName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if component %s exists or not", cmpName)
	}

	if !isCmpExists {
		log.Successf("Creating %s component with name %s", cmpType, cmpName)
		// Classic case of component creation
		if err = component.CreateComponent(cpo.Context.Client, *cpo.localConfig, cpo.componentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %+v",
				cmpName,
				err,
			)
			os.Exit(1)
		}
		log.Successf("Successfully created component %s", cmpName)
	}

	// Apply config
	err = component.ApplyConfig(cpo.Context.Client, *cpo.localConfig, stdout, isCmpExists)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed")
	}
	log.Successf("Successfully updated component with name: %v", cmpName)

	return nil
}

// Push pushes changes as per set options
func (cpo *commonPushOptions) Push() (err error) {
	stdout := color.Output
	cmpName := cpo.localConfig.GetName()
	appName := cpo.localConfig.GetApplication()

	err = cpo.createCmpIfNotExistsAndApplyCmpConfig(stdout)
	if err != nil {
		return
	}

	if !cpo.pushSource {
		// If source is not requested for update, return
		return nil
	}

	log.Successf("Pushing changes to component: %v of type %s", cmpName, cpo.sourceType)

	// Get SourceLocation here...
	cpo.sourcePath, err = cpo.localConfig.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve OS source path to source location")
	}

	switch cpo.sourceType {
	case config.LOCAL:
		glog.V(4).Infof("Copying directory %s to pod", cpo.sourcePath)
		err = component.PushLocal(
			cpo.Context.Client,
			cmpName,
			appName,
			cpo.sourcePath,
			os.Stdout,
			[]string{},
			[]string{},
			true,
			util.GetAbsGlobExps(cpo.sourcePath, cpo.ignores),
			cpo.show,
		)

		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case config.BINARY:

		// We will pass in the directory, NOT filepath since this is a binary..
		binaryDirectory := filepath.Dir(cpo.sourcePath)

		glog.V(4).Infof("Copying binary file %s to pod", cpo.sourcePath)
		err = component.PushLocal(
			cpo.Context.Client,
			cmpName,
			appName,
			binaryDirectory,
			os.Stdout,
			[]string{cpo.sourcePath},
			[]string{},
			true,
			util.GetAbsGlobExps(cpo.sourcePath, cpo.ignores),
			cpo.show,
		)

		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case config.GIT:
		err := component.Build(
			cpo.Context.Client,
			cmpName,
			appName,
			true,
			stdout,
			cpo.show,
		)
		return errors.Wrapf(err, fmt.Sprintf("failed to push component: %v", cmpName))
	}

	log.Successf("Changes successfully pushed to component: %v", cmpName)

	return
}
