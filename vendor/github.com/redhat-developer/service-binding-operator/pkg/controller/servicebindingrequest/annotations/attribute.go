package annotations

import (
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const AttributeValue = "binding:env:attribute"

// AttributeHandler handles "binding:env:attribute" annotations.
type AttributeHandler struct {
	// inputPath is the path that should be extracted from the resource. Required.
	inputPath string
	// outputPath is the path the extracted data should be placed under in the
	// resulting unstructured object in Handler. Required.
	outputPath string
	// resource is the unstructured object to extract data using inputPath. Required.
	resource unstructured.Unstructured
}

// Handle returns a unstructured object according to the "binding:env:attribute"
// annotation strategy.
func (h *AttributeHandler) Handle() (Result, error) {
	val, _, err := nested.GetValue(h.resource.Object, h.inputPath, h.outputPath)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Data: val,
	}, nil
}

// IsAttribute returns true if the annotation value should trigger the attribute
// handler.
func IsAttribute(s string) bool {
	return AttributeValue == s
}

// NewAttributeHandler constructs an AttributeHandler.
func NewAttributeHandler(
	bindingInfo *BindingInfo,
	resource unstructured.Unstructured,
) *AttributeHandler {
	outputPath := bindingInfo.SourcePath
	if len(bindingInfo.ResourceReferencePath) > 0 {
		outputPath = bindingInfo.ResourceReferencePath
	}

	// the current implementation removes "status." and "spec." from fields exported through
	// annotations.
	for _, prefix := range []string{"status.", "spec."} {
		if strings.HasPrefix(outputPath, prefix) {
			outputPath = strings.Replace(outputPath, prefix, "", 1)
		}
	}

	return &AttributeHandler{
		inputPath:  bindingInfo.SourcePath,
		outputPath: outputPath,
		resource:   resource,
	}
}
