package genericclioptions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/AlecAivazis/survey.v1/terminal"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/machineoutput"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	commonutil "github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/version"

	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"

	"k8s.io/klog"
	"k8s.io/utils/pointer"

	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/util"
)

type Runnable interface {
	SetClientset(clientset *clientset.Clientset)
	Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) error
	Validate(ctx context.Context) error
	Run(ctx context.Context) error
}

type SignalHandler interface {
	HandleSignal() error
}

type Cleanuper interface {
	Cleanup(ctx context.Context, err error)
}

// A PreIniter command is a command that will run `init` command if no file is present in current directory
// Commands implementing this interfaec must add FILESYSTEM and INIT dependencies
type PreIniter interface {
	// PreInit indicates a command will run `init`, and display the message returned by the method
	PreInit() string
}

// JsonOutputter must be implemented by commands with JSON output
// For these commands, the `-o json` flag will be added
// when err is not nil, the text of the error will be returned in a `message` field on stderr with an exit status of 1
// when err is nil, the result of RunForJsonOutput will be returned in JSON format on stdout with an exit status of 0
type JsonOutputter interface {
	RunForJsonOutput(ctx context.Context) (result interface{}, err error)
}

const (
	// defaultAppName is the default name of the application when an application name is not provided
	defaultAppName = "app"
)

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	var err error
	startTime := time.Now()
	userConfig, _ := preference.NewClient()

	envConfig, err := config.GetConfiguration()
	if err != nil {
		util.LogErrorAndExit(err, "")
	}

	//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still need to support it
	disableTelemetryEnvSet := envConfig.OdoDisableTelemetry != nil
	var disableTelemetry bool
	if disableTelemetryEnvSet {
		disableTelemetry = *envConfig.OdoDisableTelemetry
	}
	debugTelemetry := pointer.StringDeref(envConfig.OdoDebugTelemetryFile, "")
	isTrackingConsentEnabled, trackingConsentEnvSet, trackingConsentErr := segment.IsTrackingConsentEnabled()

	// check for conflicting settings
	if trackingConsentErr == nil && disableTelemetryEnvSet && trackingConsentEnvSet && disableTelemetry == isTrackingConsentEnabled {
		//lint:ignore SA1019 We deprecated this env var, but we really want users to know there is a conflict here
		util.LogErrorAndExit(
			fmt.Errorf("%[1]s and %[2]s values are in conflict. %[1]s is deprecated, please use only %[2]s",
				segment.DisableTelemetryEnv, segment.TrackingConsentEnv), "")
	}

	// Prompt the user to consent for telemetry if a value is not set already
	// Skip prompting if the preference command is called
	// This prompt has been placed here so that it does not prompt the user when they call --help
	if !userConfig.IsSet(preference.ConsentTelemetrySetting) && cmd.Parent().Name() != "preference" {
		if !segment.RunningInTerminal() {
			klog.V(4).Infof("Skipping telemetry question because there is no terminal (tty)\n")
		} else {
			var askConsent bool
			if trackingConsentErr != nil {
				klog.V(4).Infof("error in determining value of tracking consent env var: %v", trackingConsentErr)
				askConsent = true
			} else if trackingConsentEnvSet {
				trackingConsent := os.Getenv(segment.TrackingConsentEnv)
				if isTrackingConsentEnabled {
					klog.V(4).Infof("Skipping telemetry question due to %s=%s\n", segment.TrackingConsentEnv, trackingConsent)
					klog.V(4).Info("Telemetry is enabled!\n")
					if err1 := userConfig.SetConfiguration(preference.ConsentTelemetrySetting, "true"); err1 != nil {
						klog.V(4).Info(err1.Error())
					}
				} else {
					klog.V(4).Infof("Skipping telemetry question due to %s=%s\n", segment.TrackingConsentEnv, trackingConsent)
				}
			} else if disableTelemetry {
				//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still need to support it
				klog.V(4).Infof("Skipping telemetry question due to %s=%t\n", segment.DisableTelemetryEnv, disableTelemetry)
			} else {
				askConsent = true
			}
			if askConsent {
				var consentTelemetry bool
				prompt := &survey.Confirm{Message: "Help odo improve by allowing it to collect usage data. Read about our privacy statement: https://developers.redhat.com/article/tool-data-collection. You can change your preference later by changing the ConsentTelemetry preference.", Default: true}
				err = survey.AskOne(prompt, &consentTelemetry, nil)
				ui.HandleError(err)
				if err == nil {
					if err1 := userConfig.SetConfiguration(preference.ConsentTelemetrySetting, strconv.FormatBool(consentTelemetry)); err1 != nil {
						klog.V(4).Info(err1.Error())
					}
				}
			}
		}
	}
	if len(debugTelemetry) > 0 {
		klog.V(4).Infof("WARNING: debug telemetry, if enabled, will be logged in %s", debugTelemetry)
	}

	err = scontext.SetCaller(cmd.Context(), os.Getenv(segment.TelemetryCaller))
	if err != nil {
		klog.V(3).Infof("error handling caller property for telemetry: %v", err)
	}

	scontext.SetFlags(cmd.Context(), cmd.Flags())
	// set value for telemetry status in context so that we do not need to call IsTelemetryEnabled every time to check its status
	scontext.SetTelemetryStatus(cmd.Context(), segment.IsTelemetryEnabled(userConfig))

	// Send data to telemetry in case of user interrupt
	captureSignals := []os.Signal{syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt}
	go commonutil.StartSignalWatcher(captureSignals, func(receivedSignal os.Signal) {
		err = fmt.Errorf("user interrupted the command execution: %w", terminal.InterruptErr)
		if handler, ok := o.(SignalHandler); ok {
			err = handler.HandleSignal()
			if err != nil {
				log.Errorf("error handling interrupt signal : %v", err)
			}
		}
		scontext.SetSignal(cmd.Context(), receivedSignal)
		startTelemetry(cmd, err, startTime)
	})

	util.LogErrorAndExit(commonflags.CheckMachineReadableOutputCommand(cmd), "")
	util.LogErrorAndExit(commonflags.CheckRunOnCommand(cmd), "")
	util.LogErrorAndExit(commonflags.CheckVariablesCommand(cmd), "")

	cmdLineObj := cmdline.NewCobra(cmd)
	platform := commonflags.GetRunOnValue(cmdLineObj)
	deps, err := clientset.Fetch(cmd, platform)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	o.SetClientset(deps)

	ctx := cmdLineObj.Context()
	ctx = fcontext.WithJsonOutput(ctx, commonflags.GetJsonOutputValue(cmdLineObj))
	ctx = fcontext.WithRunOn(ctx, platform)
	ctx = odocontext.WithApplication(ctx, defaultAppName)

	if deps.KubernetesClient != nil {
		namespace := deps.KubernetesClient.GetCurrentNamespace()
		ctx = odocontext.WithNamespace(ctx, namespace)
	}

	if deps.FS != nil {
		var cwd string
		cwd, err = deps.FS.Getwd()
		if err != nil {
			startTelemetry(cmd, err, startTime)
		}
		util.LogErrorAndExit(err, "")
		ctx = odocontext.WithWorkingDirectory(ctx, cwd)

		var variables map[string]string
		variables, err = commonflags.GetVariablesValues(cmdLineObj)
		util.LogErrorAndExit(err, "")
		ctx = fcontext.WithVariables(ctx, variables)

		if preiniter, ok := o.(PreIniter); ok {
			msg := preiniter.PreInit()
			err = runPreInit(cwd, deps, cmdLineObj, msg)
			if err != nil {
				startTelemetry(cmd, err, startTime)
			}
			util.LogErrorAndExit(err, "")
		}

		var devfilePath, componentName string
		var devfileObj *parser.DevfileObj
		devfilePath, devfileObj, componentName, err = getDevfileInfo(cwd, variables)
		if err != nil {
			startTelemetry(cmd, err, startTime)
		}
		util.LogErrorAndExit(err, "")
		ctx = odocontext.WithDevfilePath(ctx, devfilePath)
		ctx = odocontext.WithDevfileObj(ctx, devfileObj)
		ctx = odocontext.WithComponentName(ctx, componentName)
	}

	// Run completion, validation and run.
	// Only upload data to segment for completion and validation if a non-nil error is returned.
	err = o.Complete(ctx, cmdLineObj, args)
	if err != nil {
		startTelemetry(cmd, err, startTime)
	}
	util.LogErrorAndExit(err, "")

	err = o.Validate(ctx)
	if err != nil {
		startTelemetry(cmd, err, startTime)
	}
	util.LogErrorAndExit(err, "")

	if jsonOutputter, ok := o.(JsonOutputter); ok && log.IsJSON() {
		var out interface{}
		out, err = jsonOutputter.RunForJsonOutput(ctx)
		if err == nil {
			machineoutput.OutputSuccess(out)
		}
	} else {
		err = o.Run(ctx)
	}
	startTelemetry(cmd, err, startTime)
	util.LogError(err, "")
	if cleanuper, ok := o.(Cleanuper); ok {
		cleanuper.Cleanup(ctx, err)
	}
	if err != nil {
		os.Exit(1)
	}
}

