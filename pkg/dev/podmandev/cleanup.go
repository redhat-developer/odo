package podmandev

import (
	"context"
	"fmt"
	"io"
)

func (o *DevClient) CleanupResources(ctx context.Context, out io.Writer) error {
	fmt.Printf("Cleaning up resources\n")
	if o.deployedPod == nil {
		return nil
	}
	return o.podmanClient.CleanupPodResources(o.deployedPod)
}
