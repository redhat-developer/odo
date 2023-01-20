package podmandev

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"

	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/util"
)

func (o *DevClient) CleanupResources(ctx context.Context, out io.Writer) error {
	fmt.Printf("Cleaning up resources\n")
	if o.deployedPod == nil {
		compName := odocontext.GetComponentName(ctx)
		appName := odocontext.GetApplication(ctx)
		name, err := util.NamespaceKubernetesObject(compName, appName)
		if err != nil {
			return nil
		}
		o.deployedPod = &corev1.Pod{}
		o.deployedPod.SetName(name)
	}
	return o.podmanClient.CleanupPodResources(o.deployedPod)
}