// startTelemetry uploads the data to segment if user has consented to usage data collection and the command is not telemetry
// TODO: move this function to a more suitable place, preferably pkg/segment
func startTelemetry(cmd *cobra.Command, err error, startTime time.Time) {
	if scontext.GetTelemetryStatus(cmd.Context()) && !strings.Contains(cmd.CommandPath(), "telemetry") {
		uploadData := &segment.TelemetryData{
			Event: cmd.CommandPath(),
			Properties: segment.TelemetryProperties{
				Duration:      time.Since(startTime).Milliseconds(),
				Success:       err == nil,
				Tty:           segment.RunningInTerminal(),
				Version:       fmt.Sprintf("odo %v (%v)", version.VERSION, version.GITCOMMIT),
				CmdProperties: scontext.GetContextProperties(cmd.Context()),
			},
		}
		if err != nil {
			uploadData.Properties.Error = segment.SetError(err)
			uploadData.Properties.ErrorType = segment.ErrorType(err)
		}
		data, err1 := json.Marshal(uploadData)
		if err1 != nil {
			klog.V(4).Infof("Failed to marshall telemetry data. %q", err1.Error())
		}
		command := exec.Command(os.Args[0], "telemetry", string(data))
		if err1 = command.Start(); err1 != nil {
			klog.V(4).Infof("Failed to start the telemetry process. Error: %q", err1.Error())
			return
		}
		if err1 = command.Process.Release(); err1 != nil {
			klog.V(4).Infof("Failed to release the process. %q", err1.Error())
			return
		}
	}
}

// NoArgsAndSilenceJSON returns the NoArgs value, and silence output when JSON output is activated
func NoArgsAndSilenceJSON(cmd *cobra.Command, args []string) error {
	if log.IsJSON() {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	}
	return cobra.NoArgs(cmd, args)
}
