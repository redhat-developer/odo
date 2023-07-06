package apiserver

import (
	"context"
	"fmt"

	apiserver_impl "github.com/redhat-developer/odo/pkg/apiserver-impl"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const (
	RecommendedCommandName = "api-server"
)

type ApiServerOptions struct {
	clientset         *clientset.Clientset
	apiServerPortFlag int
}

func NewApiServerOptions() *ApiServerOptions {
	return &ApiServerOptions{}
}

var _ genericclioptions.Runnable = (*ApiServerOptions)(nil)

func (o *ApiServerOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ApiServerOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) error {
	return nil
}

func (o *ApiServerOptions) Validate(ctx context.Context) error {
	return nil
}

func (o *ApiServerOptions) Run(ctx context.Context) (err error) {
	err = o.clientset.StateClient.Init(ctx)
	if err != nil {
		err = fmt.Errorf("unable to save state file: %w", err)
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	_ = apiserver_impl.StartServer(
		ctx,
		cancel,
		o.apiServerPortFlag,
		nil,
		nil,
		o.clientset.StateClient,
	)
	<-ctx.Done()
	return nil
}

func (o *ApiServerOptions) Cleanup(ctx context.Context, commandError error) error {
	err := o.clientset.StateClient.SaveExit(ctx)
	if err != nil {
		klog.V(1).Infof("unable to persist dev state: %v", err)
	}
	return nil
}

func (o *ApiServerOptions) HandleSignal(ctx context.Context, cancelFunc context.CancelFunc) error {
	cancelFunc()
	return nil
}

// NewCmdDev implements the odo dev command
func NewCmdApiServer(ctx context.Context, name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	o := NewApiServerOptions()
	apiserverCmd := &cobra.Command{
		Use:   name,
		Short: "Start the API server",
		Long:  "Start the API server",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, testClientset, cmd, args)
		},
	}
	clientset.Add(apiserverCmd,
		clientset.STATE,
	)
	apiserverCmd.Flags().IntVar(&o.apiServerPortFlag, "api-server-port", 0, "Define custom port for API Server.")
	return apiserverCmd
}
