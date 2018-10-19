package ui

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/occlient"
	"sort"
)

type serviceInstanceCreateParameterSchema struct {
	Required   []string
	Properties map[string]property
}

type property struct {
	name        string
	Title       string
	Type        string
	Description string
	required    bool
}

type serviceClass struct {
	Name            string
	Description     string
	LongDescription string
	Class           scv1beta1.ClusterServiceClass
}

type serviceClasses []serviceClass

func (classes serviceClasses) Len() int {
	return len(classes)
}

func (classes serviceClasses) Less(i, j int) bool {
	return classes[i].Name < classes[j].Name
}

func (classes serviceClasses) Swap(i, j int) {
	classes[i], classes[j] = classes[j], classes[i]
}

type properties []property

func (props properties) Len() int {
	return len(props)
}

func (props properties) Less(i, j int) bool {
	if props[i].required == props[j].required {
		return props[i].name < props[j].name
	} else {
		return props[i].required && !props[j].required
	}
}

func (props properties) Swap(i, j int) {
	props[i], props[j] = props[j], props[i]
}

// Retrieve the list of existing service class categories
// TODO: should match what the okd web console is doing
func GetServiceClassesCategories(categories map[string][]scv1beta1.ClusterServiceClass) (keys []string) {
	keys = make([]string, len(categories))

	i := 0
	for k := range categories {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

func GetServicePlanNames(stringMap map[string]scv1beta1.ClusterServicePlan) (keys []string) {
	keys = make([]string, len(stringMap))

	i := 0
	for k := range stringMap {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

// Convert the provided ClusterServiceClasses to serviceClasses
func getUIServiceClasses(classes []scv1beta1.ClusterServiceClass) (uiClasses serviceClasses) {
	uiClasses = make(serviceClasses, 0, len(classes))
	for _, v := range classes {
		uiClasses = append(uiClasses, ConvertToUI(v))
	}

	sort.Sort(uiClasses)
	return uiClasses
}

// Convert the provided ClusterServiceClass to its UI representation
func ConvertToUI(class scv1beta1.ClusterServiceClass) serviceClass {
	var meta map[string]interface{}
	err := json.Unmarshal(class.Spec.ExternalMetadata.Raw, &meta)
	if err != nil {
		glog.V(4).Infof("Unable unmarshal Extension metadata for ClusterServiceClass '%v'", class.Spec.ExternalName)
	}
	longDescription := ""
	if val, ok := meta["longDescription"]; ok {
		longDescription = val.(string)
	}
	return serviceClass{
		Name:            class.Spec.ExternalName,
		Description:     class.Spec.Description,
		LongDescription: longDescription,
		Class:           class,
	}
}

func getProperties(plan scv1beta1.ClusterServicePlan) (props properties, err error) {
	paramBytes := plan.Spec.CommonServicePlanSpec.ServiceInstanceCreateParameterSchema.Raw
	schema := serviceInstanceCreateParameterSchema{}

	err = json.Unmarshal(paramBytes, &schema)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable unmarshal response: %s", string(paramBytes[:]))
	}

	props = make(properties, 0, len(schema.Properties))
	for k, v := range schema.Properties {
		v.name = k
		v.required = isRequired(schema.Required, k)
		props = append(props, v)
	}

	sort.Sort(props)
	return props, err
}

func isRequired(required []string, name string) bool {
	for _, n := range required {
		if n == name {
			return true
		}
	}
	return false
}

func SelectPlanNameInteractively(plans map[string]scv1beta1.ClusterServicePlan, promptLabel string) string {
	prompt := promptui.Select{
		Label: promptLabel,
		Items: GetServicePlanNames(plans),
	}
	_, plan, _ := prompt.Run()
	return plan
}

func SelectServiceNameInteractively(defaultValue, promptLabel string, validateName func(string) error) string {
	// if only one arg is given, ask to Name the service providing the class Name as default
	instancePrompt := promptui.Prompt{
		Label:     promptLabel,
		Default:   defaultValue,
		AllowEdit: true,
		Validate:  validateName,
	}
	serviceName, _ := instancePrompt.Run()
	return serviceName
}

func SelectClassInteractively(client *occlient.Client) (class scv1beta1.ClusterServiceClass, serviceType string) {
	classesByCategory, _ := client.GetServiceClassesByCategory()
	prompt := promptui.Select{
		Label: "Which kind of service do you wish to create?",
		Items: GetServiceClassesCategories(classesByCategory),
	}
	_, category, _ := prompt.Run()
	templates := &promptui.SelectTemplates{
		Active:   "\U00002620 {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\U00002620 {{ .Name | red | cyan }}",
		Details: `
--------- Service Class ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ "Long:" | faint }}	{{ .LongDescription }}`,
	}
	uiClasses := getUIServiceClasses(classesByCategory[category])
	prompt = promptui.Select{
		Label:     "Which " + category + " service class should we use?",
		Items:     uiClasses,
		Templates: templates,
	}
	i, _, _ := prompt.Run()
	uiClass := uiClasses[i]

	return uiClass.Class, uiClass.Name
}

func EnterServicePropertiesInteractively(svcPlan scv1beta1.ClusterServicePlan, passedValues map[string]string) (values map[string]string) {
	properties, _ := getProperties(svcPlan)
	propsNb := len(properties)
	values = make(map[string]string, propsNb)

	var i = 0
	for i < propsNb && properties[i].required {
		prop := properties[i]
		if _, ok := passedValues[prop.name]; !ok {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Enter a value for %s property %s ", prop.Type, prop.Title),
				AllowEdit: true,
			}

			result, _ := prompt.Run()
			values[prop.name] = result
		}

		i++
	}
	// if we have non-required properties, ask if user wants to provide values
	if i < propsNb-1 {
		// todo
	}

	return values
}
