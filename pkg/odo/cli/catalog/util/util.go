package util

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/log"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// DisplayServices displays the specified services
func DisplayServices(services catalog.ServiceTypeList) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w) // this newline helps when cluster has both Operator and Service Catalog enabled
	log.Info("Services available through Service Catalog")
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

// DisplayClusterServiceVersions displays installed Operators in a human friendly manner
func DisplayClusterServiceVersions(csvs *olm.ClusterServiceVersionList) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	log.Info("Services available through Operators")
	fmt.Fprintln(w, "NAME", "\t", "CRDs")
	for _, csv := range csvs.Items {
		fmt.Fprintln(w, csv.ObjectMeta.Name, "\t", CsvOperators(csv.Spec.CustomResourceDefinitions))
	}
	w.Flush()
}

// CsvOperators returns a string contains all the Kind from the input crds
func CsvOperators(crds olm.CustomResourceDefinitions) string {
	var crdsSlice []string
	for _, crd := range crds.Owned {
		crdsSlice = append(crdsSlice, crd.Kind)
	}
	return strings.Join(crdsSlice, ", ")
}
