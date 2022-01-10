package telemetry

import (
	"encoding/json"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const RecommendedCommandName = "telemetry"

type TelemetryOptions struct {
	prefClient    preference.Client
	telemetryData segment.TelemetryData
}

func NewTelemetryOptions(prefClient preference.Client) *TelemetryOptions {
	return &TelemetryOptions{
		prefClient: prefClient,
	}
}

func (o *TelemetryOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	err = json.Unmarshal([]byte(args[0]), &o.telemetryData)
	return err
}

func (o *TelemetryOptions) Validate() (err error) {
	return err
}

func (o *TelemetryOptions) Run() (err error) {
	if !segment.IsTelemetryEnabled(o.prefClient) {
		return nil
	}

	segmentClient, err := segment.NewClient(o.prefClient)
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
	prefClient, _ := preference.NewClient()
	o := NewTelemetryOptions(prefClient)
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
	return telemetryCmd
}
