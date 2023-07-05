package devstate

import "k8s.io/apimachinery/pkg/api/resource"

func IsQuantityValid(quantity string) bool {
	_, err := resource.ParseQuantity(quantity)
	return err == nil
}
