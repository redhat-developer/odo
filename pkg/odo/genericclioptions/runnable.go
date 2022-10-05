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

	"github.com/redhat-developer/odo/pkg/machineoutput"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	commonutil "github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/version"

	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"

	"k8s.io/klog"

	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
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
	Cleanup(err error)
}

// JsonOutputter must be implemented by commands with JSON output
// For these commands, the `-o json` flag will be added
// when err is not nil, the text of the error will be returned in a `message` field on stderr with an exit status of 1
// when err is nil, the result of RunForJsonOutput will be returned in JSON format on stdout with an exit status of 0
type JsonOutputter interface {
	RunForJsonOutput(ctx context.Context) (result interface{}, err error)
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	var err error
	startTime := time.Now()
	cfg, _ := preference.NewClient()
	disableTelemetry, _ := strconv.ParseBool(os.Getenv(segment.DisableTelemetryEnv))
	debugTelemetry := segment.GetDebugTelemetryFile()

	// Prompt the user to consent for telemetry if a value is not set already
	// Skip prompting if the preference command is called
	// This prompt has been placed here so that it does not prompt the user when they call --help
	if !cfg.IsSet(preference.ConsentTelemetrySetting) && cmd.Parent().Name() != "preference" {
		if !segment.RunningInTerminal() {
			klog.V(4).Infof("Skipping telemetry question because there is no terminal (tty)\n")
		} else if disableTelemetry {
			klog.V(4).Infof("Skipping telemetry question due to %s=%t\n", segment.DisableTelemetryEnv, disableTelemetry)
		} else {
			var consentTelemetry bool
			prompt := &survey.Confirm{Message: "Help odo improve by allowing it to collect usage data. Read about our privacy statement: https://developers.redhat.com/article/tool-data-collection. You can change your preference later by changing the ConsentTelemetry preference.", Default: true}
			err = survey.AskOne(prompt, &consentTelemetry, nil)
			ui.HandleError(err)
			if err == nil {
				if err1 := cfg.SetConfiguration(preference.ConsentTelemetrySetting, strconv.FormatBool(consentTelemetry)); err1 != nil {
					klog.V(4).Info(err1.Error())
				}
			}
		}
	}
	if len(debugTelemetry) > 0 {
		klog.V(4).Infof("WARNING: debug telemetry, if enabled, will be logged in %s", debugTelemetry)
	}

	// set value for telemetry status in context so that we do not need to call IsTelemetryEnabled every time to check its status
	scontext.SetTelemetryStatus(cmd.Context(), segment.IsTelemetryEnabled(cfg))

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

	deps, err := clientset.Fetch(cmd)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	o.SetClientset(deps)

	cmdLineObj := cmdline.NewCobra(cmd)

	ctx := cmdLineObj.Context()
	ctx = fcontext.WithJsonOutput(ctx, commonflags.GetJsonOutputValue(cmdLineObj))
	ctx = fcontext.WithRunOn(ctx, commonflags.GetRunOnValue(cmdLineObj))

	variables, err := commonflags.GetVariablesValues(cmdLineObj)
	util.LogErrorAndExit(err, "")
	ctx = fcontext.WithVariables(ctx, variables)

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
		out, err = jsonOutputter.RunForJsonOutput(cmdLineObj.Context())
		if err == nil {
			machineoutput.OutputSuccess(out)
		}
	} else {
		err = o.Run(ctx)
	}
	startTelemetry(cmd, err, startTime)
	util.LogError(err, "")
	if cleanuper, ok := o.(Cleanuper); ok {
		cleanuper.Cleanup(err)
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
