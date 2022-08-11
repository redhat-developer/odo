package portForward

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/watch"
)

var _ Client = (*PFClient)(nil)

type PFClient struct {
	kubernetesClient kclient.ClientInterface
	stateClient      state.Client

	appliedEndpoints map[string][]int

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

func (o *PFClient) StartPortForwarding(
	devFileObj parser.DevfileObj,
	componentName string,
	randomPorts bool,
	errOut io.Writer,
) error {

	ceMapping, err := o.GetPortsToForward(devFileObj)
	if err != nil {
		return err
	}

	if o.stopChan != nil && reflect.DeepEqual(ceMapping, o.appliedEndpoints) {
		return nil
	}

	o.appliedEndpoints = ceMapping

	o.StopPortForwarding()

	o.stopChan = make(chan struct{}, 1)

	var portPairs map[string][]string
	if randomPorts {
		portPairs = randomPortPairsFromContainerEndpoints(ceMapping)
	} else {
		portPairs = portPairsFromContainerEndpoints(ceMapping)
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

func (o *PFClient) StopPortForwarding() {
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

func (o *PFClient) GetForwardedPorts() map[string][]int {
	return o.appliedEndpoints
}

func (o *PFClient) GetPortsToForward(devFileObj parser.DevfileObj) (map[string][]int, error) {

	// get the endpoint/port information for containers in devfile
	containers, err := devFileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType},
	})
	if err != nil {
		return nil, err
	}
	ceMapping := libdevfile.GetContainerEndpointMapping(containers)
	return ceMapping, nil
}

// randomPortPairsFromContainerEndpoints assigns a random (empty) port on localhost to each port in the provided containerEndpoints map
// it returns a map of the format "<container-name>":{"<local-port-1>:<remote-port-1>", "<local-port-2>:<remote-port-2>"}
// "container1": {":3000", ":3001"}
func randomPortPairsFromContainerEndpoints(ceMap map[string][]int) map[string][]string {
	portPairs := make(map[string][]string)

	for name, ports := range ceMap {
		for _, p := range ports {
			pair := fmt.Sprintf(":%d", p)
			portPairs[name] = append(portPairs[name], pair)
		}
	}
	return portPairs
}

// portPairsFromContainerEndpoints assigns a port on localhost to each port in the provided containerEndpoints map
// it returns a map of the format "<container-name>":{"<local-port-1>:<remote-port-1>", "<local-port-2>:<remote-port-2>"}
// "container1": {"400001:3000", "400002:3001"}
func portPairsFromContainerEndpoints(ceMap map[string][]int) map[string][]string {
	portPairs := make(map[string][]string)
	port := 40000

	for name, ports := range ceMap {
		for _, p := range ports {
			port++
			for {
				isPortFree := util.IsPortFree(port)
				if isPortFree {
					pair := fmt.Sprintf("%d:%d", port, p)
					portPairs[name] = append(portPairs[name], pair)
					break
				}
				port++
			}
		}
	}
	return portPairs
}
