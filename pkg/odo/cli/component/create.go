package component

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/registry/util"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/zalando/go-keyring"

	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/catalog"
	odoDevfile "github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// CreateOptions encapsulates create options
type CreateOptions struct {
	// Push context
	*PushOptions

	// Flags
	contextFlag string
	portFlag    []string
	envFlag     []string
	nowFlag     bool
	appFlag     string

	interactive bool

	// devfileName stores the "componentType" passed by user irrespective of it being a valid componentType
	// we use it for telemetry
	devfileName string

	createMethod    CreateMethod
	devfileMetadata DevfileMetadata
}

// Path of user's own devfile, user specifies the path via --devfile flag
type devfilePath struct {
	protocol string
	value    string
}

// DevfileMetadata includes devfile component metadata
type DevfileMetadata struct {
	componentType      string
	componentName      string
	componentNamespace string
	devfileLink        string
	devfileRegistry    catalog.Registry
	devfilePath        devfilePath
	userCreatedDevfile bool
	starter            string
	token              string
	starterToken       string
}

// CreateRecommendedCommandName is the recommended watch command name
const CreateRecommendedCommandName = "create"

// LocalDirectoryDefaultLocation is the default location of where --local files should always be..
// since the application will always be in the same directory as `.odo`, we will always set this as: ./
const LocalDirectoryDefaultLocation = "./"

var (
	EnvYAMLFilePath = filepath.Join(".odo", "env", "env.yaml")
	EnvDirectory    = filepath.Join(".odo", "env")
)

var createLongDesc = ktemplates.LongDesc(`Create a configuration describing a component.`)

var createExample = ktemplates.Examples(`# Create a new Node.JS component with existing sourcecode as well as specifying a name
%[1]s nodejs mynodejs

# Name is not required and will be automatically generated if not passed
%[1]s nodejs

# List all available components before deploying
odo catalog list components
%[1]s java-quarkus

# Download an example devfile and application before deploying
%[1]s nodejs --starter

# Using a specific devfile
%[1]s mynodejs --devfile ./devfile.yaml
%[1]s mynodejs --devfile https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml

# Create new Node.js component named 'frontend' with the source in './frontend' directory
%[1]s nodejs frontend --context ./frontend

# Create new Node.js component with custom ports and environment variables
%[1]s nodejs --port 8080,8100/tcp,9100/udp --env key=value,key1=value1

# Create a new Node.js component that is a part of 'myapp' app inside the 'myproject' project 
%[1]s nodejs --app myapp --project myproject`)

// NewCreateOptions returns new instance of CreateOptions
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		PushOptions: NewPushOptions(),
	}
}

func (co *CreateOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	// GETTERS
	// Get context
	co.Context, err = getContext(co.nowFlag, cmdline)
	if err != nil {
		return err
	}
	// Get the app name
	co.appFlag = genericclioptions.ResolveAppFlag(cmdline)
	// Get the project name
	co.devfileMetadata.componentNamespace = co.Context.GetProject()
	// Get DevfilePath
	co.DevfilePath = location.DevfileLocation(co.contextFlag)
	//Check whether the directory already contains a devfile, this check should happen early
	co.devfileMetadata.userCreatedDevfile = util.CheckPathExists(co.DevfilePath)
	// EnvFilePath is the path of env file for devfile component
	envFilePath := getEnvFilePath(co.contextFlag)
	// This is required so that .odo is created in the correct context
	co.PushOptions.componentContext = co.contextFlag
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

	// Initialize envinfo
	err = co.InitEnvInfoFromContext()
	if err != nil {
		return err
	}

	// Fetch the necessary devfile and create the component
	log.Info("Devfile Object Creation")
	switch {
	case co.devfileMetadata.userCreatedDevfile:
		co.createMethod = UserCreatedDevfileMethod{}
	case co.devfileMetadata.devfilePath.value != "":
		//co.devfileName = "" for user provided devfile
		fileErr := util.ValidateFile(co.devfileMetadata.devfilePath.value)
		urlErr := util.ValidateURL(co.devfileMetadata.devfilePath.value)
		if fileErr != nil && urlErr != nil {
			return errors.Errorf("the devfile path you specify is invalid with either file error %q or url error %q", fileErr, urlErr)
		} else if fileErr == nil {
			co.createMethod = FileCreateMethod{}
		} else if urlErr == nil {
			co.createMethod = HTTPCreateMethod{}
		}
	case co.interactive:
		co.createMethod = InteractiveCreateMethod{}
	default:
		co.createMethod = DirectCreateMethod{}
	}
	err = co.createMethod.CheckConflicts(co, args)
	if err != nil {
		return err
	}
	err = co.createMethod.FetchDevfileAndCreateComponent(co, cmdline, args)
	if err != nil {
		co.createMethod.Rollback(co.DevfilePath, co.contextFlag)
		return err
	}

	// From this point forward, rollback should be triggered if an error is encountered; rollback should delete all the files that were created by odo
	defer func() {
		if err != nil {
			co.createMethod.Rollback(co.DevfilePath, co.contextFlag)
		}
	}()
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

	scontext.SetDevfileName(cmdline.Context(), co.devfileName)
	// Adding component type to telemetry data
	scontext.SetComponentType(cmdline.Context(), co.devfileMetadata.componentType)

	return nil
}

