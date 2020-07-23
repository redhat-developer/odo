package servicebindingrequest

import "k8s.io/apimachinery/pkg/types"

// IsNamespacedNameEmpty returns true if any of the fields from the given namespacedName is empty.
func IsNamespacedNameEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}
