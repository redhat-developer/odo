package v1alpha2

// +kubebuilder:validation:Enum=replace;delete
type OverridingPatchDirective string

const (
	ReplaceOverridingDirective OverridingPatchDirective = "replace"
	DeleteOverridingDirective  OverridingPatchDirective = "delete"
)

const (
	DeleteFromPrimitiveListOverridingPatchDirective OverridingPatchDirective = "replace"
)

type OverrideDirective struct {
	// Path of the element the directive should be applied on
	//
	// For the following path tree:
	//
	// 	```json
	// 	commands:
	// 	  - exec
	// 	      id: commandId
	// 	```
	//
	// the path would be: `commands["commandId"]`.
	Path string `json:"path"`

	// `$Patch` directlive as defined in
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md#basic-patch-format
	//
	// This is an enumeration that allows the following values:
	//
	// - *replace*: indicates that the element matched by the `jsonPath` field should be replaced instead of being merged.
	//
	// - *delete*: indicates that the element matched by the `jsonPath` field should be deleted.
	//
	// +optional
	Patch OverridingPatchDirective `json:"patch,omitempty"`

	// `DeleteFromPrimitiveList` directive as defined in
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md#deletefromprimitivelist-directive
	//
	// This indicates that the elements in this list should be deleted from the original primitive list.
	// The original primitive list is the element matched by the `jsonPath` field.
	// +optional
	DeleteFromPrimitiveList []string `json:"deleteFromPrimitiveList,omitempty"`

	// `SetElementOrder` directive as defined in
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md#deletefromprimitivelist-directive
	//
	// This provides a way to specify the order of a list. The relative order specified in this directive will be retained.
	// The list whose order is controller is the element matched by the `jsonPath` field.
	// If the controller list is a list of objects, then the values in this list should be
	// the merge keys of the objects to order.
	// +optional
	SetElementOrder []string `json:"setElementOrder,omitempty"`
}
