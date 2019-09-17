package debug

import (
	"net/http"
	"net/url"

	"github.com/openshift/odo/pkg/occlient"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	k8sgenclioptions "k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

// PortForwarder is the interface which is needed to be implemented by a port forwarding implementation
type PortForwarder interface {
	ForwardPorts(method string, url *url.URL, ports []string, stopChan, readyChan chan struct{}) error
}

// DefaultPortForwarder implements the SPDY based port forwarder
type DefaultPortForwarder struct {
	client *occlient.Client
	k8sgenclioptions.IOStreams
}

func NewDefaultPortForwarder(client *occlient.Client, streams k8sgenclioptions.IOStreams) *DefaultPortForwarder {
	return &DefaultPortForwarder{
		client:    client,
		IOStreams: streams,
	}
}

// ForwardPorts forwards the ports using the url for the remote pod.
// stop Chan is used to stop port forwarding
// ready Chan is used to signal failure to the channel receiver
func (f *DefaultPortForwarder) ForwardPorts(method string, url *url.URL, ports []string, stopChan, readyChan chan struct{}) error {
	conf, err := f.client.KubeConfig.ClientConfig()
	if err != nil {
		return err
	}
	transport, upgrader, err := spdy.RoundTripperFor(conf)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.New(dialer, ports, stopChan, readyChan, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
