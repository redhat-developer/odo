package service

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	cmplabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	svc "github.com/redhat-developer/odo/pkg/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const ServiceItemKind = "Service"

type clusterInfo struct {
	Labels            map[string]string `json:"labels"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

type serviceItem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	ClusterInfo       *clusterInfo           `json:"clusterInfo,omitempty"`
	InDevfile         bool                   `json:"inDevfile"`
	Deployed          bool                   `json:"deployed"`
	Manifest          map[string]interface{} `json:"manifest"`
}

type serviceItemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []serviceItem `json:"items"`
}

func NewServiceItem(name string) *serviceItem {
	return &serviceItem{
		TypeMeta: metav1.TypeMeta{
			Kind:       ServiceItemKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
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

	// get the services defined in the devfile
	// and the name of the component of the devfile
	var devfileList map[string]unstructured.Unstructured
	var devfileComponent string
	if o.EnvSpecificInfo != nil {
		devfileList, err = svc.ListDevfileServices(o.KClient, o.EnvSpecificInfo.GetDevfileObj(), o.componentContext)
		if err != nil {
			return fmt.Errorf("error reading devfile")
		}
		devfileComponent = o.EnvSpecificInfo.GetComponentSettings().Name
	}

	servicesItems := mixServices(clusterList, devfileList)

	if len(servicesItems.Items) == 0 {
		if len(failedListingCR) > 0 {
			fmt.Printf("Failed to fetch services for operator(s): %q\n\n", strings.Join(failedListingCR, ", "))
		}
		return fmt.Errorf("no operator backed services found in namespace: %s", o.KClient.GetCurrentNamespace())
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(servicesItems)
		return nil
	}

	// output result
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "MANAGED BY ODO", "\t", "STATE", "\t", "AGE")
	for i := range servicesItems.Items {
		item := servicesItems.Items[i]
		managedByOdo, state, duration := getTabularInfo(&item, devfileComponent)
		fmt.Fprintln(w, item.Name, "\t", managedByOdo, "\t", state, "\t", duration)
	}
	w.Flush()

	if len(failedListingCR) > 0 {
		fmt.Printf("\nFailed to fetch services for operator(s): %q\n", strings.Join(failedListingCR, ", "))
	}

	return nil
}

// mixServices returns a structure containing both the services in cluster and defined in devfile
func mixServices(clusterList []unstructured.Unstructured, devfileList map[string]unstructured.Unstructured) serviceItemList {
	servicesItems := map[string]*serviceItem{}
	for _, item := range clusterList {
		if item.GetKind() == "ServiceBinding" {
			continue
		}
		name := strings.Join([]string{item.GetKind(), item.GetName()}, "/")
		if _, ok := servicesItems[name]; !ok {
			servicesItems[name] = NewServiceItem(name)
		}
		servicesItems[name].Manifest = item.Object
		servicesItems[name].Deployed = true
		servicesItems[name].ClusterInfo = &clusterInfo{
			Labels:            item.GetLabels(),
			CreationTimestamp: item.GetCreationTimestamp().Time,
		}
	}

	for name, manifest := range devfileList {
		if manifest.GetKind() == "ServiceBinding" {
			continue
		}
		if _, ok := servicesItems[name]; !ok {
			servicesItems[name] = NewServiceItem(name)
		}
		servicesItems[name].InDevfile = true
		if !servicesItems[name].Deployed {
			servicesItems[name].Manifest = manifest.Object
		}
	}

	return serviceItemList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: machineoutput.APIVersion,
		},
		Items: getOrderedServices(servicesItems),
	}
}

// getOrderedServices returns the services as a slice, ordered by name
func getOrderedServices(items map[string]*serviceItem) []serviceItem {
	orderedNames := getOrderedServicesNames(items)
	result := make([]serviceItem, len(items))
	i := 0
	for _, name := range orderedNames {
		result[i] = *items[name]
		i++
	}
	return result
}

// getOrderedServicesNames returns the names of the services ordered in alphabetic order
func getOrderedServicesNames(items map[string]*serviceItem) []string {
	orderedNames := make([]string, len(items))
	i := 0
	for name := range items {
		orderedNames[i] = name
		i++
	}
	sort.Strings(orderedNames)
	return orderedNames
}

// getTabularInfo returns information to be displayed in the output for a specific service and a specific current devfile component
func getTabularInfo(serviceItem *serviceItem, devfileComponent string) (managedByOdo, state, duration string) {
	clusterItem := serviceItem.ClusterInfo
	inDevfile := serviceItem.InDevfile
	if clusterItem != nil {
		// service deployed into cluster
		var component string
		labels := clusterItem.Labels
		isManagedByOdo := labels[applabels.ManagedBy] == "odo"
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
