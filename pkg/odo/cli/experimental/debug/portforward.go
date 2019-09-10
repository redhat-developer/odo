package debug

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	k8sgenclioptions "k8s.io/kubernetes/pkg/kubectl/genericclioptions"

	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	// "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

// PortForwardOptions contains all the options for running the port-forward cli command.
type PortForwardOptions struct {
	Namespace  string
	Address    []string
	Ports      []string
	contextDir string

	PortForwarder PortForwarder
	StopChannel   chan struct{}
	ReadyChannel  chan struct{}
	*genericclioptions.Context
	localConfigInfo *config.LocalConfigInfo
}

var (
	portforwardLong = templates.LongDesc(i18n.T(`
                Forward one or more local ports to a pod. This command requires the node to have 'socat' installed.

                Use resource type/name such as deployment/mydeployment to select a pod. Resource type defaults to 'pod' if omitted.

                If there are multiple pods matching the criteria, a pod will be selected automatically. The
                forwarding session ends when the selected pod terminates, and rerun of the command is needed
                to resume forwarding.`))

	portforwardExample = templates.Examples(i18n.T(`
		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod
		odo experimental port-forward pod/mypod 5000 6000

		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in a pod selected by the deployment
		odo experimental debug port-forward deployment/mydeployment 5000 6000

		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in a pod selected by the service
		odo experimental debug port-forward service/myservice 5000 6000

		# Listen on port 8888 locally, forwarding to 5000 in the pod
		odo experimental debug port-forward pod/mypod 8888:5000

		# Listen on port 8888 on all addresses, forwarding to 5000 in the pod
		odo experimental debug port-forward --address 0.0.0.0 pod/mypod 8888:5000

		# Listen on port 8888 on localhost and selected IP, forwarding to 5000 in the pod
		odo experimental debug port-forward --address localhost,10.19.21.23 pod/mypod 8888:5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		odo experimental debug port-forward pod/mypod :5000`))
)

const (
	// Amount of time to wait until at least one pod is running
	defaultPodPortForwardWaitTimeout = 60 * time.Second
)

func NewPortForwardOptions() *PortForwardOptions {
	return &PortForwardOptions{
		PortForwarder: &DefaultPortForwarder{
			IOStreams: streams,
		},
	}
}

type PortForwarder interface {
	ForwardPorts(method string, url *url.URL, opts PortForwardOptions) error
}

type DefaultPortForwarder struct {
	k8sgenclioptions.IOStreams
}

func (f *DefaultPortForwarder) ForwardPorts(method string, url *url.URL, opts PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor()
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.New(dialer, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// splitPort splits port string which is in form of [LOCAL PORT]:REMOTE PORT
// and returns local and remote ports separately
func splitPort(port string) (local, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}

// ConvertPodNamedPortToNumber converts named ports into port numbers
// It returns an error when a named port can't be found in the pod containers
// func ConvertPodNamedPortToNumber(ports []string, pod corev1.Pod) ([]string, error) {
// 	var converted []string
// 	for _, port := range ports {
// 		localPort, remotePort := splitPort(port)

// 		containerPortStr := remotePort
// 		_, err := strconv.Atoi(remotePort)
// 		if err != nil {
// 			containerPort, err := util.LookupContainerPortNumberByName(pod, remotePort)
// 			if err != nil {
// 				return nil, err
// 			}

// 			containerPortStr = strconv.Itoa(int(containerPort))
// 		}

// 		if localPort != remotePort {
// 			converted = append(converted, fmt.Sprintf("%s:%s", localPort, containerPortStr))
// 		} else {
// 			converted = append(converted, containerPortStr)
// 		}
// 	}

// 	return converted, nil
// }

// Complete completes all the required options for port-forward cmd.
func (o *PortForwardOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if len(args) < 2 {
		return cmdutil.UsageErrorf(cmd, "TYPE/NAME and list of ports are required for port-forward")
	}

	o.Context = genericclioptions.NewContext(cmd)
	cfg, err := config.NewLocalConfigInfo(o.contextDir)
	o.localConfigInfo = cfg

	// o.Client.GetOnePodFromSelector

	// o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	// if err != nil {
	// 	return err
	// }

	// builder := f.NewBuilder().
	// 	WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
	// 	ContinueOnError().
	// 	NamespaceParam(o.Namespace).DefaultNamespace()

	// getPodTimeout, err := cmdutil.GetPodRunningTimeoutFlag(cmd)
	// if err != nil {
	// 	return cmdutil.UsageErrorf(cmd, err.Error())
	// }

	// resourceName := args[0]
	// builder.ResourceNames("pods", resourceName)

	// obj, err := builder.Do().Object()
	// if err != nil {
	// 	return err
	// }

	// forwardablePod, err := polymorphichelpers.AttachablePodForObjectFn(f, obj, getPodTimeout)
	// if err != nil {
	// 	return err
	// }

	// o.PodName = pod.Name

	// handle service port mapping to target port if needed

	// o.Ports, err = ConvertPodNamedPortToNumber(args[1:], *forwardablePod)
	// if err != nil {
	// 	return err
	// }
	// clientset, err := f.KubernetesClientSet()
	// if err != nil {
	// 	return err
	// }

	// o.PodClient = clientset.CoreV1()

	// o.Config, err = f.ToRESTConfig()
	// if err != nil {
	// 	return err
	// }
	// o.RESTClient, err = f.RESTClient()
	// if err != nil {
	// 	return err
	// }

	o.StopChannel = make(chan struct{}, 1)
	o.ReadyChannel = make(chan struct{})
	return nil
}

// Validate validates all the required options for port-forward cmd.
func (o PortForwardOptions) Validate() error {

	if len(o.PodName) == 0 {
		return fmt.Errorf("pod name or resource type/name must be specified")
	}

	if len(o.Ports) < 1 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}

	if o.PortForwarder == nil {
		return fmt.Errorf("client, client config, restClient, and portforwarder must be provided")
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

	req := o.RESTClient.Post().
		Resource("pods").
		Namespace(o.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	return o.PortForwarder.ForwardPorts("POST", req.URL(), o)
}

// NewCmdPortForward implements the port-forward odo command
func NewCmdPortForward(name, fullName string) *cobra.Command {

	opts := NewPortForwardOptions()
	cmd := &cobra.Command{
		Use:     name + "port-forward TYPE/NAME [options] [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N]",
		Short:   "Forward one or more local ports to a pod",
		Long:    portforwardLong,
		Example: portforwardExample,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(opts, cmd, args)
		},
	}
	cmdutil.AddPodRunningTimeoutFlag(cmd, defaultPodPortForwardWaitTimeout)
	genericclioptions.AddContextFlag(cmd, &opts.contextDir)

	cmd.Flags().StringSliceVar(&opts.Address, "address", []string{"localhost"}, "Addresses to listen on (comma separated). Only accepts IP addresses or localhost as a value. When localhost is supplied, odo will try to bind on both 127.0.0.1 and ::1 and will fail if neither of these addresses are available to bind.")
	return cmd
}
