package kubedev

import (
	"context"
	"fmt"
	"io"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/redhat-developer/odo/pkg/labels"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
)

func (o *DevClient) CleanupResources(ctx context.Context, out io.Writer) error {
	var (
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
	)
	fmt.Fprintln(out, "Cleaning resources, please wait")
	appname := odocontext.GetApplication(ctx)
	isInnerLoopDeployed, resources, err := o.deleteClient.ListClusterResourcesToDeleteFromDevfile(*devfileObj, appname, componentName, labels.ComponentDevMode)
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
		err = o.deleteClient.ExecutePreStopEvents(ctx, *devfileObj, appname, componentName)
		if err != nil {
			fmt.Fprint(out, "Failed to execute preStop events")
		}
	}
	// delete all the resources
	failed := o.deleteClient.DeleteResources(resources, true)
	if len(failed) == 0 {
		return nil
	}
	var list []string
	for _, fail := range failed {
		list = append(list, fmt.Sprintf("- %s/%s", fail.GetKind(), fail.GetName()))
	}

	return fmt.Errorf("could not delete the following resource(s): \n%v", strings.Join(list, "\n"))
}
