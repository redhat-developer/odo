package debug

import (
	"github.com/redhat-developer/odo/pkg/kclient"
	"k8s.io/client-go/rest"

	"fmt"
	"net/http"

	"github.com/redhat-developer/odo/pkg/log"
	corev1 "k8s.io/api/core/v1"

	k8sgenclioptions "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// DefaultPortForwarder implements the SPDY based port forwarder
type DefaultPortForwarder struct {
	kClient kclient.ClientInterface
	k8sgenclioptions.IOStreams
	componentName string
	appName       string
	projectName   string
}

func NewDefaultPortForwarder(componentName, appName string, projectName string, kClient kclient.ClientInterface, streams k8sgenclioptions.IOStreams) *DefaultPortForwarder {
	return &DefaultPortForwarder{
		kClient:       kClient,
		IOStreams:     streams,
		componentName: componentName,
		appName:       appName,
		projectName:   projectName,
	}
}

// ForwardPorts forwards the port using the url for the remote pod.
// portPair is a pair of port in format "localPort:RemotePort" that is to be forwarded
// stop Chan is used to stop port forwarding
// ready Chan is used to signal failure to the channel receiver
func (f *DefaultPortForwarder) ForwardPorts(portPair string, stopChan, readyChan chan struct{}, isDevfile bool) error {
	var pod *corev1.Pod
	var conf *rest.Config
	var err error

	if f.kClient != nil && isDevfile {
		conf, err = f.kClient.GetConfig().ClientConfig()
		if err != nil {
			return err
		}

		pod, err = f.kClient.GetOnePod(f.componentName, f.appName)
		if err != nil {
			return err
		}
	} else {
		conf = f.kClient.GetClientConfig()
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	transport, upgrader, err := spdy.RoundTripperFor(conf)
	if err != nil {
		return err
	}

	req := f.kClient.GeneratePortForwardReq(pod.Name)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(dialer, []string{portPair}, stopChan, readyChan, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	log.Info("Started port forwarding at ports -", portPair)
	return fw.ForwardPorts()
}
