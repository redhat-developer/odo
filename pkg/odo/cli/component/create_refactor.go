package component

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	scontext "github.com/openshift/odo/pkg/segment/context"

	"github.com/openshift/odo/pkg/catalog"
	registryUtil "github.com/openshift/odo/pkg/odo/cli/registry/util"
	"github.com/zalando/go-keyring"

	odoDevfile "github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/envinfo"

	"github.com/devfile/library/pkg/devfile"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (co *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	// GETTERS
	// Get context
	co.Context, err = getContext(co.now, cmd)
	if err != nil {
		return err
	}
	// Get the app name
	co.appName = genericclioptions.ResolveAppFlag(cmd)
	// Get the project name
	co.devfileMetadata.componentNamespace = co.Context.Project
	// Get DevfilePath
	co.DevfilePath = location.DevfileLocation(co.componentContext)
	//Check whether the directory already contains a devfile, this check should happen early
	co.devfileMetadata.userCreatedDevfile = util.CheckPathExists(co.DevfilePath)
	// EnvFilePath is the path of env file for devfile component
	envFilePath := getEnvFilePath(co.componentContext)
	// This is required so that .odo is created in the correct context
	co.PushOptions.componentContext = co.componentContext
	// Use Interactive mode if: 1) no args are passed || 2) the devfile exists || 3) --devfile is used
	if len(args) == 0 && !util.CheckPathExists(co.DevfilePath) && co.devfileMetadata.devfilePath.value == "" {
		co.interactive = true
	}

	// CONFLICT CHECK
	// Check if a component exists
	if util.CheckPathExists(envFilePath) && util.CheckPathExists(co.DevfilePath) {
		return errors.New("this directory already contains a component")
	}
	// Check if there is a dangling env file; delete the env file if found
	if util.CheckPathExists(envFilePath) && !util.CheckPathExists(co.DevfilePath) {
		log.Warningf("Found a dangling env file without a devfile, overwriting it")
		// Note: if the IF condition seems to have a side-effect, it is better to do the condition check separately, like below
		err = util.DeletePath(envFilePath)
		if err != nil {
			return err
		}
	}
	//Check if the directory already contains a devfile when --devfile flag is passed
	if util.CheckPathExists(co.DevfilePath) && co.devfileMetadata.devfilePath.value != "" && !util.PathEqual(co.DevfilePath, co.devfileMetadata.devfilePath.value) {
		return errors.New("this directory already contains a devfile, you can't specify devfile via --devfile")
	}
	// Check if both --devfile and --registry flag are used, in which case raise an error
	if co.devfileMetadata.devfileRegistry.Name != "" && co.devfileMetadata.devfilePath.value != "" {
		return errors.New("you can't specify registry via --registry if you want to use the devfile that is specified via --devfile")
	}
	// More than one arguments should not be allowed when a devfile exists or --devfile is used
	if len(args) > 1 && (co.devfileMetadata.userCreatedDevfile || co.devfileMetadata.devfilePath.value != "") {
		return errors.Errorf("accepts between 0 and 1 arg when using existing devfile, received %d", len(args))
	}

	// Initialize envinfo
	err = co.InitEnvInfoFromContext()
	if err != nil {
		return err
	}

	// Set the starter project token if required
	if co.devfileMetadata.starter != "" {
		var secure bool
		secure, err = registryUtil.IsSecure(co.devfileMetadata.devfileRegistry.Name)
		if err != nil {
			return err
		}
		if co.devfileMetadata.starterToken == "" && secure {
			var token string
			token, err = keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, co.devfileMetadata.devfileRegistry.Name), registryUtil.RegistryUser)
			if err != nil {
				return errors.Wrap(err, "unable to get secure registry credential from keyring")
			}
			co.devfileMetadata.starterToken = token
		}
	}

	log.Info("Devfile Object Creation")
	var catalogDevfileList catalog.DevfileComponentTypeList
	var createMethod CreateMethod
	if co.devfileMetadata.devfilePath.value != "" {
		fileErr := util.ValidateFile(co.devfileMetadata.devfilePath.value)
		urlErr := util.ValidateURL(co.devfileMetadata.devfilePath.value)
		if fileErr != nil && urlErr != nil {
			return errors.Errorf("the devfile path you specify is invalid with either file error %q or url error %q", fileErr, urlErr)
		} else if fileErr == nil {
			co.devfileMetadata.devfilePath.protocol = "file"
			createMethod = FileCreateMethod{}

		} else if urlErr == nil {
			co.devfileMetadata.devfilePath.protocol = "http(s)"
			createMethod = HTTPCreateMethod{}
		}
	}
	switch {
	case co.devfileMetadata.devfilePath.value != "" || co.devfileMetadata.userCreatedDevfile:
		//co.devfileName = "" for user provided devfile
		err = createMethod.FetchDevfileAndCreateComponent(co, catalogDevfileList)
		if err != nil {
			createMethod.Rollback(co.DevfilePath)
			return err
		}
		err = createMethod.SetMetadata(co, cmd, args, catalogDevfileList)
		if err != nil {
			createMethod.Rollback(co.DevfilePath)
			return err
		}
	default:
		if co.interactive {
			createMethod = InteractiveCreateMethod{}
		} else {
			createMethod = DirectCreateMethod{}
		}

		catalogDevfileList, err = validateAndFetchRegistry(co.devfileMetadata.devfileRegistry.Name)
		if err != nil {
			return err
		}
		err = createMethod.SetMetadata(co, cmd, args, catalogDevfileList)
		if err != nil {
			return err
		}
		err = createMethod.FetchDevfileAndCreateComponent(co, catalogDevfileList)
		if err != nil {
			createMethod.Rollback(co.DevfilePath)
			return err
		}
	}
	// Adding user provided devfile name to telemetry data
	scontext.SetDevfileName(cmd.Context(), co.devfileName)

	return nil
}

