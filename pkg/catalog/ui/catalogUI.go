package ui

import (
	"encoding/json"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"sort"
)

type serviceInstanceCreateParameterSchema struct {
	Required   []string
	Properties map[string]property
}

type property struct {
	Title       string
	Type        string
	Description string
}

type ServiceClass struct {
	Name            string
	Description     string
	LongDescription string
	Class           scv1beta1.ClusterServiceClass
}

type ServiceClasses []ServiceClass

func (classes ServiceClasses) Len() int {
	return len(classes)
}

func (classes ServiceClasses) Less(i, j int) bool {
	return classes[i].Name < classes[j].Name
}

func (classes ServiceClasses) Swap(i, j int) {
	classes[i], classes[j] = classes[j], classes[i]
}

type Property struct {
	Name        string
	Title       string
	Description string
	Type        string
	Required    bool
}

type Properties []Property

func (props Properties) Len() int {
	return len(props)
}

func (props Properties) Less(i, j int) bool {
	if props[i].Required == props[j].Required {
		return props[i].Name < props[j].Name
	} else {
		return props[i].Required && !props[j].Required
	}
}

func (props Properties) Swap(i, j int) {
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

// Convert the provided ClusterServiceClasses to ServiceClasses
func GetUIServiceClasses(classes []scv1beta1.ClusterServiceClass) (uiClasses ServiceClasses) {
	uiClasses = make(ServiceClasses, 0, len(classes))
	for _, v := range classes {
		uiClasses = append(uiClasses, ConvertToUI(v))
	}

	sort.Sort(uiClasses)
	return uiClasses
}

// Convert the provided ClusterServiceClass to its UI representation
func ConvertToUI(class scv1beta1.ClusterServiceClass) ServiceClass {
	var meta map[string]interface{}
	json.Unmarshal(class.Spec.ExternalMetadata.Raw, &meta)
	longDescription := ""
	if val, ok := meta["longDescription"]; ok {
		longDescription = val.(string)
	}
	return ServiceClass{
		Name:            class.Spec.ExternalName,
		Description:     class.Spec.Description,
		LongDescription: longDescription,
		Class:           class,
	}
}

func GetProperties(plan scv1beta1.ClusterServicePlan) (properties Properties, err error) {
	paramBytes := plan.Spec.CommonServicePlanSpec.ServiceInstanceCreateParameterSchema.Raw
	schema := serviceInstanceCreateParameterSchema{}

	err = json.Unmarshal(paramBytes, &schema)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable unmarshal response: %s", string(paramBytes[:]))
	}

	properties = make([]Property, 0, len(schema.Properties))
	for k, v := range schema.Properties {
		propertyOut := Property{}
		propertyOut.Name = k
		propertyOut.Title = v.Title
		propertyOut.Description = v.Description
		propertyOut.Type = v.Type
		propertyOut.Required = isRequired(schema.Required, k)
		properties = append(properties, propertyOut)
	}

	sort.Sort(properties)
	return properties, err
}

func isRequired(required []string, name string) bool {
	for _, n := range required {
		if n == name {
			return true
		}
	}
	return false
}
