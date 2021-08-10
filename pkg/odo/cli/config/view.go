package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`# For viewing the current configuration from devfile or local config file
   %[1]s
   
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
	contextDir  string
	lci         *config.LocalConfigInfo
	devfilePath string
	devfileObj  parser.DevfileObj
	IsDevfile   bool
}

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions() *ViewOptions {

	return &ViewOptions{}
}

// Complete completes ViewOptions after they've been created
func (o *ViewOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if errY := util.IsValidKubeConfigPath(); errY != nil {
		return errY
	}
	devfilePath := filepath.Join(o.contextDir, "devfile.yaml")
	if util.CheckPathExists(devfilePath) {
		o.devfilePath = devfilePath
		o.IsDevfile = true
		o.devfileObj, err = parser.Parse(o.devfilePath)
		if err != nil {
			return err
		}
	}

	if !o.IsDevfile {
		cfg, err := config.NewLocalConfigInfo(o.contextDir)
		if err != nil {
			return err
		}
		o.lci = cfg
	}
	return
}

// Validate validates the ViewOptions based on completed values
func (o *ViewOptions) Validate() (err error) {
	if !o.IsDevfile {
		if !o.lci.Exists() {
			return errors.New("the directory doesn't contain a component. Use 'odo create' to create a component")
		}
	}

	return
}

// DevfileRun is ran when the context detects a devfile locally
func (o *ViewOptions) DevfileRun() (err error) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
	repr, err := component.ToDevfileRepresentation(o.devfileObj)
	if err != nil {
		return err
	}
	if log.IsJSON() {
		machineoutput.OutputSuccess(component.WrapFromJSONOutput(repr))
		return
	}
	representation, err := yaml.Marshal(repr)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(representation))
	return err
}

// Run contains the logic for the command
func (o *ViewOptions) Run(cmd *cobra.Command) (err error) {

	if o.IsDevfile {
		return o.DevfileRun()
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)

	cs := o.lci.GetComponentSettings()
	envVarList := o.lci.GetEnvVars()
	if len(envVarList) != 0 {
		fmt.Fprintln(w, "ENVIRONMENT VARIABLES")
		fmt.Fprintln(w, "------------------------------------------------")
		fmt.Fprintln(w, "NAME", "\t", "VALUE")
		for _, envVar := range envVarList {
			fmt.Fprintln(w, envVar.Name, "\t", envVar.Value)
		}

		fmt.Fprintln(w)

	}
	fmt.Fprintln(w, "COMPONENT SETTINGS")
	fmt.Fprintln(w, "------------------------------------------------")

	fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
	fmt.Fprintln(w, "Type", "\t", showBlankIfNil(cs.Type))
	fmt.Fprintln(w, "Application", "\t", showBlankIfNil(cs.Application))
	fmt.Fprintln(w, "Project", "\t", showBlankIfNil(cs.Project))
	fmt.Fprintln(w, "SourceType", "\t", showBlankIfNil(cs.SourceType))
	fmt.Fprintln(w, "Ref", "\t", showBlankIfNil(cs.Ref))
	fmt.Fprintln(w, "SourceLocation", "\t", showBlankIfNil(cs.SourceLocation))
	fmt.Fprintln(w, "Ports", "\t", formatArray(cs.Ports))
	fmt.Fprintln(w, "Name", "\t", showBlankIfNil(cs.Name))
	fmt.Fprintln(w, "MinMemory", "\t", showBlankIfNil(cs.MinMemory))
	fmt.Fprintln(w, "MaxMemory", "\t", showBlankIfNil(cs.MaxMemory))
	fmt.Fprintln(w, "DebugPort", "\t", showBlankIfNil(cs.DebugPort))
	fmt.Fprintln(w, "Ignore", "\t", showBlankIfNil(cs.Ignore))
	fmt.Fprintln(w, "MinCPU", "\t", showBlankIfNil(cs.MinCPU))
	fmt.Fprintln(w, "MaxCPU", "\t", showBlankIfNil(cs.MaxCPU))
	w.Flush()
	return

}

func showBlankIfNil(intf interface{}) interface{} {
	imm := reflect.ValueOf(intf)

	// if the value is nil then we should return a blank string
	if imm.IsNil() {
		return ""
	}

	// if its a pointer then we should de-ref it because we cant de-ref an interface{}
	if imm.Kind() == reflect.Ptr {
		return imm.Elem().Interface()
	}

	return intf
}
func formatArray(arr *[]string) string {
	if arr == nil {
		return ""
	}
	if len(*arr) == 0 {
		return ""
	}
	return strings.Join(*arr, ",")
}

// NewCmdView implements the config view odo command
func NewCmdView(name, fullName string) *cobra.Command {
	o := NewViewOptions()
	configurationViewCmd := &cobra.Command{
		Use:         name,
		Short:       "View current configuration values",
		Long:        "View current configuration values",
		Annotations: map[string]string{"machineoutput": "json"},
		Example:     fmt.Sprintf(fmt.Sprint("\n", viewExample), fullName),
		Args:        cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	genericclioptions.AddContextFlag(configurationViewCmd, &o.contextDir)

	return configurationViewCmd
}
