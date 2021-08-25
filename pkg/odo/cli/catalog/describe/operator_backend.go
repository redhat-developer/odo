package describe

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/olekukonko/tablewriter"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type operatorBackend struct {
	Name           string
	OperatorType   string
	CustomResource string
	CSV            olm.ClusterServiceVersion
	CR             *olm.CRDDescription
	CRDSpec        *spec.Schema
}

func NewOperatorBackend() *operatorBackend {
	return &operatorBackend{}
}

func (ohb *operatorBackend) CompleteDescribeService(dso *DescribeServiceOptions, args []string) error {
	ohb.Name = args[0]
	oprType, CR, err := service.SplitServiceKindName(ohb.Name)
	if err != nil {
		return err
	}
	// we check if the cluster supports ClusterServiceVersion or not.
	isCSVSupported, err := service.IsCSVSupported()
	if err != nil {
		// if there is an error checking it, we return the error.
		return err
	}
	// if its not supported then we return an error
	if !isCSVSupported {
		return errors.New("it seems the cluster doesn't support Operators. Please install OLM and try again")
	}
	ohb.OperatorType = oprType
	ohb.CustomResource = CR
	return nil
}

func (ohb *operatorBackend) ValidateDescribeService(dso *DescribeServiceOptions) error {
	var err error
	if ohb.OperatorType == "" || ohb.CustomResource == "" {
		return errors.New("invalid service name, use the format <operator-type>/<crd-name>")
	}
	// make sure that CSV of the specified OperatorType exists
	ohb.CSV, err = dso.KClient.GetClusterServiceVersion(ohb.OperatorType)
	if err != nil {
		// error only occurs when OperatorHub is not installed.
		// k8s does't have it installed by default but OCP does
		return err
	}

	// Get the specific CR that matches "kind"
	crs := dso.KClient.GetCustomResourcesFromCSV(&ohb.CSV)

	var cr *olm.CRDDescription
	hasCR := false
	for _, custRes := range *crs {
		c := custRes
		if c.Kind == ohb.CustomResource {
			cr = &c
			hasCR = true
			break
		}
	}
	if !hasCR {
		return fmt.Errorf("the %q resource doesn't exist in specified %q operator", ohb.CustomResource, ohb.OperatorType)
	}

	ohb.CR = cr

	crd, err := dso.KClient.GetResourceSpecDefinition(ohb.CR.Name, ohb.CR.Version, ohb.CustomResource)
	if err != nil {
		log.Warning("Unable to get CRD specifications:", err)
		return nil
	}
	ohb.CRDSpec = crd

	if crd == nil {
		ohb.CRDSpec = toOpenAPISpec(cr)
	}
	return nil

}

func (ohb *operatorBackend) RunDescribeService(dso *DescribeServiceOptions) error {

	if dso.isExample {
		almExample, err := service.GetAlmExample(ohb.CSV, ohb.CustomResource, ohb.OperatorType)
		if err != nil {
			return err
		}
		if log.IsJSON() {
			jsonExample := service.NewOperatorExample(almExample)
			jsonCR, err := json.MarshalIndent(jsonExample, "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(string(jsonCR))

		} else {
			yamlCR, err := yaml.Marshal(almExample)
			if err != nil {
				return err
			}

			log.Info(string(yamlCR))
		}
		return nil
	}

	svc := service.NewOperatorBackedService(ohb.Name, ohb.CR.Kind, ohb.CR.Version, ohb.CR.Description, ohb.CR.DisplayName, ohb.CRDSpec)
	if log.IsJSON() {
		machineoutput.OutputSuccess(svc)
	} else {
		HumanReadableOutput(os.Stdout, svc)
	}
	return nil
}

func HumanReadableOutput(w io.Writer, service service.OperatorBackedService) {
	fmt.Fprintf(w, "Kind: %s\n", service.Spec.Kind)
	fmt.Fprintf(w, "Version: %s\n", service.Spec.Version)
	fmt.Fprintf(w, "Description: %s\n", service.Spec.Description)
	fmt.Fprintln(w, "Parameters:")

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Path", "DisplayName", "Description"})
	displayProperties(table, service.Spec.Schema, "")
	table.Render()
	fmt.Fprint(w, tableString)
}

// displayProperties displays the properties of an OpenAPI schema in a human readable form
func displayProperties(table *tablewriter.Table, schema *spec.Schema, prefix string) {
	keys := make([]string, len(schema.Properties))
	i := 0
	for key := range schema.Properties {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		property := schema.Properties[key]
		if property.Type.Contains("object") {
			displayProperties(table, &property, prefix+key+".")
		} else {
			table.Append([]string{prefix + key, property.Title, property.Description})
		}
	}
}

// toOpenAPISpec transforms Spec descriptors from a CRD description to an OpenAPI schema
func toOpenAPISpec(repr *olm.CRDDescription) *spec.Schema {
	schema := new(spec.Schema).Typed("object", "")
	for _, param := range repr.SpecDescriptors {
		addParam(schema, param)
	}
	return schema
}

// addParam adds a Spec Descriptor parameter to an OpenAPI schema
func addParam(schema *spec.Schema, param olm.SpecDescriptor) {
	parts := strings.SplitN(param.Path, ".", 2)
	if len(parts) == 1 {
		child := spec.StringProperty().WithTitle(param.DisplayName).WithDescription(param.Description)
		schema.SetProperty(parts[0], *child)
	} else {
		var child *spec.Schema
		if _, ok := schema.Properties[parts[0]]; ok {
			c := schema.Properties[parts[0]]
			child = &c
		} else {
			child = new(spec.Schema).Typed("object", "")
		}
		param.Path = parts[1]
		addParam(child, param)
		schema.SetProperty(parts[0], *child)
	}
}
