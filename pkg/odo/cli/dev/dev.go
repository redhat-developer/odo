package dev

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/version"
	"github.com/redhat-developer/odo/pkg/watch"

	dfutil "github.com/devfile/library/pkg/util"
	ododevfile "github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "dev"

type DevOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Variables
	ignorePaths []string
	out         io.Writer
	errOut      io.Writer
	envFileInfo *envinfo.EnvSpecificInfo
	// it's called "initial" because it has to be set only once when running odo dev for the first time
	initialDevfileObj parser.DevfileObj

	// working directory
	contextDir string
}

type DevHandler struct{}

func NewDevHandler() *DevHandler {
	return &DevHandler{}
}

func NewDevOptions() *DevOptions {
	return &DevOptions{
		out:    log.GetStdout(),
		errOut: log.GetStderr(),
	}
}

var devExample = templates.Examples(`
	# Deploy component to the development cluster
	%[1]s
`)

func (o *DevOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *DevOptions) Complete(cmdline cmdline.Cmdline, args []string) error {
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
		return errors.New("this command cannot run in an empty directory, you need to run it in a directory containing source code")
	}

	err = o.clientset.InitClient.InitDevfile(cmdline.GetFlags(), o.contextDir,
		func(interactiveMode bool) {
			if interactiveMode {
				fmt.Println("The current directory already contains source code. " +
					"odo will try to autodetect the language and project type in order to select the best suited Devfile for your project.")
			}
		},
		func(newDevfileObj parser.DevfileObj) error {
			return newDevfileObj.WriteYamlDevfile()
		})
	if err != nil {
		return err
	}

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	if err != nil {
		return fmt.Errorf("unable to create context: %v", err)
	}

	o.envFileInfo, err = envinfo.NewEnvSpecificInfo("")
	if err != nil {
		return fmt.Errorf("unable to retrieve configuration information: %v", err)
	}

	o.initialDevfileObj = o.Context.EnvSpecificInfo.GetDevfileObj()

	if !o.envFileInfo.Exists() {
		// if env.yaml doesn't exist, get component name from the devfile.yaml
		var cmpName string
		cmpName, err = component.GatherName(o.EnvSpecificInfo.GetDevfileObj(), o.GetDevfilePath())
		if err != nil {
			return fmt.Errorf("unable to retrieve component name: %w", err)
		}
		// create env.yaml file with component, project/namespace and application info
		// TODO - store only namespace into env.yaml, we don't want to track component or application name via env.yaml
		err = o.envFileInfo.SetComponentSettings(envinfo.ComponentSettings{Name: cmpName, Project: o.GetProject(), AppName: "app"})
		if err != nil {
			return fmt.Errorf("failed to write new env.yaml file: %w", err)
		}
	} else if o.envFileInfo.GetComponentSettings().Project != o.GetProject() {
		// set namespace if the evn.yaml exists; that's the only piece we care about in env.yaml
		err = o.envFileInfo.SetConfiguration("project", o.GetProject())
		if err != nil {
			return fmt.Errorf("failed to update project in env.yaml file: %w", err)
		}
	}

	// 3 steps to evaluate the paths to be ignored when "watching" the pwd/cwd for changes
	// 1. create an empty string slice to which paths like .gitignore, .odo/odo-file-index.json, etc. will be added
	var ignores []string
	err = genericclioptions.ApplyIgnore(&ignores, "")
	if err != nil {
		return err
	}
	// 2. get absolute path of pwd/cwd
	sourcePath, err := dfutil.GetAbsPath("")
	if err != nil {
		return fmt.Errorf("unable to get source path: %w", err)
	}
	// 3. combine 1 & 2 to have absolute paths of all files to be ignored
	o.ignorePaths = dfutil.GetAbsGlobExps(sourcePath, ignores)

	return nil
}

func (o *DevOptions) Validate() error {
	var err error
	return err
}

func (o *DevOptions) Run() error {
	var err error
	var platformContext = kubernetes.KubernetesContext{
		Namespace: o.Context.GetProject(),
	}
	var path = filepath.Dir(o.Context.EnvSpecificInfo.GetDevfilePath())
	devfileName := o.EnvSpecificInfo.GetDevfileObj().GetMetadataName()
	namespace := o.GetProject()

	// Output what the command is doing / information
	log.Title("Developing using the "+devfileName+" Devfile",
		"Namespace: "+namespace,
		"odo version: "+version.VERSION)

	log.Section("Deploying to the cluster in developer mode")
	err = o.clientset.DevClient.Start(o.Context.EnvSpecificInfo.GetDevfileObj(), platformContext, path)
	if err != nil {
		return err
	}
	fmt.Fprintf(o.out, "\nYour application is running on cluster.\n ")

	containers, err := libdevfile.GetContainerComponents(o.Context.EnvSpecificInfo.GetDevfileObj())
	if err != nil {
		return err
	}

	ceMapping, err := libdevfile.GetContainerEndpointMapping(containers)
	if err != nil {
		return err
	}

	portPairs := portPairsFromContainerEndpoints(ceMapping)
	portPairsSlice := []string{}
	for _, v1 := range portPairs {
		portPairsSlice = append(portPairsSlice, v1...)
	}

	go func() {
		err = o.clientset.DevClient.SetupPortForwarding(o.Context.EnvSpecificInfo.GetDevfileObj(), portPairsSlice, o.errOut)
		if err != nil {
			fmt.Printf("failed to setup port-forwarding: %v\n", err)
		}
	}()
	printPortForwardingInfo(portPairs, o.out)

	d := DevHandler{}
	err = o.clientset.DevClient.Watch(o.Context.EnvSpecificInfo.GetDevfileObj(), path, o.ignorePaths, o.out, &d)
	return err
}

