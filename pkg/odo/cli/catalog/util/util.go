package util

import (
	"fmt"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/occlient"
	"os"
	"strings"
	"text/tabwriter"
)

// DisplayServices displays the specified services
func DisplayServices(services []occlient.Service) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "PLANS")
	for _, service := range services {
		fmt.Fprintln(w, service.Name, "\t", strings.Join(service.PlanList, ","))
	}
	w.Flush()
}

// FilterHiddenServices filters out services that should be hidden from the specified list
func FilterHiddenServices(input []occlient.Service) []occlient.Service {
	inputLength := len(input)
	filteredServices := make([]occlient.Service, 0, inputLength)
	for _, service := range input {
		if !service.Hidden {
			filteredServices = append(filteredServices, service)
		}
	}
	return filteredServices
}

// FilterHiddenComponents filters out components that should be hidden from the specified list
func FilterHiddenComponents(input []catalog.CatalogImage) []catalog.CatalogImage {
	inputLength := len(input)
	filteredComponents := make([]catalog.CatalogImage, 0, inputLength)
	for _, component := range input {
		// we keep the image if it has tags that are no hidden
		if len(component.NonHiddenTags) > 0 {
			filteredComponents = append(filteredComponents, component)
		}
	}
	return filteredComponents
}
