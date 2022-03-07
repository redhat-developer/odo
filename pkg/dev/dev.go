package dev

import (
	"fmt"
	"io"
	"net/http"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/watch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
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

	var endpoints []v1.Endpoint
	for i := range containers {
		for _, e := range containers[i].Container.Endpoints {
			if e.Exposure != v1.NoneEndpointExposure {
				endpoints = append(endpoints, e)
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
	fw, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, []string{"8080:3000"}, stopChan, readyChan, out, errOut)
	if err != nil {
		return err
	}

	err = fw.ForwardPorts()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "\nStarted port forwarding.\n")
	log.Finfof(out, "\nYour application is now running on your cluster.")

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
