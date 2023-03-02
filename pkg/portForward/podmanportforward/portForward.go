package podmanportforward

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/remotecmd"
)

const pfHelperContainer = "odo-helper-port-forwarding"

type PFClient struct {
	remoteProcessHandler remotecmd.RemoteProcessHandler

	appliedPorts map[api.ForwardedPort]struct{}
}

var _ portForward.Client = (*PFClient)(nil)

func NewPFClient(execClient exec.Client) *PFClient {
	return &PFClient{
		remoteProcessHandler: remotecmd.NewKubeExecProcessHandler(execClient),
		appliedPorts:         make(map[api.ForwardedPort]struct{}),
	}
}

func (o *PFClient) StartPortForwarding(
	devFileObj parser.DevfileObj,
	componentName string,
	debug bool,
	randomPorts bool,
	out io.Writer,
	errOut io.Writer,
	definedPorts []api.ForwardedPort,
) error {
	var appliedPorts []api.ForwardedPort
	for port := range o.appliedPorts {
		appliedPorts = append(appliedPorts, port)
	}
	if reflect.DeepEqual(appliedPorts, definedPorts) {
		klog.V(3).Infof("Port forwarding should already be running for defined ports: %v", definedPorts)
		return nil
	}

	o.StopPortForwarding(componentName)

	outputHandler := func(fwPort api.ForwardedPort) remotecmd.CommandOutputHandler {
		return func(status remotecmd.RemoteProcessStatus, stdout []string, stderr []string, err error) {
			klog.V(4).Infof("Status for port-forwarding (from %s:%d -> %d): %s", fwPort.LocalAddress, fwPort.LocalPort, fwPort.ContainerPort, status)
			klog.V(4).Info(strings.Join(stdout, "\n"))
			klog.V(4).Info(strings.Join(stderr, "\n"))
			switch status {
			case remotecmd.Running:
				o.appliedPorts[fwPort] = struct{}{}
			case remotecmd.Stopped, remotecmd.Errored:
				delete(o.appliedPorts, fwPort)
				if status == remotecmd.Stopped {
					fmt.Fprintf(out, "Stopped port-forwarding from %s:%d -> %d", fwPort.LocalAddress, fwPort.LocalPort, fwPort.ContainerPort)
				}
			}
		}
	}

	for _, port := range definedPorts {
		err := o.remoteProcessHandler.StartProcessForCommand(getCommandDefinition(port), getPodName(componentName), pfHelperContainer, outputHandler(port))
		if err != nil {
			return fmt.Errorf("error while creating port-forwarding for container port %d: %w", port.ContainerPort, err)
		}
		o.appliedPorts[port] = struct{}{}
	}
	return nil
}

func (o *PFClient) StopPortForwarding(componentName string) {
	if len(o.appliedPorts) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(o.appliedPorts))
	for port := range o.appliedPorts {
		port := port
		go func() {
			defer wg.Done()
			err := o.remoteProcessHandler.StopProcessForCommand(getCommandDefinition(port), getPodName(componentName), pfHelperContainer)
			if err != nil {
				klog.V(4).Infof("error while stopping port-forwarding for container port %d: %v", port.ContainerPort, err)
			}
		}()
	}
	wg.Wait()

	o.appliedPorts = make(map[api.ForwardedPort]struct{})
}

func (o *PFClient) GetForwardedPorts() map[string][]v1alpha2.Endpoint {
	result := make(map[string][]v1alpha2.Endpoint)
	for port := range o.appliedPorts {
		result[port.ContainerName] = append(result[port.ContainerName], v1alpha2.Endpoint{
			Name:       port.PortName,
			TargetPort: port.ContainerPort,
			Exposure:   v1alpha2.EndpointExposure(port.Exposure),
		})
	}
	return result
}

func getPodName(componentName string) string {
	return fmt.Sprintf("%s-app", componentName)
}

func getCommandDefinition(port api.ForwardedPort) remotecmd.CommandDefinition {
	proto := "tcp"
	switch {
	case strings.EqualFold(port.Protocol, string(corev1.ProtocolUDP)):
		proto = "udp"
	case strings.EqualFold(port.Protocol, string(corev1.ProtocolSCTP)):
		proto = "sctp"
	}
	return remotecmd.CommandDefinition{
		Id: fmt.Sprintf("pf-%s", port.PortName),
		// PidDirectory needs to be writable
		PidDirectory: "/projects/",
		CmdLine:      fmt.Sprintf("socat -d %[1]s-listen:%[2]d,reuseaddr,fork %[1]s:localhost:%[3]d", proto, port.LocalPort, port.ContainerPort),
	}
}
