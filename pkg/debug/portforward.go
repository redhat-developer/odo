package debug

import (
	"github.com/openshift/odo/pkg/occlient"

	componentlabels "github.com/openshift/odo/pkg/component/labels"

	"fmt"
	"net/http"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	k8sgenclioptions "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// DefaultPortForwarder implements the SPDY based port forwarder
type DefaultPortForwarder struct {
	client *occlient.Client
	k8sgenclioptions.IOStreams
	componentName string
	appName       string
}

func NewDefaultPortForwarder(componentName, appName string, client *occlient.Client, streams k8sgenclioptions.IOStreams) *DefaultPortForwarder {
	return &DefaultPortForwarder{
		client:        client,
		IOStreams:     streams,
		componentName: componentName,
		appName:       appName,
	}
}

// ForwardPorts forwards the port using the url for the remote pod.
// portPair is a pair of port in format "localPort:RemotePort" that is to be forwarded
// stop Chan is used to stop port forwarding
// ready Chan is used to signal failure to the channel receiver
func (f *DefaultPortForwarder) ForwardPorts(portPair string, stopChan, readyChan chan struct{}) error {
	conf, err := f.client.KClient.KubeConfig.ClientConfig()
	if err != nil {
		return err
	}

	pod, err := f.getPodUsingComponentName()
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	transport, upgrader, err := spdy.RoundTripperFor(conf)
	if err != nil {
		return err
	}
	req := f.client.BuildPortForwardReq(pod.Name)
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(dialer, []string{portPair}, stopChan, readyChan, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	log.Info("Started port forwarding at ports -", portPair)
	return fw.ForwardPorts()
}

func (f *DefaultPortForwarder) getPodUsingComponentName() (*corev1.Pod, error) {
	componentLabels := componentlabels.GetLabels(f.componentName, f.appName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)
	dc, err := f.client.GetOneDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get deployment for component")
	}
	// Find Pod for component
	podSelector := fmt.Sprintf("deploymentconfig=%s", dc.Name)

	return f.client.GetOnePodFromSelector(podSelector)
}