func (co *CreateOptions) Validate() (err error) {
	log.Info("Validation")
	// Validate if the devfile component name that user wants to create adheres to the k8s naming convention
	spinner := log.Spinner("Validating if devfile name is correct")
	defer spinner.End(false)

	err = util.ValidateK8sResourceName("component name", co.devfileMetadata.componentName)
	if err != nil {
		return err
	}
	spinner.End(true)

	// Validate if the devfile is compatible with odo; this checks the resolved/flattened version of devfile
	spinner = log.Spinner("Validating the devfile for odo")
	defer spinner.End(false)

	_, err = odoDevfile.ParseAndValidateFromFile(co.DevfilePath)
	if err != nil {
		return err
	}
	spinner.End(true)

	return nil
}

func (co *CreateOptions) Run(cmd *cobra.Command) (err error) {
	// Adding component type to telemetry data
	scontext.SetComponentType(cmd.Context(), co.devfileMetadata.componentType)

	devObj, err := devfileParseFromFile(co.DevfilePath, false)
	if err != nil {
		return errors.New("Failed to parse the devfile")
	}

	devfileData, err := ioutil.ReadFile(co.DevfilePath)
	if err != nil {
		return err
	}
	// WARN: Starter Project uses go-git that overrides the directory content, there by deleting the existing devfile.
	err = decideAndDownloadStarterProject(devObj, co.devfileMetadata.starter, co.devfileMetadata.starterToken, co.interactive, co.componentContext)
	if err != nil {
		return errors.Wrap(err, "failed to download project for devfile component")
	}

	// TODO: We should not have to rewrite to the file. Fix the starter project.
	err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644) // #nosec G306
	if err != nil {
		return err
	}

	// If user provided a custom name, re-write the devfile
	// ENSURE: co.devfileMetadata.componentName != ""
	if co.devfileMetadata.componentName != devObj.GetMetadataName() {
		spinner := log.Spinnerf("Updating the devfile with component name %q", co.devfileMetadata.componentName)
		defer spinner.End(false)

		err = devObj.SetMetadataName(co.devfileMetadata.componentName)
		if err != nil {
			return errors.New("Failed to update the devfile")
		}
		spinner.End(true)
	}

	// Generate env file
	err = co.EnvSpecificInfo.SetComponentSettings(envinfo.ComponentSettings{
		Name:               co.devfileMetadata.componentName,
		Project:            co.devfileMetadata.componentNamespace,
		AppName:            co.appName,
		UserCreatedDevfile: co.devfileMetadata.userCreatedDevfile,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create env file for devfile component")
	}

	sourcePath, err := util.GetAbsPath(co.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	ignoreFile, err := util.TouchGitIgnoreFile(sourcePath)
	if err != nil {
		return err
	}

	err = util.AddFileToIgnoreFile(ignoreFile, filepath.Join(co.componentContext, EnvDirectory))
	if err != nil {
		return err
	}

	if co.now {
		err = co.DevfilePush()
		if err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}
	} else {
		log.Italic("\nPlease use `odo push` command to create the component with source deployed")
	}

	if log.IsJSON() {
		return co.DevfileJSON()
	}

	return nil
}

func getContext(now bool, cmd *cobra.Command) (*genericclioptions.Context, error) {
	if now {
		return genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	}
	return genericclioptions.NewOfflineContext(cmd)
}

func getEnvFilePath(componentContext string) string {
	if componentContext != "" {
		return filepath.Join(componentContext, EnvYAMLFilePath)
	}
	return filepath.Join(LocalDirectoryDefaultLocation, EnvYAMLFilePath)
}

// DevfileParseFromFile reads, parses and validates a devfile from a file without flattening it
func devfileParseFromFile(devfilePath string, resolved bool) (parser.DevfileObj, error) {
	devObj, _, err := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: &resolved})
	if err != nil {
		return parser.DevfileObj{}, err
	}

	return devObj, nil
}
