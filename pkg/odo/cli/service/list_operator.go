package service

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	applabels "github.com/openshift/odo/pkg/application/labels"
	cmplabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	svc "github.com/openshift/odo/pkg/service"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type clusterInfo struct {
	Labels            map[string]string
	CreationTimestamp time.Time
}

type serviceItem struct {
	ClusterInfo *clusterInfo
	InDevfile   bool
}

// listOperatorServices lists the operator backed services
// - deployed in the cluster
// - defined in the current devfile
func (o *ServiceListOptions) listOperatorServices() (err error) {

	// get the services deployed
	var clusterList []unstructured.Unstructured
	clusterList, failedListingCR, err := svc.ListOperatorServices(o.KClient)
	if err != nil {
		return err
	}

	// In JSON, only return services deployed
	if log.IsJSON() {
		if len(clusterList) == 0 {
			if len(failedListingCR) > 0 {
				fmt.Printf("Failed to fetch services for operator(s): %q\n\n", strings.Join(failedListingCR, ", "))
			}
			return fmt.Errorf("no operator backed services found in namespace: %s", o.KClient.Namespace)
		}

		machineoutput.OutputSuccess(clusterList)
		return nil
	}

	// get the services defined in the devfile
	// and the name of the component of the devfile
	var devfileList []string
	var devfileComponent string
	if o.EnvSpecificInfo != nil {
		devfileList, err = svc.ListDevfileServices(o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return fmt.Errorf("error reading devfile")
		}
		devfileComponent = o.EnvSpecificInfo.GetComponentSettings().Name
	}

	servicesItems := mixServices(clusterList, devfileList)

	if len(servicesItems) == 0 {
		if len(failedListingCR) > 0 {
			fmt.Printf("Failed to fetch services for operator(s): %q\n\n", strings.Join(failedListingCR, ", "))
		}
		return fmt.Errorf("no operator backed services found in namespace: %s", o.KClient.Namespace)
	}

	orderedNames := getOrderedServicesNames(servicesItems)

	// output result
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "MANAGED BY ODO", "\t", "STATE", "\t", "AGE")
	for _, name := range orderedNames {
		managedByOdo, state, duration := getTabularInfo(servicesItems[name], devfileComponent)
		fmt.Fprintln(w, strings.Join([]string{name, managedByOdo, state, duration}, "\t"))
	}
	w.Flush()

	if len(failedListingCR) > 0 {
		fmt.Printf("\nFailed to fetch services for operator(s): %q\n", strings.Join(failedListingCR, ", "))
	}

	return nil
}

// mixServices returns a structure containing both the services in cluster and defined in devfile
func mixServices(clusterList []unstructured.Unstructured, devfileList []string) (servicesItems map[string]*serviceItem) {
	servicesItems = map[string]*serviceItem{}
	for _, item := range clusterList {
		name := strings.Join([]string{item.GetKind(), item.GetName()}, "/")
		if _, ok := servicesItems[name]; !ok {
			servicesItems[name] = &serviceItem{}
		}
		servicesItems[name].ClusterInfo = &clusterInfo{
			Labels:            item.GetLabels(),
			CreationTimestamp: item.GetCreationTimestamp().Time,
		}
	}

	for _, item := range devfileList {
		name := item
		if _, ok := servicesItems[name]; !ok {
			servicesItems[name] = &serviceItem{}
		}
		servicesItems[name].InDevfile = true
	}
	return

}

// getOrderedServicesNames returns the names of the services ordered in alphabetic order
func getOrderedServicesNames(servicesItems map[string]*serviceItem) (orderedNames []string) {
	orderedNames = make([]string, len(servicesItems))
	i := 0
	for name := range servicesItems {
		orderedNames[i] = name
		i++
	}
	sort.Strings(orderedNames)
	return
}

// getTabularInfo returns information to be displayed in the output for a specific service and a specific current devfile component
func getTabularInfo(serviceItem *serviceItem, devfileComponent string) (managedByOdo, state, duration string) {
	clusterItem := serviceItem.ClusterInfo
	inDevfile := serviceItem.InDevfile
	if clusterItem != nil {
		// service deployed into cluster
		var component string
		labels := clusterItem.Labels
		isManagedByOdo := labels[applabels.OdoManagedBy] == "odo"
		if isManagedByOdo {
			component = labels[cmplabels.ComponentLabel]
			managedByOdo = fmt.Sprintf("Yes (%s)", component)
		} else {
			managedByOdo = "No"
		}
		duration = time.Since(clusterItem.CreationTimestamp).Truncate(time.Second).String()
		if inDevfile {
			// service deployed into cluster and defined in devfile
			state = "Pushed"
		} else {
			// service deployed into cluster and not defined in devfile
			if isManagedByOdo {
				if devfileComponent == component {
					state = "Deleted locally"
				} else {
					state = ""
				}
			} else {
				state = ""
			}
		}
	} else {
		if inDevfile {
			// service not deployed into cluster and defined in devfile
			state = "Not pushed"
			managedByOdo = fmt.Sprintf("Yes (%s)", devfileComponent)
		} else {
			// service not deployed into cluster and not defined in devfile
			// should not happen
			state = "Err!"
			managedByOdo = "Err!"
		}
	}
	return
}
