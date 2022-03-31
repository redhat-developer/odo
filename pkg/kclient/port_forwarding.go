package kclient

import (
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func (c *Client) SetupPortForwarding(pod *corev1.Pod, portPairs []string, out io.Writer, errOut io.Writer) error {
	transport, upgrader, err := spdy.RoundTripperFor(c.GetClientConfig())
	if err != nil {
		return err
	}

	req := c.GeneratePortForwardReq(pod.Name)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	stopChan := make(chan struct{}, 1)
	// passing nil for readyChan because it's eventually being closed if it's not nil
	// passing nil for out because we only care for error, not for output messages; we want to print our own messages
	fw, err := portforward.New(dialer, portPairs, stopChan, nil, out, errOut)
	if err != nil {
		return err
	}

	// start port-forwarding
	err = fw.ForwardPorts()
	if err != nil {
		// do cleanup when this happens
		// TODO: #5485
		return err
	}

	return nil
}