func (co *CreateOptions) Validate() (err error) {
	defer func() {
		if err != nil {
			co.createMethod.Rollback(co.DevfilePath, co.contextFlag)
		}
	}()
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

func (co *CreateOptions) Run() (err error) {
	defer func() {
		if err != nil {
			co.createMethod.Rollback(co.DevfilePath, co.contextFlag)
		}
	}()

	devObj, err := devfileParseFromFile(co.DevfilePath, false)
	if err != nil {
		return errors.New("Failed to parse the devfile")
	}

	devfileData, err := ioutil.ReadFile(co.DevfilePath)
	if err != nil {
		return err
	}
	// WARN: Starter Project uses go-git that overrides the directory content, there by deleting the existing devfile.
	err = decideAndDownloadStarterProject(devObj, co.devfileMetadata.starter, co.devfileMetadata.starterToken, co.interactive, co.contextFlag)
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

		// WARN: SetMetadataName will rewrite to the devfile
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
		AppName:            co.appFlag,
		UserCreatedDevfile: co.devfileMetadata.userCreatedDevfile,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create env file for devfile component")
	}

	// Prepare .gitignore file
	sourcePath, err := util.GetAbsPath(co.contextFlag)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	ignoreFile, err := util.TouchGitIgnoreFile(sourcePath)
	if err != nil {
		return err
	}

	err = util.AddFileToIgnoreFile(ignoreFile, filepath.Join(co.contextFlag, EnvDirectory))
	if err != nil {
		return err
	}

	if co.nowFlag {
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

// NewCmdCreate implements the create odo command
func NewCmdCreate(name, fullName string) *cobra.Command {
	co := NewCreateOptions()
	var componentCreateCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s <component_type> [component_name] [flags]", name),
		Short:       "Create a new component",
		Long:        createLongDesc,
		Example:     fmt.Sprintf(createExample, fullName),
		Args:        cobra.RangeArgs(0, 2),
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(co, cmd, args)
		},
	}
	odoutil.AddContextFlag(componentCreateCmd, &co.contextFlag)
	componentCreateCmd.Flags().StringSliceVarP(&co.portFlag, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp)")
	componentCreateCmd.Flags().StringSliceVar(&co.envFlag, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")

	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.starter, "starter", "", "Download a project specified in the devfile")
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.devfileRegistry.Name, "registry", "", "Create devfile component from specific registry")
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.devfilePath.value, "devfile", "", "Path to the user specified devfile")
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.token, "token", "", "Token to be used when downloading devfile from the devfile path that is specified via --devfile")
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.starterToken, "starter-token", "", "Token to be used when downloading starter project")
	componentCreateCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		if strings.Contains(err.Error(), "flag needs an argument: --starter") {
			return fmt.Errorf("%w: you can get the list of possible values with the command `odo catalog describe component <type>`", err)
		}
		return err
	})

	componentCreateCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	// Adding `--now` flag
	odoutil.AddNowFlag(componentCreateCmd, &co.nowFlag)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentCreateCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentCreateCmd)

	completion.RegisterCommandHandler(componentCreateCmd, completion.CreateCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "context", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "binary", completion.FileCompletionHandler)

	return componentCreateCmd
}

func getContext(now bool, cmdline cmdline.Cmdline) (*genericclioptions.Context, error) {
	params := genericclioptions.NewCreateParameters(cmdline)
	if now {
		params = params.CreateAppIfNeeded()
	} else {
		params = params.IsOffline()
	}
	return genericclioptions.New(params)
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
