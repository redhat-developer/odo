package telemetry

import (
	"encoding/json"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const RecommendedCommandName = "telemetry"

type TelemetryOptions struct {
	telemetryData segment.TelemetryData
}

func NewTelemetryOptions() *TelemetryOptions {
	return &TelemetryOptions{}
}

func (o *TelemetryOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	err = json.Unmarshal([]byte(args[0]), &o.telemetryData)
	return err
}

func (o *TelemetryOptions) Validate() (err error) {
	return err
}

func (o *TelemetryOptions) Run(cmd *cobra.Command) (err error) {
	cfg, err := preference.New()
	if err != nil {
		return errors.Wrapf(err, "unable to upload telemetry data")
	}

	if !segment.IsTelemetryEnabled(cfg) {
		return nil
	}

	segmentClient, err := segment.NewClient(cfg)
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
	return telemetryCmd
}
