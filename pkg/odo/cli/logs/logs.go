package logs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

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
}

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
	return nil
}

func (o *LogsOptions) Run(ctx context.Context) error {
	containersLogs, err := o.clientset.LogsClient.DevModeLogs(o.componentName, o.Context.GetProject())
	if err != nil {
		return err
	}

	for container, logs := range containersLogs {
		err = printLogs(container, logs, o.out)
		if err != nil {
			return err
		}
	}
	return nil
}

// printLogs prints the logs of the containers with container name prefixed to the log message
func printLogs(containerName string, rd io.ReadCloser, out io.Writer) error {
	reader := bufio.NewReader(rd)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				break
			}
		}
		line = strings.Join([]string{containerName, line}, ": ")
		lines = append(lines, line)
	}

	randomColor := selectRandomColor()
	color.Set(randomColor)
	defer color.Unset()

	for i := 0; i < len(lines); i++ {
		_, err := fmt.Fprintf(out, lines[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func selectRandomColor() color.Attribute {
	colors := []color.Attribute{color.FgRed, color.FgGreen, color.FgYellow, color.FgBlue, color.FgMagenta, color.FgCyan, color.FgWhite}
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(colors)) //#nosec G404
	return colors[i]
}
func NewCmdLogs(name, fullname string) *cobra.Command {
	o := NewLogsOptions()
	logsCmd := &cobra.Command{
		Use:   name,
		Short: "Show logs of all containers of the component",
		Long: `odo logs shows logs of all containers of the component, whether they are running in Dev mode or Deploy mode.
It prefixes each log message with the container name.`,
		Example: fmt.Sprintf(logsExample, fullname),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	clientset.Add(logsCmd, clientset.LOGS, clientset.FILESYSTEM)
	logsCmd.Annotations["command"] = "main"
	logsCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return logsCmd
}