func (o *DevHandler) RegenerateAdapterAndPush(pushParams common.PushParameters, watchParams watch.WatchParameters) error {
	var adapter common.ComponentAdapter

	adapter, err := regenerateComponentAdapterFromWatchParams(watchParams)
	if err != nil {
		return fmt.Errorf("unable to generate component from watch parameters: %w", err)
	}

	err = adapter.Push(pushParams)
	if err != nil {
		return fmt.Errorf("watch command was unable to push component: %w", err)
	}

	return nil
}

func regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (common.ComponentAdapter, error) {
	devObj, err := ododevfile.ParseAndValidateFromFile(location.DevfileLocation(""))
	if err != nil {
		return nil, err
	}

	changed, err := haveEndpointsChanged(parameters.InitialDevfileObj, devObj)
	if err != nil {
		return nil, err
	} else if changed {
		fmt.Printf("\ntotal number of endpoints in the devfile have changed; please run `odo dev` again\n\n")
	}

	platformContext := kubernetes.KubernetesContext{
		Namespace: parameters.EnvSpecificInfo.GetNamespace(),
	}

	return adapters.NewComponentAdapter(parameters.ComponentName, parameters.Path, parameters.ApplicationName, devObj, platformContext)
}

func haveEndpointsChanged(oldDevfile, newDevfile parser.DevfileObj) (bool, error) {
	oldContainers, err := libdevfile.GetContainerComponents(oldDevfile)
	if err != nil {
		return false, err
	}

	oldEndpoints, err := libdevfile.GetContainerEndpointMapping(oldContainers)
	if err != nil {
		return false, err
	}

	newContainers, err := libdevfile.GetContainerComponents(newDevfile)
	if err != nil {
		return false, err
	}

	newEndpoints, err := libdevfile.GetContainerEndpointMapping(newContainers)
	if err != nil {
		return false, nil
	}

	if !reflect.DeepEqual(oldEndpoints, newEndpoints) {
		return true, nil
	}
	return false, nil
}

func printPortForwardingInfo(portPairs map[string][]string, out io.Writer) {
	portForwardURLs := ""
	for container, ports := range portPairs {
		for _, pair := range ports {
			split := strings.Split(pair, ":")
			local := split[0]
			remote := split[1]

			portForwardURLs += fmt.Sprintf("- Port %s from %q container forwarded to localhost:%s\n", remote, container, local)
		}
	}
	fmt.Fprintf(out, "\n%s", portForwardURLs)
}

// NewCmdDev implements the odo dev command
func NewCmdDev(name, fullName string) *cobra.Command {
	o := NewDevOptions()
	devCmd := &cobra.Command{
		Use:   name,
		Short: "Deploy component to development cluster",
		Long: `odo dev is a long running command that will automatically sync your source to the cluster.
It forwards endpoints with exposure values 'public' or 'internal' to a port on localhost.`,
		Example: fmt.Sprintf(devExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	clientset.Add(devCmd, clientset.DEV, clientset.INIT)
	// Add a defined annotation in order to appear in the help menu
	devCmd.Annotations["command"] = "utility"
	devCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return devCmd
}

// portPairsFromContainerEndpoints assigns a port on localhost to each port in the provided containerEndpoints map
// it returns a map of the format "<container-name>":{"<local-port-1>:<remote-port-1>", "<local-port-2>:<remote-port-2>"}
// "container1": {"400001:3000", "400002:3001"}
func portPairsFromContainerEndpoints(ceMap map[string][]int) map[string][]string {
	portPairs := make(map[string][]string)
	port := 40000

	for name, ports := range ceMap {
		for _, p := range ports {
			port++
			for {
				isPortFree := util.IsPortFree(port)
				if isPortFree {
					pair := fmt.Sprintf("%d:%d", port, p)
					portPairs[name] = append(portPairs[name], pair)
					break
				}
				port++
			}
		}
	}
	return portPairs
}
