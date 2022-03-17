package dev

import (
	"io"

	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"

	"github.com/redhat-developer/odo/pkg/watch"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
)

type Client interface {
	// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container.
	Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, path string) error

	// SetupPortForwarding sets up port forwarding to the pod for the portPairs provided in
	// ["<local-port-1>":"<remote-port-1>", "<local-port-2>":"<remote-port-2>"] format.
	// It fetches the pod information using the devfileObj. It uses errOut to print errors while performing the port-forwarding.
	SetupPortForwarding(pod *corev1.Pod, portPairs []string, devfileObj parser.DevfileObj, errOut io.Writer) error

	// Watch watches for any changes to the files under path while ignoring the files/directories in ignorePaths.
	// It logs messages to out and uses the Handler h to perform push operation when anything changes in path.
	// It uses devfileObj to notify user to restart odo dev if they change endpoint information in the devfile.
	Watch(devfileObj parser.DevfileObj, path string, ignorePaths []string, out io.Writer, h Handler) error

	// Cleanup cleans the resources created by Start
	Cleanup() error
}

type Handler interface {
	RegenerateAdapterAndPush(common.PushParameters, watch.WatchParameters) error
}
