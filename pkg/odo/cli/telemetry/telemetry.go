package telemetry

import (
	"encoding/json"
	"os"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/segment"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const RecommendedCommandName = "telemetry"

var (
	telemetryExample = `odo telemetry '{"event": "odo list", "properties": {"duration(ms)": 18, "error": "Get \"https://kubernetes.docker.internal:6443/apis/apps.openshift.io/v1?timeout=32s\": dial tcp 127.0.0.1:6443: connect: connection refused", "errortype": "*net.OpError", "success": false, "tty": true, "version": "odo v2.0.8 (791738b24)"}}''`
)

type TelemetryOptions struct {
	telemetryData segment.TelemetryData
}

func NewTelemetryOptions() *TelemetryOptions {
	return &TelemetryOptions{}
}

func (o *TelemetryOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	err = json.Unmarshal([]byte(args[0]), &o.telemetryData)
	return err
}

func (o *TelemetryOptions) Validate() (err error) {
	return err
}

func (o *TelemetryOptions) Run() (err error) {
	var segmentClient *segment.Client
	cfg, err := preference.New()
	if err != nil {
		return errors.Errorf("unable to update data, required preference.yaml file not found.")
	}
	// Initiate the segment client if ConsentTelemetry preference is set to true
	if cfg.GetConsentTelemetry() {
		if os.Getenv(segment.DisableTelemetryEnv) == "true" {
			log.Warningf("Sending telemetry disabled by %s=%s\n", segment.DisableTelemetryEnv, os.Getenv(segment.DisableTelemetryEnv))
		} else {
			if segmentClient, err = segment.NewClient(cfg); err != nil {
				klog.V(4).Infof("Cannot create a segment client, will not send any data: %s", err.Error())
			}
			defer segmentClient.Close()
		}
	}
	if segmentClient != nil {
		if serr := segmentClient.Upload(o.telemetryData); serr != nil {
			klog.V(4).Infof("Cannot send data to telemetry: %q", serr)
		}
		return segmentClient.Close()
	}
	return nil
}

func NewCmdTelemetry(name string) *cobra.Command {
	o := NewTelemetryOptions()
	telemetryCmd := &cobra.Command{
		Use:                    name,
		Short:                  "Collect and upload usage data.",
		Example:                telemetryExample,
		BashCompletionFunction: "",
		Hidden:                 true,
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
