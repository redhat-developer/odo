package genericclioptions

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
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
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	commonutil "github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/version"

	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/util"
)

type Runnable interface {
	SetClientset(clientset *clientset.Clientset)
	Complete(cmdline cmdline.Cmdline, args []string) error
	Validate() error
	Run(ctx context.Context) error
}

type SignalHandler interface {
	HandleSignal() error
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

	// CheckMachineReadableOutput
	// fixes / checks all related machine readable output functions
	util.LogErrorAndExit(CheckMachineReadableOutputCommand(cmd), "")

	deps, err := clientset.Fetch(cmd)
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	o.SetClientset(deps)

	cmdLineObj := cmdline.NewCobra(cmd)
	// Run completion, validation and run.
	// Only upload data to segment for completion and validation if a non-nil error is returned.
	err = o.Complete(cmdLineObj, args)
	if err != nil {
		startTelemetry(cmd, err, startTime)
	}
	util.LogErrorAndExit(err, "")

	err = o.Validate()
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
		err = o.Run(cmdLineObj.Context())
	}
	startTelemetry(cmd, err, startTime)
	util.LogErrorAndExit(err, "")
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

// CheckMachineReadableOutputCommand performs machine-readable output functions required to
// have it work correctly
func CheckMachineReadableOutputCommand(cmd *cobra.Command) error {

	// Get the needed values
	outputFlag := pflag.Lookup("o")
	hasFlagChanged := outputFlag != nil && outputFlag.Changed
	machineOutput := cmd.Annotations["machineoutput"]

	// Check the valid output
	if hasFlagChanged && outputFlag.Value.String() != "json" {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return errors.New("Please input a valid output format for -o, available format: json")
		//revive:enable:error-strings
	}

	// Check that if -o json has been passed, that the command actually USES json.. if not, error out.
	if hasFlagChanged && outputFlag.Value.String() == "json" && machineOutput == "" {

		// By default we "disable" logging, so activate it so that the below error can be shown.
		_ = flag.Set("o", "")

		// Return the error
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return errors.New("Machine readable output is not yet implemented for this command")
		//revive:enable:error-strings
	}

	// Before running anything, we will make sure that no verbose output is made
	// This is a HACK to manually override `-v 4` to `-v 0` (in which we have no klog.V(0) in our code...
	// in order to have NO verbose output when combining both `-o json` and `-v 4` so json output
	// is not malformed / mixed in with normal logging
	if log.IsJSON() {
		_ = flag.Set("v", "0")
	} else {
		// Override the logging level by the value (if set) by the ODO_LOG_LEVEL env
		// The "-v" flag set on command line will take precedence over ODO_LOG_LEVEL env
		v := flag.CommandLine.Lookup("v").Value.String()
		if level, ok := os.LookupEnv("ODO_LOG_LEVEL"); ok && v == "0" {
			_ = flag.CommandLine.Set("v", level)
		}
	}
	return nil
}
