package collect

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/binding"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var DataNotMap = errors.New("Returned data are not a map, skip collecting")
var ErrorValueNotFound = errors.New("Value not found in map")

const (
	ErrorReadingServicesReason   = "ErrorReadingServices"
	ErrorReadingCRD              = "ErrorReadingCRD"
	ErrorReadingDescriptorReason = "ErrorReadingDescriptor"
	ErrorReadingBindingReason    = "ErrorReadingBinding"
	ErrorReadingSecret           = "ErrorReadingSecret"

	ValueNotFound     = "ValueNotFound"
	InvalidAnnotation = "InvalidAnnotation"
)

func PreFlight(ctx pipeline.Context) {
	_, err := ctx.Services()
	if err != nil {
		requestRetry(ctx, ErrorReadingServicesReason, err)
		return
	}
}

func BindingDefinitions(ctx pipeline.Context) {
	services, _ := ctx.Services()

	for _, service := range services {
		anns := make(map[string]string)
		crd, err := service.CustomResourceDefinition()
		if err != nil {
			requestRetry(ctx, ErrorReadingCRD, err)
			return
		}
		if crd != nil {
			descr, err := crd.Descriptor()
			if err != nil {
				requestRetry(ctx, ErrorReadingDescriptorReason, err)
				return
			}
			if descr != nil {
				util.MergeMaps(anns, descr.BindingAnnotations())
			}
			util.MergeMaps(anns, crd.Resource().GetAnnotations())
		}

		util.MergeMaps(anns, service.Resource().GetAnnotations())

		for k, v := range anns {
			definition, err := makeBindingDefinition(k, v, ctx)
			if err != nil {
				condition := notCollectionReadyCond(InvalidAnnotation, fmt.Errorf("Failed to create binding definition from \"%v: %v\": %v", k, v, err))
				ctx.SetCondition(condition)
				ctx.Error(err)
				ctx.StopProcessing()
			}
			if definition != nil {
				service.AddBindingDef(definition)
			}
		}
	}
}

func BindingItems(ctx pipeline.Context) {
	services, _ := ctx.Services()

	for _, service := range services {
		serviceResource := service.Resource()
		for _, bd := range service.BindingDefs() {
			bindingValue, err := bd.Apply(serviceResource)
			if err != nil {
				requestRetry(ctx, ErrorReadingBindingReason, err)
				return
			}
			val := bindingValue.Get()
			v := reflect.ValueOf(val)
			if v.Kind() != reflect.Map {
				requestRetry(ctx, "DataNotMap", DataNotMap)
				return
			}
			for _, n := range v.MapKeys() {
				collectItems("", ctx, service, n, v.MapIndex(n).Interface())
			}
		}
	}
}

func ProvisionedService(ctx pipeline.Context) {
	services, _ := ctx.Services()

	for _, service := range services {
		res := service.Resource()
		secretName, found, err := unstructured.NestedString(res.Object, "status", "binding", "name")
		if err != nil {
			requestRetry(ctx, ErrorReadingBindingReason, err)
			return
		}
		if found {
			if secretName != "" {
				secret, err := ctx.ReadSecret(res.GetNamespace(), secretName)
				if err != nil {
					requestRetry(ctx, ErrorReadingSecret, err)
					return
				}
				ctx.AddBindings(&pipeline.SecretBackedBindings{Service: service, Secret: secret})
			}
		} else {
			crd, err := service.CustomResourceDefinition()
			if err != nil {
				requestRetry(ctx, ErrorReadingCRD, err)
				return
			}
			if crd == nil {
				continue
			}
			v, ok := crd.Resource().GetAnnotations()[binding.ProvisionedServiceAnnotationKey]
			if ok && v == "true" {
				requestRetry(ctx, ErrorReadingBindingReason, fmt.Errorf("CRD of service %v/%v indicates provisioned service, but no secret name provided under .status.binding.name", res.GetNamespace(), res.GetName()))
				return
			}
		}
	}
}

func DirectSecretReference(ctx pipeline.Context) {
	// Error is ignored as this check is there in the PreFlight stage.
	// That stage was created to perform common checks for all followup stages.
	services, _ := ctx.Services()

	for _, service := range services {
		res := service.Resource()
		if res.GetKind() == "Secret" && res.GetAPIVersion() == "v1" && res.GroupVersionKind().Group == "" {
			annotations := res.GetAnnotations()
			for k := range annotations {
				if strings.HasPrefix(k, binding.AnnotationPrefix) {
					return
				}
			}
			name := res.GetName()
			secret, err := ctx.ReadSecret(res.GetNamespace(), name)
			if err != nil {
				requestRetry(ctx, ErrorReadingSecret, err)
				return
			}
			ctx.AddBindings(&pipeline.SecretBackedBindings{Service: service, Secret: secret})
		}
	}
}

