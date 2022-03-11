package dev

import (
	"fmt"
	"io"
	"net/http"

	"github.com/redhat-developer/odo/pkg/envinfo"

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
	"github.com/redhat-developer/odo/pkg/kclient"
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

// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container.
// It uses envSpecificInfo to create push parameters and subsequently push the component to the cluster
func (o *DevClient) Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, path string) error {
	klog.V(4).Infoln("Creating new adapter")
	adapter, err := adapters.NewComponentAdapter(devfileObj.GetMetadataName(), path, "app", devfileObj, platformContext)
	if err != nil {
		return err
	}

	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}

	pushParameters := common.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		DebugPort:       envSpecificInfo.GetDebugPort(),
		Path:            path,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	err = adapter.Push(pushParameters)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resourcs")
	return nil
}

// Cleanup cleans the resources created by Push
func (o *DevClient) Cleanup() error {
	var err error
	return err
}

// SetupPortForwarding sets up port forwarding for the endpoints in the devfile
func (o *DevClient) SetupPortForwarding(devfileObj parser.DevfileObj, path string, out io.Writer, errOut io.Writer) (map[string]string, error) {
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
		return pPairs, nil
	}

	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return pPairs, err
	}
	pod, err := o.kubernetesClient.GetOnePodFromSelector(componentlabels.GetSelector(devfileObj.GetMetadataName(), envSpecificInfo.GetApplication()))
	if err != nil {
		return pPairs, err
	}

	transport, upgrader, err := spdy.RoundTripperFor(o.kubernetesClient.GetClientConfig())
	if err != nil {
		return pPairs, err
	}

	req := o.kubernetesClient.GeneratePortForwardReq(pod.Name)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	stopChan := make(chan struct{}, 1)
	// passing nil for readyChan because it's eventually being closed if it's not nil
	// passing nil for out because we only care for error, not for output messages; we want to print our own messages
	fw, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, portPairs, stopChan, nil, nil, errOut)
	if err != nil {
		return pPairs, err
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

	return pPairs, nil
}

func (o *DevClient) Watch(devfileObj parser.DevfileObj, path string, ignorePaths []string, out io.Writer, h Handler) error {
	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
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
