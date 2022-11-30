package backend

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"k8s.io/klog"
)

// handleApplicationPorts updates the ports in the Devfile as needed.
// If there are multiple container components in the Devfile, nothing is done. This will be handled in https://github.com/redhat-developer/odo/issues/6264.
// Otherwise, all the container component endpoints/ports (other than Debug) are updated with the specified ports.
func handleApplicationPorts(w io.Writer, devfileobj parser.DevfileObj, ports []int) (parser.DevfileObj, error) {
	if len(ports) == 0 {
		return devfileobj, nil
	}

	components, err := devfileobj.Data.GetDevfileContainerComponents(parsercommon.DevfileOptions{})
	if err != nil {
		return parser.DevfileObj{}, err
	}
	nbContainerComponents := len(components)
	klog.V(3).Infof("Found %d container components in Devfile at path %q", nbContainerComponents, devfileobj.Ctx.GetAbsPath())
	if nbContainerComponents == 0 {
		// no container components => nothing to do
		return devfileobj, nil
	}
	if nbContainerComponents > 1 {
		klog.V(3).Infof("found more than 1 container components in Devfile at path %q => cannot find out which component needs to be updated."+
			"This case will be handled in https://github.com/redhat-developer/odo/issues/6264", devfileobj.Ctx.GetAbsPath())
		fmt.Fprintln(w, "\nApplication ports detected but the current Devfile contains multiple container components. Could not determine which component to update. "+
			"Please feel free to customize the Devfile configuration below.")
		return devfileobj, nil
	}

	component := components[0]

	//Remove all but Debug endpoints
	var portsToRemove []string
	for _, ep := range component.Container.Endpoints {
		if ep.Name == "debug" || strings.HasPrefix(ep.Name, "debug-") {
			continue
		}
		portsToRemove = append(portsToRemove, strconv.Itoa(ep.TargetPort))
	}
	err = devfileobj.Data.RemovePorts(map[string][]string{component.Name: portsToRemove})
	if err != nil {
		return parser.DevfileObj{}, err
	}

	portsToSet := make([]string, 0, len(ports))
	for _, p := range ports {
		portsToSet = append(portsToSet, strconv.Itoa(p))
	}
	err = devfileobj.Data.SetPorts(map[string][]string{component.Name: portsToSet})
	if err != nil {
		return parser.DevfileObj{}, err
	}

	return devfileobj, err
}
