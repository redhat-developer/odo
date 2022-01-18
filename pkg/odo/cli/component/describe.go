package component

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/kclient"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"github.com/redhat-developer/odo/pkg/component"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
)

// DescribeRecommendedCommandName is the recommended describe command name
const DescribeRecommendedCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe nodejs component
%[1]s nodejs
`)

// DescribeOptions is a dummy container to attach complete, validate and run pattern
type DescribeOptions struct {
	// Component context
	*ComponentOptions
	// Clients
	componentClient component.Client
	// Flags
	contextFlag string
}

// NewDescribeOptions returns new instance of ListOptions
func NewDescribeOptions(compClient component.Client) *DescribeOptions {
	return &DescribeOptions{
		ComponentOptions: &ComponentOptions{},
		componentClient:  compClient,
	}
}

// Complete completes describe args
func (do *DescribeOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	if do.contextFlag == "" {
		do.contextFlag, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	err = do.ComponentOptions.Complete(cmdline, args)
	if err != nil {
		return err
	}
	return nil
}

// Validate validates the describe parameters
func (do *DescribeOptions) Validate() (err error) {

	if !((do.GetApplication() != "" && do.GetProject() != "") || do.EnvSpecificInfo.Exists()) {
		return fmt.Errorf("component %v does not exist", do.componentName)
	}

	return nil
}

// Run has the logic to perform the required actions as part of command
func (do *DescribeOptions) Run() (err error) {

	cfd, err := do.componentClient.GetComponentFullDescription(do.EnvSpecificInfo, do.componentName, do.Context.GetApplication(), do.Context.GetProject(), do.contextFlag)
	if err != nil {
		return err
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(cfd)
	} else {
		err = humanReadableDescribeOutput(cfd, do.componentClient)
		if err != nil {
			return err
		}
	}
	return
}

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe(name, fullName string) *cobra.Command {
	kubeclient, _ := kclient.New()
	do := NewDescribeOptions(component.NewClient(kubeclient))

	var describeCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component_name]", name),
		Short:       "Describe component",
		Long:        `Describe component.`,
		Example:     fmt.Sprintf(describeExample, fullName),
		Args:        cobra.RangeArgs(0, 1),
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(do, cmd, args)
		},
	}

	describeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(describeCmd, completion.ComponentNameCompletionHandler)
	// Adding --context flag
	odoutil.AddContextFlag(describeCmd, &do.contextFlag)

	// Adding `--project` flag
	projectCmd.AddProjectFlag(describeCmd)
	// Adding `--application` flag
	appCmd.AddApplicationFlag(describeCmd)

	return describeCmd
}

// Print prints the complete information of component onto stdout (Note: long term this function should not need to access any parameters, but just print the information in struct)
func humanReadableDescribeOutput(cfd *component.Component, compClient component.Client) error {
	log.Describef("Component Name: ", cfd.GetName())
	log.Describef("Type: ", cfd.Spec.Type)

	// Env
	if cfd.Spec.Env != nil {

		// Retrieve all the environment variables
		var output string
		for _, env := range cfd.Spec.Env {
			output += fmt.Sprintf(" · %v=%v\n", env.Name, env.Value)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			log.Describef("Environment Variables:\n", output[:len(output)-1])
		}

	}

	// Storage
	if len(cfd.Spec.StorageSpec) > 0 {

		// Gather the output
		var output string
		for _, store := range cfd.Spec.StorageSpec {
			var eph string
			if store.Spec.Ephemeral != nil {
				if *store.Spec.Ephemeral {
					eph = " as ephemeral volume"
				} else {
					eph = " as persistent volume"
				}
			}
			output += fmt.Sprintf(" · %v of size %v mounted to %v%s\n", store.Name, store.Spec.Size, store.Spec.Path, eph)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			log.Describef("Storage:\n", output[:len(output)-1])
		}

	}

	// URL
	if len(cfd.Spec.URLSpec) > 0 {
		var output string
		// if the component is not pushed
		for _, componentURL := range cfd.Spec.URLSpec {
			if componentURL.Status.State == urlpkg.StateTypePushed {
				output += fmt.Sprintf(" · %v exposed via %v\n", urlpkg.GetURLString(componentURL.Spec.Protocol, componentURL.Spec.Host, ""), componentURL.Spec.Port)
			} else {
				output += fmt.Sprintf(" · URL named %s will be exposed via %v\n", componentURL.Name, componentURL.Spec.Port)
			}
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			log.Describef("URLs:\n", output[:len(output)-1])
		}
	}

	// Linked services
	if len(cfd.Status.LinkedServices) > 0 {

		// Gather the output
		var output string
		for _, linkedService := range cfd.Status.LinkedServices {

			if linkedService.SecretName == "" {
				output += fmt.Sprintf(" · %s\n", linkedService.ServiceName)
				continue
			}
			// Let's also get the secrets / environment variables that are being passed in. (if there are any)
			//  FIXME: This data is not available via JSON
			secretsData, err := compClient.GetLinkedServicesSecretData(cfd.GetNamespace(), linkedService.SecretName)
			if err != nil {
				return err
			}

			if len(secretsData) > 0 {
				// Iterate through the secrets to throw in a string
				var secretOutput string
				for i := range secretsData {
					if linkedService.MountVolume {
						secretOutput += fmt.Sprintf("    · %v\n", filepath.ToSlash(filepath.Join(linkedService.MountPath, i)))
					} else {
						secretOutput += fmt.Sprintf("    · %v\n", i)
					}
				}

				if len(secretOutput) > 0 {
					// Cut off the last newline
					secretOutput = secretOutput[:len(secretOutput)-1]
					if linkedService.MountVolume {
						output += fmt.Sprintf(" · %s\n   Files:\n%s\n", linkedService.ServiceName, secretOutput)
					} else {
						output += fmt.Sprintf(" · %s\n   Environment Variables:\n%s\n", linkedService.ServiceName, secretOutput)
					}
				}

			} else {
				output += fmt.Sprintf(" · %s\n", linkedService.SecretName)
			}

		}

		if len(output) > 0 {
			// Cut off the last newline and output
			output = output[:len(output)-1]
			log.Describef("Linked Services:\n", output)
		}

	}
	return nil
}
