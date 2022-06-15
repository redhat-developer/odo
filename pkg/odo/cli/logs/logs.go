package logs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	odolabels "github.com/redhat-developer/odo/pkg/labels"

	"github.com/fatih/color"

	"github.com/redhat-developer/odo/pkg/log"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const RecommendedCommandName = "logs"

type LogsOptions struct {
	// context
	Context *genericclioptions.Context
	// clients
	clientset *clientset.Clientset

	// variables
	componentName string
	contextDir    string
	out           io.Writer

	// flags
	devMode    bool
	deployMode bool
}

type logsMode string

const (
	DevMode    logsMode = "dev"
	DeployMode logsMode = "deploy"
)

func NewLogsOptions() *LogsOptions {
	return &LogsOptions{
		out: log.GetStdout(),
	}
}

var logsExample = ktemplates.Examples(`
	# Show logs of all containers
	%[1]s
`)

func (o *LogsOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *LogsOptions) Complete(cmdline cmdline.Cmdline, args []string) error {
	var err error
	o.contextDir, err = os.Getwd()
	if err != nil {
		return err
	}
	isEmptyDir, err := location.DirIsEmpty(o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
	if isEmptyDir {
		return errors.New("this command cannot run in an empty directory, run the command in a directory containing source code or initialize using 'odo init'")
	}

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	if err != nil {
		return fmt.Errorf("unable to create context: %v", err)
	}

	o.componentName = o.Context.EnvSpecificInfo.GetDevfileObj().GetMetadataName()

	o.clientset.KubernetesClient.SetNamespace(o.Context.GetProject())

	return nil
}

func (o *LogsOptions) Validate() error {
	if o.devMode && o.deployMode {
		return errors.New("pass only one of --dev or --deploy flags; pass no flag to see logs for both modes")
	}
	return nil
}

func (o *LogsOptions) Run(ctx context.Context) error {
	var logMode logsMode
	var err error
	var containersLogs []map[string]io.ReadCloser

	if o.devMode {
		logMode = DevMode
	} else if o.deployMode {
		logMode = DeployMode
	}

	switch logMode {
	case DevMode:
		containersLogs, err = o.clientset.LogsClient.GetLogsForMode(odolabels.ComponentDevMode, o.componentName, o.Context.GetProject())
	case DeployMode:
		containersLogs, err = o.clientset.LogsClient.GetLogsForMode(odolabels.ComponentDeployMode, o.componentName, o.Context.GetProject())
	default:
		containersLogs, err = o.clientset.LogsClient.GetLogsForMode(odolabels.ComponentAnyMode, o.componentName, o.Context.GetProject())
	}
	if err != nil {
		return err
	}
	if len(containersLogs) == 0 {
		// This will be the case when:
		// 1. user specifies --dev flag, but the component's running in Deploy mode
		// 2. user specified --deploy flag, but the component's running in Dev mode
		// 3. user passes no flag, but component is running in neither Dev nor Deploy mode
		fmt.Fprintf(o.out, "no containers running in the specified mode for the component %q", o.componentName)
		return nil
	}

	uniqueContainerNames := map[string]struct{}{}
	for _, entry := range containersLogs {
		for container, logs := range entry {
			uniqueName := getUniqueContainerName(container, uniqueContainerNames)
			uniqueContainerNames[uniqueName] = struct{}{}
			err = printLogs(uniqueName, logs, o.out)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getUniqueContainerName(name string, uniqueNames map[string]struct{}) string {
	if _, ok := uniqueNames[name]; ok {
		// name already present in uniqueNames; find another name
		// first check if last character in name is a number; if so increment it, else append name with [1]
		last, err := strconv.Atoi(string(name[len(name)-2]))
		if err == nil {
			last++
			name = fmt.Sprintf("%s[%d]", name[:len(name)-3], last)
		} else {
			last = 1
			name = fmt.Sprintf("%s[%d]", name, last)
		}
		return getUniqueContainerName(name, uniqueNames)
	}
	return name
}

// printLogs prints the logs of the containers with container name prefixed to the log message
func printLogs(containerName string, rd io.ReadCloser, out io.Writer) error {
	color.Set(log.ColorPicker())
	defer color.Unset()
	scanner := bufio.NewScanner(rd)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		_, err := fmt.Fprintln(out, containerName+": "+line)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewCmdLogs(name, fullname string) *cobra.Command {
	o := NewLogsOptions()
	logsCmd := &cobra.Command{
		Use:   name,
		Short: "Show logs of all containers of the component",
		Long: `odo logs shows logs of all containers of the component. 
By default it shows logs of all containers running in both Dev and Deploy mode. It prefixes each log message with the container name.`,
		Example: fmt.Sprintf(logsExample, fullname),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	logsCmd.Flags().BoolVar(&o.devMode, string(DevMode), false, "Show logs for containers running only in Dev mode")
	logsCmd.Flags().BoolVar(&o.deployMode, string(DeployMode), false, "Show logs for containers running only in Deploy mode")

	clientset.Add(logsCmd, clientset.LOGS, clientset.FILESYSTEM)
	logsCmd.Annotations["command"] = "main"
	logsCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return logsCmd
}
