package util

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/catalog"
)

// DisplayServices displays the specified services
func DisplayServices(services catalog.ServiceTypeList) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "PLANS")
	for _, service := range services.Items {
		fmt.Fprintln(w, service.ObjectMeta.Name, "\t", strings.Join(service.Spec.PlanList, ","))
	}
	w.Flush()
}

// DisplayComponents displays the specified  components
func DisplayComponents(components []string) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME")
	for _, component := range components {
		fmt.Fprintln(w, component)
	}
	w.Flush()
}

// FilterHiddenServices filters out services that should be hidden from the specified list
func FilterHiddenServices(input catalog.ServiceTypeList) catalog.ServiceTypeList {
	inputLength := len(input.Items)
	filteredServices := make([]catalog.ServiceType, 0, inputLength)

	for _, service := range input.Items {
		if !service.Spec.Hidden {
			filteredServices = append(filteredServices, service)
		}
	}

	return catalog.ServiceTypeList{
		TypeMeta:   input.TypeMeta,
		ObjectMeta: input.ObjectMeta,
		Items:      filteredServices,
	}
}

// FilterHiddenComponents filters out components that should be hidden from the specified list
func FilterHiddenComponents(input []catalog.ComponentType) []catalog.ComponentType {
	inputLength := len(input)
	filteredComponents := make([]catalog.ComponentType, 0, inputLength)
	for _, component := range input {
		// we keep the image if it has tags that are no hidden
		if len(component.Spec.NonHiddenTags) > 0 {
			filteredComponents = append(filteredComponents, component)
		}
	}
	return filteredComponents
}
