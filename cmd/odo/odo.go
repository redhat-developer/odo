package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/odo/cli"
	"github.com/redhat-developer/odo/pkg/odo/cli/version"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

const pluginSuffix = ".odo.plugin"

var plugins = make(map[string]plugin, 10)

type plugin interface {
	execute(args []string) error
}

type pluginImpl struct {
	path string
}

func (p pluginImpl) execute(args []string) error {
	return syscall.Exec(p.path, append([]string{p.path}, args...), os.Environ())
}

func main() {
	// create the complete command
	root := cli.NewCmdOdo(cli.OdoRecommendedName, cli.OdoRecommendedName)

	loadPlugins(root)

	rootCmp := createCompletion(root)
	cmp := complete.New("odo", rootCmp)

	// AddFlags adds the completion flags to the program flags, specifying custom names
	cmp.CLI.InstallName = "complete"
	cmp.CLI.UninstallName = "uncomplete"
	cmp.AddFlags(nil)

	// add the completion flags to the root command, though they won't appear in completions
	root.Flags().AddGoFlagSet(flag.CommandLine)
	// override usage so that flag.Parse uses root command's usage instead of default one when invoked with -h
	flag.Usage = func() {
		_ = root.Help()
	}

	// parse the flags but hack around to avoid exiting with error code 2 on help
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	args := os.Args[1:]
	if err := flag.CommandLine.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
	}

	// run the completion, in case that the completion was invoked
	// and ran as a completion script or handled a flag that passed
	// as argument, the Run method will return true,
	// in that case, our program have nothing to do and should return.
	if cmp.Complete() {
		return
	}

	// Call commands
	// checking the value of updatenotification in config
	// before proceeding with fetching the latest version
	cfg, err := config.New()
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	if cfg.GetUpdateNotification() {
		updateInfo := make(chan string)
		go version.GetLatestReleaseInfo(updateInfo)

		util.LogErrorAndExit(root.Execute(), "")
		select {
		case message := <-updateInfo:
			fmt.Println(message)
		default:
			glog.V(4).Info("Could not get the latest release information in time. Never mind, exiting gracefully :)")
		}
	} else {
		util.LogErrorAndExit(root.Execute(), "")
	}
}

func createCompletion(root *cobra.Command) complete.Command {
	rootCmp := complete.Command{}
	rootCmp.Flags = make(complete.Flags)
	addFlags := func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		var handler complete.Predictor
		handler, ok := completion.GetCommandFlagHandler(root, flag.Name)
		if !ok {
			handler = complete.PredictAnything
		}

		if len(flag.Shorthand) > 0 {
			rootCmp.Flags["-"+flag.Shorthand] = handler
		}

		rootCmp.Flags["--"+flag.Name] = handler
	}
	root.LocalFlags().VisitAll(addFlags)
	root.InheritedFlags().VisitAll(addFlags)
	if root.HasAvailableSubCommands() {
		rootCmp.Sub = make(complete.Commands)
		for _, c := range root.Commands() {
			if !c.Hidden {
				rootCmp.Sub[c.Name()] = createCompletion(c)
			}
		}
	}

	var handler complete.Predictor
	handler, ok := completion.GetCommandHandler(root)
	if !ok {
		handler = complete.PredictNothing
	}
	rootCmp.Args = handler

	return rootCmp
}

func loadPlugins(root *cobra.Command) {
	configDir, err := config.GetPluginsDir()
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, pluginSuffix) {
			pluginName, err := getPluginNameFromPath(path)
			if err != nil {
				return err
			}
			plugin := pluginImpl{
				path: path,
			}
			plugins[pluginName] = plugin
			root.AddCommand(&cobra.Command{
				Use:   pluginName,
				Short: fmt.Sprintf("Execute %s plugin", pluginName),
				RunE: func(cmd *cobra.Command, args []string) error {
					return plugin.execute(args)
				},
				Annotations: map[string]string{"command": "plugin"},
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func getPluginNameFromPath(pluginPath string) (string, error) {
	_, file := path.Split(pluginPath)
	index := strings.LastIndex(file, pluginSuffix)
	if index < 0 {
		return "", fmt.Errorf("plugin at %s doesn't follow the '<name>%s' format", pluginPath, pluginSuffix)
	}

	return file[:index], nil
}
