package kubeportforward

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/watch"
)

var _ portForward.Client = (*PFClient)(nil)

type PFClient struct {
	kubernetesClient kclient.ClientInterface
	stateClient      state.Client

	appliedEndpoints map[string][]v1alpha2.Endpoint

	// stopChan on which to write to stop the port forwarding
	stopChan chan struct{}
	// finishedChan is written when the port forwarding is finished
	finishedChan chan struct{}

	originalErrorHandlers []func(error)

	// indicates that the port forwarding is started, and not stopped
	isRunning bool
}

func NewPFClient(kubernetesClient kclient.ClientInterface, stateClient state.Client) *PFClient {
	return &PFClient{
		kubernetesClient: kubernetesClient,
		stateClient:      stateClient,
	}
}

func (o *PFClient) StartPortForwarding(devFileObj parser.DevfileObj, componentName string, debug bool, randomPorts bool, out io.Writer, errOut io.Writer, definedPorts []api.ForwardedPort) error {

	ceMapping, err := libdevfile.GetDevfileContainerEndpointMapping(devFileObj, debug)
	if err != nil {
		return err
	}

	if o.stopChan != nil && reflect.DeepEqual(ceMapping, o.appliedEndpoints) {
		return nil
	}

	o.appliedEndpoints = ceMapping

	o.StopPortForwarding(componentName)

	if len(ceMapping) == 0 {
		klog.V(4).Infof("no endpoint declared in the component, no ports are forwarded")
		return nil
	}

	o.stopChan = make(chan struct{}, 1)

	var portPairs map[string][]string
	if len(definedPorts) != 0 {
		portPairs = make(map[string][]string)
		for _, port := range definedPorts {
			s := fmt.Sprintf("%d:%d", port.LocalPort, port.ContainerPort)
			if port.LocalPort == 0 {
				// random port
				s = fmt.Sprintf(":%d", port.ContainerPort)
			}
			portPairs[port.ContainerName] = append(portPairs[port.ContainerName], s)
		}
	} else {
		if randomPorts {
			portPairs = randomPortPairsFromContainerEndpoints(ceMapping)
		} else {
			portPairs = portPairsFromContainerEndpoints(ceMapping)
		}
	}
	var portPairsSlice []string
	for _, v1 := range portPairs {
		portPairsSlice = append(portPairsSlice, v1...)
	}
	pod, err := o.kubernetesClient.GetPodUsingComponentName(componentName)
	if err != nil {
		return err
	}

	o.originalErrorHandlers = append([]func(error){}, runtime.ErrorHandlers...)

	runtime.ErrorHandlers = append(runtime.ErrorHandlers, func(err error) {
		if err.Error() == "lost connection to pod" {
			// Stop the low-level port forwarding
			// the infinite loop will restart it
			if o.stopChan == nil {
				return
			}
			o.stopChan <- struct{}{}
			o.stopChan = make(chan struct{}, 1)
		}
	})

	o.isRunning = true

	devstateChan := make(chan error)
	go func() {
		backo := watch.NewExpBackoff()
		for {
			o.finishedChan = make(chan struct{}, 1)
			portsBuf := NewPortWriter(log.GetStdout(), len(portPairsSlice), ceMapping)

			go func() {
				portsBuf.Wait()
				err = o.stateClient.SetForwardedPorts(portsBuf.GetForwardedPorts())
				if err != nil {
					err = fmt.Errorf("unable to save forwarded ports to state file: %v", err)
				}
				devstateChan <- err
			}()

			err = o.kubernetesClient.SetupPortForwarding(pod, portPairsSlice, portsBuf, errOut, o.stopChan)
			if err != nil {
				fmt.Fprintf(errOut, "Failed to setup port-forwarding: %v\n", err)
				d := backo.Delay()
				time.Sleep(d)
			} else {
				backo.Reset()
			}
			if !o.isRunning {
				break
			}
		}
		o.finishedChan <- struct{}{}
	}()

	// Wait the first time the devstate file is written
	timeout := 1 * time.Minute
	select {
	case err = <-devstateChan:
		return err
	case <-time.After(timeout):
		return errors.New("unable to setup port forwarding")
	}
}

func (o *PFClient) StopPortForwarding(componentName string) {
	if o.stopChan == nil {
		return
	}
	// Ask the low-level port forward to stop
	o.stopChan <- struct{}{}
	o.stopChan = nil

	// Ask the infinite loop to stop
	o.isRunning = false

	// Wait for low level port forward to be finished
	// and the infinite loop to exit
	<-o.finishedChan
	o.finishedChan = nil
	runtime.ErrorHandlers = o.originalErrorHandlers
}

func (o *PFClient) GetForwardedPorts() map[string][]v1alpha2.Endpoint {
	return o.appliedEndpoints
}

// randomPortPairsFromContainerEndpoints assigns a random (empty) port on localhost to each port in the provided containerEndpoints map
// it returns a map of the format "<container-name>":{"<local-port-1>:<remote-port-1>", "<local-port-2>:<remote-port-2>"}
// "container1": {":3000", ":3001"}
func randomPortPairsFromContainerEndpoints(ceMap map[string][]v1alpha2.Endpoint) map[string][]string {
	portPairs := make(map[string][]string)

	for name, ports := range ceMap {
		for _, p := range ports {
			pair := fmt.Sprintf(":%d", p.TargetPort)
			portPairs[name] = append(portPairs[name], pair)
		}
	}
	return portPairs
}

// portPairsFromContainerEndpoints assigns a port on localhost to each port in the provided containerEndpoints map
// it returns a map of the format "<container-name>":{"<local-port-1>:<remote-port-1>", "<local-port-2>:<remote-port-2>"}
// "container1": {"20001:3000", "20002:3001"}
func portPairsFromContainerEndpoints(ceMap map[string][]v1alpha2.Endpoint) map[string][]string {
	portPairs := make(map[string][]string)
	startPort := 20001
	endPort := startPort + 10000
	for name, ports := range ceMap {
		for _, p := range ports {
			freePort, err := util.NextFreePort(startPort, endPort, nil)
			if err != nil {
				klog.Infof("%s", err)
				continue
			}
			pair := fmt.Sprintf("%d:%d", freePort, p.TargetPort)
			portPairs[name] = append(portPairs[name], pair)
			startPort = freePort + 1
		}
	}
	return portPairs
}
