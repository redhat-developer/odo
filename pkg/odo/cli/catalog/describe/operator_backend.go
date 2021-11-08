package describe

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/go-openapi/spec"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

type operatorBackend struct {
	Name           string
	OperatorType   string
	CustomResource string
	CSV            olm.ClusterServiceVersion
	CR             *olm.CRDDescription
	CRDSpec        *spec.Schema
	CRDList        *service.OperatorBackedServiceCRList
}

func NewOperatorBackend() *operatorBackend {
	return &operatorBackend{}
}

func (ohb *operatorBackend) CompleteDescribeService(dso *DescribeServiceOptions, args []string) error {
	ohb.Name = args[0]
	oprType, CR, err := service.SplitServiceKindName(ohb.Name)
	if err != nil {
		klog.V(2).Infof("could not determine csv, falling back to describing all of them")
		oprType = args[0]
		CR = ""
	}
	// we check if the cluster supports ClusterServiceVersion or not.
	isCSVSupported, err := dso.Client.GetKubeClient().IsCSVSupported()
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
	if ohb.OperatorType == "" {
		return errors.New("invalid service name provided. should either be <operator-type> or <operator-type>/<crd-name>")
	}
	// make sure that CSV of the specified OperatorType exists
	ohb.CSV, err = dso.KClient.GetClusterServiceVersion(ohb.OperatorType)
	if err != nil {
		// error only occurs when OperatorHub is not installed.
		// k8s does't have it installed by default but OCP does
		return err
	}

	//if both operator type and cr are known, validate that it exists
	if ohb.OperatorType != "" && ohb.CustomResource != "" {
		var hasCR bool
		hasCR, ohb.CR = dso.KClient.CheckCustomResourceInCSV(ohb.CustomResource, &ohb.CSV)
		if !hasCR {
			return fmt.Errorf("the %q resource doesn't exist in specified %q operator", ohb.CustomResource, ohb.OperatorType)
		}

		ohb.CRDSpec, err = dso.KClient.GetCRDSpec(ohb.CR, ohb.OperatorType, ohb.CustomResource)
		if err != nil {
			return err
		}
	}
	return nil

}

func (ohb *operatorBackend) RunDescribeService(dso *DescribeServiceOptions) error {
	if ohb.OperatorType != "" && ohb.CustomResource == "" {
		//we don't have cr so list all possible crds
		ohb.CRDList = service.NewOperatorBackedCRList(ohb.OperatorType, ohb.CSV.Spec.DisplayName, ohb.CSV.Spec.Description)
		crds := *dso.KClient.GetCustomResourcesFromCSV(&ohb.CSV)
		for _, custRes := range crds {
			ohb.CRDList.Spec.CRDS = append(ohb.CRDList.Spec.CRDS, service.OperatorServiceCRItem{
				Kind:        custRes.Kind,
				Description: custRes.Description,
			})
		}
		if log.IsJSON() {
			machineoutput.OutputSuccess(ohb.CRDList)
		} else {
			HumanReadableCRListOutput(os.Stdout, ohb.CRDList)
		}
		return nil
	}
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
	fmt.Fprintf(w, "KIND:    %s\n", service.Spec.Kind)
	fmt.Fprintf(w, "VERSION: %s\n", service.Spec.Version)
	fmt.Fprintf(w, "\nDESCRIPTION:\n%s", indentText(service.Spec.Description, 5))

	if service.Spec.Schema == nil {
		log.Warningf("Unable to get parameters from CRD or CSV; Operator %q doesn't have the required information", service.Name)
		return
	}
	fmt.Fprintln(w, "\nFIELDS:")
	displayProperties(w, service.Spec.Schema, "")
}

func HumanReadableCRListOutput(w io.Writer, crsList *service.OperatorBackedServiceCRList) {
	fmt.Fprintf(w, "NAME:\t%s\n", crsList.Name)
	descriptionLines := strings.ReplaceAll(crsList.Spec.Description, "\n", "\n\t")
	fmt.Fprintf(w, "DESCRIPTION:\n\n\t%s\n\n", descriptionLines)
	fmt.Fprintf(w, "CRDs:\n")
	tw := tabwriter.NewWriter(w, 4, 4, 3, ' ', tabwriter.TabIndent)
	defer tw.Flush()
	fmt.Fprintf(tw, "\tNAME\tDESCRIPTION\n")
	for _, it := range crsList.Spec.CRDS {
		fmt.Fprintf(tw, "\t%s\t%s\n", it.Kind, it.Description)
	}
}

