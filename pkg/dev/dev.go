package dev

import (
	"fmt"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-developer/odo/pkg/envinfo"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

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
	klog.V(4).Infoln("Successfully created inner-loop resources")
	return nil
}

func (o *DevClient) Cleanup() error {
	var err error
	return err
}

func (o *DevClient) SetupPortForwarding(pod *corev1.Pod, portPairs []string, errOut io.Writer) error {
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
	err = fw.ForwardPorts()
	if err != nil {
		fmt.Fprint(errOut, fmt.Errorf("error setting up port forwarding: %v", err).Error())
		// do cleanup when this happens
		// TODO: #5485
	}

	return nil
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
		InitialDevfileObj:   devfileObj,
	}

	return o.watchClient.WatchAndPush(out, watchParameters)
}
