package dev

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/redhat-developer/odo/pkg/util"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/watch"
	"k8s.io/klog/v2"
)

// this causes compilation to fail if DevClient struct doesn't implement Client interface
var _ Client = (*DevClient)(nil)

type DevClient struct {
	watchClient      watch.Client
	kubernetesClient kclient.ClientInterface
}

func NewDevClient(watchClient watch.Client, kubernetesClient kclient.ClientInterface) *DevClient {
	return &DevClient{
		watchClient:      watchClient,
		kubernetesClient: kubernetesClient,
	}
}

// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container,
// forwards remote port on the pod to a port on localhost using the rest config, and watches the component for changes.
// It prints all the logs/output to out.
func (o *DevClient) Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, ignorePaths []string, path string, out io.Writer, errOut io.Writer, h Handler) error {
	var err error

	var adapter common.ComponentAdapter
	klog.V(4).Infoln("Creating new adapter")
	adapter, err = adapters.NewComponentAdapter(devfileObj.GetMetadataName(), path, "app", devfileObj, platformContext)
	if err != nil {
		return err
	}

	var envSpecificInfo *envinfo.EnvSpecificInfo
	envSpecificInfo, err = envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}
	pushParameters := common.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		Path:            path,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	err = adapter.Push(pushParameters)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resourcs")

	// port forwarding for all endpoints in the devfileObj
	var containers []v1.Component
	containers, err = devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1.ContainerComponentType},
	})

	var portPairs []string
	port := 40000
	for i := range containers {
		for _, e := range containers[i].Container.Endpoints {
			if e.Exposure != v1.NoneEndpointExposure {
			loop:
				port++
				isPortFree := util.IsPortFree(port)
				if !isPortFree {
					goto loop
				}
				pair := fmt.Sprintf("%d:%d", port, e.TargetPort)
				portPairs = append(portPairs, pair)
			}
		}
	}

	var pod *corev1.Pod
	pod, err = o.kubernetesClient.GetOnePodFromSelector(componentlabels.GetSelector(devfileObj.GetMetadataName(), envSpecificInfo.GetApplication()))
	if err != nil {
		return err
	}

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	transport, upgrader, err := spdy.RoundTripperFor(o.kubernetesClient.GetClientConfig())
	if err != nil {
		return err
	}

	req := o.kubernetesClient.GeneratePortForwardReq(pod.Name)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	// passing nil in below call since we only care for error, not for output messages
	fw, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, portPairs, stopChan, readyChan, nil, errOut)
	if err != nil {
		return err
	}

	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			fmt.Fprintf(out, fmt.Errorf("error setting up port forwarding: %v", err).Error())
			os.Exit(1)
		}
	}()

	log.Finfof(out, "\nYour application is now running on your cluster.")

	err = o.SetupPortForwarding(devfileObj, envSpecificInfo, out, errOut)
	if err != nil {
		return err
	}

	watchParameters := watch.WatchParameters{
		Path:                path,
		ComponentName:       devfileObj.GetMetadataName(),
		ApplicationName:     "app",
		ExtChan:             make(chan bool),
		DevfileWatchHandler: h.RegenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         ignorePaths,
	}

	return o.watchClient.WatchAndPush(out, watchParameters)
}

// Cleanup cleans the resources created by Push
func (o *DevClient) Cleanup() error {
	var err error
	return err
}

// SetupPortForwarding sets up port forwarding for the endpoints in the devfile
func (o *DevClient) SetupPortForwarding(devfileObj parser.DevfileObj, envSpecificInfo *envinfo.EnvSpecificInfo, out, errOut io.Writer) error {
	containers, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1.ContainerComponentType},
	})

	ceMap := containerEndpointsFromContainers(containers)
	pPairs := portPairsFromContainerEndpoints(ceMap)

	var portPairs []string
	for i := range pPairs {
		portPairs = append(portPairs, i)
	}

	if len(portPairs) == 0 {
		// no endpoints with exposure set to public or internal; no ports to be forwarded
		return nil
	}

	pod, err := o.kubernetesClient.GetOnePodFromSelector(componentlabels.GetSelector(devfileObj.GetMetadataName(), envSpecificInfo.GetApplication()))
	if err != nil {
		return err
	}

	transport, upgrader, err := spdy.RoundTripperFor(o.kubernetesClient.GetClientConfig())
	if err != nil {
		return err
	}

	req := o.kubernetesClient.GeneratePortForwardReq(pod.Name)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	stopChan := make(chan struct{}, 1)
	// passing nil for readyChan because it's eventually being closed if it's not nil
	// passing nil for out because we only care for error, not for output messages; we want to print our own messages
	fw, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, portPairs, stopChan, nil, nil, errOut)
	if err != nil {
		return err
	}

	// start port-forwarding
	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			fmt.Fprint(out, fmt.Errorf("error setting up port forwarding: %v", err).Error())
			// do cleanup when this happens
			// TODO: #5485
		}
	}()

	printPortForwardingInfo(pPairs, out)
	return nil
}

func printPortForwardingInfo(portPairs map[string]string, out io.Writer) {
	portFowardURLs := ""
	for pair, container := range portPairs {
		split := strings.Split(pair, ":")
		local := split[0]
		remote := split[1]

		portFowardURLs += fmt.Sprintf("- Port %s from %q container forwarded to localhost:%s\n", remote, container, local)
	}
	fmt.Fprintf(out, "\n%s", portFowardURLs)
}

// containerEndpointsFromContainers returns a map of with container name as key and its endpoints as a slice of strings
// it considers only ports that don't have exposure status "None"
func containerEndpointsFromContainers(containers []v1alpha2.Component) map[string][]int {
	ceMap := make(map[string][]int, 0)

	for _, c := range containers {
		for _, ep := range c.Container.Endpoints {
			if ep.Exposure != v1.NoneEndpointExposure {
				port := ep.TargetPort
				if _, ok := ceMap[c.Name]; !ok {
					ceMap[c.Name] = []int{port}
					continue
				}
				ceMap[c.Name] = append(ceMap[c.Name], port)
			}
		}
	}

	return ceMap
}

// portPairsFromContainerEndpoints returns a map of the format "<local-port>:<remote-port>":"<container-name>"
func portPairsFromContainerEndpoints(ceMap map[string][]int) map[string]string {
	portPairs := make(map[string]string, 0)
	port := 40000

	for name, ports := range ceMap {
		for _, p := range ports {
			port++
			for {
				isPortFree := util.IsPortFree(port)
				if isPortFree {
					pair := fmt.Sprintf("%d:%d", port, p)
					portPairs[pair] = name
					break
				}
				port++
			}
		}
	}

	return portPairs
}
