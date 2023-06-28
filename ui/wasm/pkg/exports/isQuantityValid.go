package exports

import (
	"syscall/js"

	"k8s.io/apimachinery/pkg/api/resource"
)

func IsQuantityValidWrapper(this js.Value, args []js.Value) interface{} {
	return isQuantityValid(args[0].String())
}

func isQuantityValid(quantity string) bool {
	_, err := resource.ParseQuantity(quantity)
	return err == nil
}
