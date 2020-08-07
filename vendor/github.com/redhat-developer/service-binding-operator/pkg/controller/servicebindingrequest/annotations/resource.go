package annotations

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
)

// ResourceHandler handles annotations related to external resources.
type ResourceHandler struct {
	// bindingInfo contains the binding details related to the annotation handler.
	bindingInfo *BindingInfo
	// client is the client used to retrieve a related secret.
	client dynamic.Interface
	// relatedGroupVersionResource is the related resource GVR, used to retrieve the related resource
	// using the client.
	relatedGroupVersionResource schema.GroupVersionResource
	// relatedResourceName is the name of the related resource that is referenced by the handler
	// annotation key/value pair.
	relatedResourceName string
	// resource is the unstructured object to extract data using inputPath.
	resource unstructured.Unstructured
	// stringValue is a function used to decode values from the resource being handled; for example,
	// to decode Base64 keys the decodeBase64String can be used.
	stringValue func(interface{}) (string, error)
	// inputPathRoot indicates the root where input paths will be applied to extract a value from the
	// resource.
	inputPathRoot *string
	// restMapper allows clients to map resources to kind, and map kind and version
	// to interfaces for manipulating those objects.
	restMapper meta.RESTMapper
}

// discoverRelatedResourceName returns the resource name referenced by the handler. Can return an
// error in the case the expected information doesn't exist in the handler's resource object.
func discoverRelatedResourceName(obj map[string]interface{}, bindingInfo *BindingInfo) (string, error) {
	resourceNameValue, ok, err := unstructured.NestedFieldCopy(
		obj,
		strings.Split(bindingInfo.ResourceReferencePath, ".")...,
	)
	if !ok {
		return "", ResourceNameFieldNotFoundErr
	}
	if err != nil {
		return "", err
	}
	name, ok := resourceNameValue.(string)
	if !ok {
		return "", InvalidArgumentErr(bindingInfo.ResourceReferencePath)
	}
	return name, nil
}

// discoverBindingType attempts to extract a binding type from the given annotation value val.
func discoverBindingType(val string) (bindingType, error) {
	re := regexp.MustCompile("^binding:(.*?):.*$")
	parts := re.FindStringSubmatch(val)
	if len(parts) == 0 {
		return "", ErrInvalidBindingValue(val)
	}
	t := bindingType(parts[1])
	_, ok := supportedBindingTypes[t]
	if !ok {
		return "", UnknownBindingTypeErr(t)
	}
	return t, nil
}

// getInputPathFields infers the input path fields based on the given bindingInfo value.
//
// In the case the resource reference path and source path are the same and no input path prefix has
// been given, an empty slice is returned.
//
// In the case inputPathPrefix is present, it is prepended to the resulting slice.
//
// In the case the resource reference and source paths are different, the source path is appended to
// the resulting slice.
func getInputPathFields(bindingInfo *BindingInfo, inputPathPrefix *string) []string {
	inputPathFields := []string{}
	if bindingInfo.ResourceReferencePath != bindingInfo.SourcePath {
		inputPathFields = append(inputPathFields, bindingInfo.SourcePath)
	}
	if inputPathPrefix != nil && len(*inputPathPrefix) > 0 {
		inputPathFields = append([]string{*inputPathPrefix}, inputPathFields...)
	}
	return inputPathFields
}

// Handle returns the value for an external resource strategy.
func (h *ResourceHandler) Handle() (Result, error) {
	ns := h.resource.GetNamespace()
	resource, err := h.
		client.
		Resource(h.relatedGroupVersionResource).
		Namespace(ns).
		Get(h.relatedResourceName, metav1.GetOptions{})
	if err != nil {
		return Result{}, err
	}

	inputPathFields := getInputPathFields(h.bindingInfo, h.inputPathRoot)
	val, ok, err := unstructured.NestedFieldCopy(resource.Object, inputPathFields...)
	if !ok {
		return Result{}, InvalidArgumentErr(strings.Join(inputPathFields, ", "))
	}
	if err != nil {
		return Result{}, err
	}

	if mapVal, ok := val.(map[string]interface{}); ok {
		tmpVal := make(map[string]interface{})
		for k, v := range mapVal {
			decodedVal, err := h.stringValue(v)
			if err != nil {
				return Result{}, err
			}
			tmpVal[k] = decodedVal
		}
		val = tmpVal
	} else {
		val, err = h.stringValue(val)
		if err != nil {
			return Result{}, err
		}
	}

	typ, err := discoverBindingType(h.bindingInfo.Value)
	if err != nil {
		return Result{}, err
	}

	// get resource's kind.
	gvk, err := h.restMapper.KindFor(h.relatedGroupVersionResource)
	if err != nil {
		return Result{}, err
	}

	// prefix the output path with the kind of the resource.
	outputPath := strings.Join([]string{
		strings.ToLower(gvk.Kind),
		h.bindingInfo.SourcePath,
	}, ".")

	return Result{
		Data: nested.ComposeValue(val, nested.NewPath(outputPath)),
		Type: typ,
		Path: outputPath,
	}, nil
}

// stringValue asserts the given value 'v' and returns its string value.
func stringValue(v interface{}) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}
	return s, nil
}

// NewSecretHandler constructs a SecretHandler.
func NewResourceHandler(
	client dynamic.Interface,
	bindingInfo *BindingInfo,
	resource unstructured.Unstructured,
	relatedGroupVersionResource schema.GroupVersionResource,
	inputPathPrefix *string,
	restMapper meta.RESTMapper,
) (*ResourceHandler, error) {
	if client == nil {
		return nil, InvalidArgumentErr("client")
	}

	if bindingInfo == nil {
		return nil, InvalidArgumentErr("bindingInfo")
	}

	if len(bindingInfo.SourcePath) == 0 {
		return nil, InvalidArgumentErr("bindingInfo.Path")
	}

	if len(bindingInfo.ResourceReferencePath) == 0 {
		return nil, InvalidArgumentErr("bindingInfo.ResourceReferencePath")
	}

	relatedResourceName, err := discoverRelatedResourceName(resource.Object, bindingInfo)
	if err != nil {
		return nil, err
	}

	return &ResourceHandler{
		bindingInfo:                 bindingInfo,
		client:                      client,
		inputPathRoot:               inputPathPrefix,
		relatedGroupVersionResource: relatedGroupVersionResource,
		relatedResourceName:         relatedResourceName,
		resource:                    resource,
		stringValue:                 stringValue,
		restMapper:                  restMapper,
	}, nil
}