type pathMapping struct {
	input     string
	transform func(interface{}) (interface{}, error)
	output    string
}

var bindableResources = map[schema.GroupVersionKind]pathMapping{
	schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}: {
		input:  "data",
		output: "",
	},
	schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}: {
		input: "data",
		transform: func(i interface{}) (interface{}, error) {
			v := reflect.ValueOf(i)
			if v.Kind() != reflect.Map {
				return nil, errors.New("data is not map")
			}
			result := map[string]string{}

			for _, n := range v.MapKeys() {
				b, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", v.MapIndex(n).Interface()))
				if err != nil {
					return nil, err
				}
				key := fmt.Sprintf("%v", n.Interface())
				result[key] = string(b)
			}
			return result, nil
		},
		output: "",
	},
	schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}: {
		input:  "spec.clusterIP",
		output: "clusterIP",
	},
	schema.GroupVersionKind{
		Group:   "route.openshift.io",
		Version: "v1",
		Kind:    "Route",
	}: {
		input:  "spec.host",
		output: "host",
	},
}

func OwnedResources(ctx pipeline.Context) {
	services, err := ctx.Services()
	if err != nil {
		requestRetry(ctx, ErrorReadingServicesReason, err)
		return
	}
	for _, service := range services {
		ownedResources, err := service.OwnedResources()
		if err != nil {
			requestRetry(ctx, ErrorReadingServicesReason, err)
			return
		}
		for _, res := range ownedResources {
			pathMapping, ok := bindableResources[res.GroupVersionKind()]
			if !ok {
				continue
			}
			val, found, err := unstructured.NestedFieldNoCopy(res.Object, strings.Split(pathMapping.input, ".")...)
			if !found {
				err = errors.New("Not found")
			}
			if err != nil {
				requestRetry(ctx, ErrorReadingServicesReason, err)
				return
			}
			if pathMapping.transform != nil {
				val, err = pathMapping.transform(val)
				if err != nil {
					requestRetry(ctx, ErrorReadingServicesReason, err)
					return
				}
			}
			collectItems("", ctx, service, reflect.ValueOf(pathMapping.output), val)
		}
	}
}

func collectItems(prefix string, ctx pipeline.Context, service pipeline.Service, k reflect.Value, val interface{}) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Map:
		p := prefix + k.String() + "_"
		if p == "_" {
			p = ""
		}
		for _, n := range v.MapKeys() {
			if mapVal := v.MapIndex(n).Interface(); mapVal != nil {
				collectItems(p, ctx, service, n, mapVal)
			} else {
				condition := notCollectionReadyCond(ValueNotFound, fmt.Errorf("Value for key %v_%v not found", prefix+k.String(), n.String()))
				ctx.SetCondition(condition)
				ctx.Error(ErrorValueNotFound)
				ctx.StopProcessing()
				return
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			ctx.AddBindingItem(&pipeline.BindingItem{Name: fmt.Sprintf("%v_%v", prefix+k.String(), i), Value: v.Index(i).Interface(), Source: service})
		}
	default:
		ctx.AddBindingItem(&pipeline.BindingItem{Name: prefix + k.String(), Value: v.Interface(), Source: service})
	}
}

func requestRetry(ctx pipeline.Context, reason string, err error) {
	ctx.RetryProcessing(err)
	ctx.SetCondition(notCollectionReadyCond(reason, err))
}

func notCollectionReadyCond(reason string, err error) *metav1.Condition {
	return apis.Conditions().NotCollectionReady().Reason(reason).Msg(err.Error()).Build()
}

func makeBindingDefinition(key string, value string, ctx pipeline.Context) (binding.Definition, error) {
	return binding.NewDefinitionBuilder(key,
		value,
		func(namespace string, name string) (*unstructured.Unstructured, error) {
			return ctx.ReadConfigMap(namespace, name)
		},
		func(namespace string, name string) (*unstructured.Unstructured, error) {
			return ctx.ReadSecret(namespace, name)
		}).Build()
}
