package component

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"encoding/json"

	"github.com/redhat-developer/odo/pkg/component"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	"github.com/spf13/cobra"
)

// DescribeRecommendedCommandName is the recommended describe command name
const DescribeRecommendedCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe nodejs component,
%[1]s nodejs
`)

// DescribeOptions is a dummy container to attach complete, validate and run pattern
type DescribeOptions struct {
	outputFlag string
	*ComponentOptions
}

// NewDescribeOptions returns new instance of ListOptions
func NewDescribeOptions() *DescribeOptions {
	return &DescribeOptions{"", &ComponentOptions{}}
}

// Complete completes describe args
func (do *DescribeOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	err = do.ComponentOptions.Complete(name, cmd, args)
	return
}

// Validate validates the describe parameters
func (do *DescribeOptions) Validate() (err error) {
	isExists, err := component.Exists(do.Context.Client, do.componentName, do.Context.Application)
	if err != nil {
		return err
	}
	if !isExists {
		return fmt.Errorf("component %s does not exist", do.componentName)
	}

	return odoutil.CheckOutputFlag(do.outputFlag)
}

// Run has the logic to perform the required actions as part of command
func (do *DescribeOptions) Run() (err error) {
	componentDesc, err := component.GetComponentDesc(do.Context.Client, do.componentName, do.Context.Application, do.Context.Project)
	if err != nil {
		return err
	}
	if do.outputFlag == "json" {
		componentDef := getMachineReadableFormat(componentDesc, do.Application, do.Project)
		out, err := json.Marshal(componentDef)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {

		odoutil.PrintComponentInfo(do.componentName, componentDesc)
	}

	return
}

func getMachineReadableFormat(componentDesc component.Description, applicationName, projectName string) component.Component {
	var urls []string
	for _, url := range componentDesc.URLs {
		urls = append(urls, url.Name)
	}

	var storage []string
	for _, store := range componentDesc.Storage {
		storage = append(storage, store.Name)
	}

	currentComponent, err := component.GetCurrent(applicationName, projectName)
	odoutil.LogErrorAndExit(err, "")

	componentDef := component.Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentDesc.ComponentName,
		},
		Spec: component.ComponentSpec{
			Type:    componentDesc.ComponentImageType,
			Source:  componentDesc.Path,
			URL:     urls,
			Storage: storage,
			Env:     componentDesc.Env,
		},
		Status: component.ComponentStatus{
			Active: componentDesc.ComponentName == currentComponent,
		},
	}

	return componentDef
}

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe(name, fullName string) *cobra.Command {
	do := NewDescribeOptions()

	var describeCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [component_name]", name),
		Short:   "Describe the given component",
		Long:    `Describe the given component.`,
		Example: fmt.Sprintf(describeExample, fullName),
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(do, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(describeCmd, completion.ComponentNameCompletionHandler)
	describeCmd.Flags().StringVarP(&do.outputFlag, "output", "o", "", "output in json format")

	//Adding `--project` flag
	projectCmd.AddProjectFlag(describeCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(describeCmd)

	return describeCmd
}
