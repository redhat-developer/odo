package debug

import (
	"fmt"
	"os"
	"os/signal"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/debug"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	k8sgenclioptions "k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

// PortForwardOptions contains all the options for running the port-forward cli command.
type PortForwardOptions struct {
	Namespace string
	// PortPair is the combination of local and remote port in the format "local:remote"
	PortPair string

	localPort  int
	contextDir string

	PortForwarder *debug.DefaultPortForwarder
	// StopChannel is used to stop port forwarding
	StopChannel chan struct{}
	// ReadChannel is used to receive status of port forwarding ( ready or not ready )
	ReadyChannel chan struct{}
	*genericclioptions.Context
	localConfigInfo *config.LocalConfigInfo
}

var (
	portforwardLong = templates.LongDesc(`
			Forward a local port to a remote port on the pod where the application is listening for a debugger.

			By default the local port and the remote port will be same. To change the local port use can use --local-port argument and to change the remote port use "odo config set DebugPort <port>"   		  
	`)

	portforwardExample = templates.Examples(`
		# Listen on default port and forwarding to the default port in the pod
		odo debug port-forward 

		# Listen on the 5000 port locally, forwarding to default port in the pod
		odo debug port-forward --local-port 5000
		
		`)
)

const (
	portforwardCommandName = "port-forward"
)

func NewPortForwardOptions() *PortForwardOptions {
	return &PortForwardOptions{}
}

// Complete completes all the required options for port-forward cmd.
func (o *PortForwardOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	o.Context = genericclioptions.NewContext(cmd)
	cfg, err := config.NewLocalConfigInfo(o.contextDir)
	o.localConfigInfo = cfg

	remotePort := cfg.GetDebugPort()
	o.PortPair = fmt.Sprintf("%d:%d", o.localPort, remotePort)

	kubeConf, err := o.Context.Client.KubeConfig.ClientConfig()
	if err != nil {
		return err
	}
	// Using Discard streams because nothing important is logged
	o.PortForwarder = debug.NewDefaultPortForwarder(kubeConf, k8sgenclioptions.NewTestIOStreamsDiscard())

	o.StopChannel = make(chan struct{}, 1)
	o.ReadyChannel = make(chan struct{})
	return nil
}

// Validate validates all the required options for port-forward cmd.
func (o PortForwardOptions) Validate() error {

	if len(o.PortPair) < 1 {
		return fmt.Errorf("ports cannot be empty")
	}
	return nil
}

// Run implements all the necessary functionality for port-forward cmd.
func (o PortForwardOptions) Run() error {
	componentName := o.localConfigInfo.GetName()
	appName := o.localConfigInfo.GetApplication()
	componentLabels := componentlabels.GetLabels(componentName, appName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)
	dc, err := o.Client.GetOneDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return errors.Wrap(err, "unable to get deployment for component")
	}
	// Find Pod for component
	podSelector := fmt.Sprintf("deploymentconfig=%s", dc.Name)

	pod, err := o.Client.GetOnePodFromSelector(podSelector)
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		<-signals
		if o.StopChannel != nil {
			close(o.StopChannel)
		}
	}()

	req := o.Client.BuildPortForwardReq(pod.Name)
	log.Info("Started port forwarding at ports -", o.PortPair)
	return o.PortForwarder.ForwardPorts("POST", req.URL(), []string{o.PortPair}, o.StopChannel, o.ReadyChannel)
}

// NewCmdPortForward implements the port-forward odo command
func NewCmdPortForward(name, fullName string) *cobra.Command {

	opts := NewPortForwardOptions()
	cmd := &cobra.Command{
		Use:     name,
		Short:   "Forward one or more local ports to a pod",
		Long:    portforwardLong,
		Example: portforwardExample,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(opts, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(cmd, &opts.contextDir)
	cmd.Flags().IntVarP(&opts.localPort, "local-port", "l", config.DefaultDebugPort, "Set the local port")

	return cmd
}
