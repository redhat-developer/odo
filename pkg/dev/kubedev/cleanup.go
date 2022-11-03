package kubedev

import (
	"context"
	"fmt"
	"io"

	"github.com/redhat-developer/odo/pkg/labels"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

func (o *DevClient) CleanupResources(ctx context.Context, out io.Writer) error {
	var (
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetDevfileObj(ctx)
	)
	fmt.Fprintln(out, "Cleaning resources, please wait")
	appname := odocontext.GetApplication(ctx)
	isInnerLoopDeployed, resources, err := o.deleteClient.ListResourcesToDeleteFromDevfile(*devfileObj, appname, componentName, labels.ComponentDevMode)
	if err != nil {
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			fmt.Fprintf(out, "Error connecting to the cluster, the resources were not cleaned up.\nPlease log in again and cleanup the resource with `odo delete component`\n\n")
		} else {
			fmt.Fprintf(out, "Failed to delete inner loop resources: %v\n", err)
		}
		return err
	}
	// if innerloop deployment resource is present, then execute preStop events
	if isInnerLoopDeployed {
		err = o.deleteClient.ExecutePreStopEvents(*devfileObj, appname, componentName)
		if err != nil {
			fmt.Fprint(out, "Failed to execute preStop events")
		}
	}
	// delete all the resources
	failed := o.deleteClient.DeleteResources(resources, true)
	for _, fail := range failed {
		fmt.Fprintf(out, "Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
	}

	return nil
}
