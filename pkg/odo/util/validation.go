package util

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ValidateName will do validation of application & component names
// Criteria for valid name in kubernetes: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
func ValidateName(name string) error {

	errorList := validation.IsDNS1123Label(name)

	if len(errorList) != 0 {
		return errors.New(fmt.Sprintf("%s is not a valid name:  %s", name, strings.Join(errorList, " ")))
	}

	return nil

}
