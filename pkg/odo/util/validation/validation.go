package validation

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// ValidateName will do validation of application & component names according to DNS (RFC 1123) rules
// Criteria for valid name in kubernetes: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/identifiers.md
func ValidateName(name string) error {

	errorList := validation.IsDNS1123Label(name)

	if len(errorList) != 0 {
		return fmt.Errorf("%s is not a valid name:  %s", name, strings.Join(errorList, " "))
	}

	return nil

}
