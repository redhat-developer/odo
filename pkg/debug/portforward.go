package debug

import (
	"net/http"
	"net/url"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	k8sgenclioptions "k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

// DefaultPortForwarder implements the SPDY based port forwarder
type DefaultPortForwarder struct {
	config *restclient.Config
	k8sgenclioptions.IOStreams
}

func NewDefaultPortForwarder(config *restclient.Config, streams k8sgenclioptions.IOStreams) *DefaultPortForwarder {
	return &DefaultPortForwarder{
		config:    config,
		IOStreams: streams,
	}
}

// ForwardPorts forwards the ports using the url for the remote pod.
// ports are list of pair of ports in format "localPort:RemotePort" that are to be forwarded
// stop Chan is used to stop port forwarding
// ready Chan is used to signal failure to the channel receiver
func (f *DefaultPortForwarder) ForwardPorts(method string, url *url.URL, ports []string, stopChan, readyChan chan struct{}) error {

	transport, upgrader, err := spdy.RoundTripperFor(f.config)
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
