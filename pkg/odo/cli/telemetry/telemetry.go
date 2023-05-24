package telemetry

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/segment"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/util"

	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

const RecommendedCommandName = "telemetry"

type TelemetryOptions struct {
	// clients
	clientset *clientset.Clientset

	telemetryData segment.TelemetryData
}

var _ genericclioptions.Runnable = (*TelemetryOptions)(nil)

func NewTelemetryOptions() *TelemetryOptions {
	return &TelemetryOptions{}
}

func (o *TelemetryOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *TelemetryOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	err = json.Unmarshal([]byte(args[0]), &o.telemetryData)
	return err
}

func (o *TelemetryOptions) Validate(ctx context.Context) (err error) {
	return err
}

func (o *TelemetryOptions) Run(ctx context.Context) (err error) {
	// Telemetry is sent only if it was enabled previously (GetPreviousTelemetryStatus) or if it is currently enabled (GetTelemetryStatus).
	// For example, if it was enabled previously, and user disables telemetry, we still want the event disabling it to be recorded.
	// And if it was disabled, and now it is enabled, we want to track this event as well.
	var wasTelemetryEnabled bool
	val, ok := o.telemetryData.Properties.CmdProperties[scontext.PreviousTelemetryStatus]
	if ok {
		wasTelemetryEnabled = val.(bool)
	}
	if !(wasTelemetryEnabled || segment.IsTelemetryEnabled(o.clientset.PreferenceClient, envcontext.GetEnvConfig(ctx))) {
		klog.V(2).Infof("Telemetry not enabled!")
		return nil
	}

	dt := pointer.StringDeref(envcontext.GetEnvConfig(ctx).OdoDebugTelemetryFile, "")
	if len(dt) > 0 {
		klog.V(4).Infof("WARNING: telemetry debug enabled, data logged to file %s", dt)
		return util.WriteToJSONFile(o.telemetryData, dt)
	}

	segmentClient, err := segment.NewClient()
	if err != nil {
		klog.V(4).Infof("Cannot create a segment client. Will not send any data: %q", err)
	}
	defer segmentClient.Close()

	err = segmentClient.Upload(ctx, o.telemetryData)
	if err != nil {
		klog.V(4).Infof("Cannot send data to telemetry: %q", err)
	}

	return segmentClient.Close()
}

func NewCmdTelemetry(name string) *cobra.Command {
	o := NewTelemetryOptions()
	telemetryCmd := &cobra.Command{
		Use:                    name,
		Short:                  "Collect and upload usage data.",
		BashCompletionFunction: "",
		Hidden:                 true,
		Args:                   cobra.ExactArgs(1),
		Annotations:            map[string]string{},
		SilenceErrors:          true,
		SilenceUsage:           true,
		DisableFlagsInUseLine:  true,
		DisableSuggestions:     true,
		FParseErrWhitelist:     cobra.FParseErrWhitelist{},
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(telemetryCmd, clientset.PREFERENCE)
	return telemetryCmd
}
