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
	if !segment.IsTelemetryEnabled(o.clientset.PreferenceClient) {
		return nil
	}

	dt := pointer.StringDeref(envcontext.GetEnvConfig(ctx).OdoDebugTelemetryFile, "")
	if len(dt) > 0 {
		klog.V(4).Infof("WARNING: telemetry debug enabled, data logged to file %s", dt)
		return util.WriteToJSONFile(o.telemetryData, dt)
	}

	segmentClient, err := segment.NewClient(o.clientset.PreferenceClient)
	if err != nil {
		klog.V(4).Infof("Cannot create a segment client. Will not send any data: %q", err)
	}
	defer segmentClient.Close()

	err = segmentClient.Upload(o.telemetryData)
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
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(telemetryCmd, clientset.PREFERENCE)
	return telemetryCmd
}