// displayProperties displays the properties of an OpenAPI schema in a human readable form
// required fields are displayed first
func displayProperties(w io.Writer, schema *spec.Schema, prefix string) {
	required := schema.Required
	requiredMap := map[string]bool{}
	for _, req := range required {
		requiredMap[req] = true
	}

	reqKeys := []string{}
	for key := range schema.Properties {
		if requiredMap[key] {
			reqKeys = append(reqKeys, key)
		}
	}
	sort.Strings(reqKeys)

	nonReqKeys := []string{}
	for key := range schema.Properties {
		if !requiredMap[key] {
			nonReqKeys = append(nonReqKeys, key)
		}
	}
	sort.Strings(nonReqKeys)
	keys := append(reqKeys, nonReqKeys...)

	for _, key := range keys {
		property := schema.Properties[key]
		requiredInfo := ""
		if requiredMap[key] {
			requiredInfo = "-required-"
		}
		fmt.Fprintf(w, "%s%s (%s)   %s\n", strings.Repeat(" ", 3+2*strings.Count(prefix, ".")), prefix+key, getTypeString(property), requiredInfo)
		nl := false
		if len(property.Title) > 0 {
			fmt.Fprintf(w, "%s\n", indentText(property.Title, 5+2*strings.Count(prefix, ".")))
			nl = true
		}
		if len(property.Description) > 0 {
			fmt.Fprintf(w, "%s\n", indentText(property.Description, 5+2*strings.Count(prefix, ".")))
			nl = true
		}
		if !nl {
			fmt.Fprintln(w)
		}
		if property.Type.Contains("object") {
			displayProperties(w, &property, prefix+key+".")
		} else if property.Type.Contains("array") && property.Items.Schema.Type.Contains("object") {
			displayProperties(w, property.Items.Schema, prefix+key+".*.")
		}
	}
}

func getTypeString(property spec.Schema) string {
	if len(property.Type) != 1 {
		// should not happen
		return strings.Join(property.Type, ", ")
	}
	tpe := property.Type[0]
	if tpe == "array" {
		tpe = "[]" + getTypeString(*property.Items.Schema)
	}
	return tpe
}

func indentText(t string, indent int) string {
	lines := wrapString(t, 80-indent)
	res := ""
	for _, line := range lines {
		res += strings.Repeat(" ", indent) + line + "\n"
	}
	return res
}

// Following code from https://github.com/kubernetes/kubectl/blob/159a770147fb28337c6807abb1b2b9db843d0aff/pkg/explain/formatter.go

type line struct {
	wrap  int
	words []string
}

func (l *line) String() string {
	return strings.Join(l.words, " ")
}

func (l *line) Empty() bool {
	return len(l.words) == 0
}

func (l *line) Len() int {
	return len(l.String())
}

// Add adds the word to the line, returns true if we could, false if we
// didn't have enough room. It's always possible to add to an empty line.
func (l *line) Add(word string) bool {
	newLine := line{
		wrap:  l.wrap,
		words: append(l.words, word),
	}
	if newLine.Len() <= l.wrap || len(l.words) == 0 {
		l.words = newLine.words
		return true
	}
	return false
}

func wrapString(str string, wrap int) []string {
	wrapped := []string{}
	l := line{wrap: wrap}
	// track the last word added to the current line
	lastWord := ""
	flush := func() {
		if !l.Empty() {
			lastWord = ""
			wrapped = append(wrapped, l.String())
			l = line{wrap: wrap}
		}
	}

	// iterate over the lines in the original description
	for _, str := range strings.Split(str, "\n") {
		// preserve code blocks and blockquotes as-is
		if strings.HasPrefix(str, "    ") {
			flush()
			wrapped = append(wrapped, str)
			continue
		}

		// preserve empty lines after the first line, since they can separate logical sections
		if len(wrapped) > 0 && len(strings.TrimSpace(str)) == 0 {
			flush()
			wrapped = append(wrapped, "")
			continue
		}

		// flush if we should start a new line
		if shouldStartNewLine(lastWord, str) {
			flush()
		}
		words := strings.Fields(str)
		for _, word := range words {
			lastWord = word
			if !l.Add(word) {
				flush()
				if !l.Add(word) {
					panic("Couldn't add to empty line.")
				}
			}
		}
	}
	flush()
	return wrapped
}

var bullet = regexp.MustCompile(`^(\d+\.?|-|\*)\s`)

func shouldStartNewLine(lastWord, str string) bool {
	// preserve line breaks ending in :
	if strings.HasSuffix(lastWord, ":") {
		return true
	}

	// preserve code blocks
	if strings.HasPrefix(str, "    ") {
		return true
	}
	str = strings.TrimSpace(str)
	// preserve empty lines
	if len(str) == 0 {
		return true
	}
	// preserve lines that look like they're starting lists
	if bullet.MatchString(str) {
		return true
	}
	// otherwise combine
	return false
}
