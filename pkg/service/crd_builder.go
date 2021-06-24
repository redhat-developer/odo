package service

import (
	"strings"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
)

// CRDBuilder is responsible for build the full CR including the meta and spec.
type CRDBuilder struct {
	CRDSpecBuilder *CRDSpecBuilder
	crd            *olm.CRDDescription
	cr             map[string]interface{}
}

func NewCRDBuilder(crd *olm.CRDDescription) *CRDBuilder {
	return &CRDBuilder{
		CRDSpecBuilder: NewCRDSpecBuilder(crd.SpecDescriptors),
		crd:            crd,
		cr:             make(map[string]interface{}),
	}
}

func (crb *CRDBuilder) SetAndValidate(param string, value string) error {
	return crb.CRDSpecBuilder.SetAndValidate(param, value)
}

func (crb *CRDBuilder) Map() (map[string]interface{}, error) {
	group, version, _, err := GetGVRFromCR(crb.crd)
	if err != nil {
		return nil, err
	}
	crb.cr["apiVersion"] = group + "/" + version
	crb.cr["kind"] = crb.crd.Kind
	crb.cr["metadata"] = make(map[string]interface{})
	specMap, err := crb.CRDSpecBuilder.Map()
	if err != nil {
		return nil, err
	}
	crb.cr["spec"] = specMap
	return crb.cr, nil
}

// BuildCRDFromParams iterates over the parameter maps provided by the user and builds the CR
func BuildCRDFromParams(cr *olm.CRDDescription, paramMap map[string]string) (map[string]interface{}, error) {

	crBuilder := NewCRDBuilder(cr)
	var errorStrs []string

	for key, value := range paramMap {
		err := crBuilder.SetAndValidate(key, value)
		if err != nil {
			errorStrs = append(errorStrs, err.Error())
		}
	}

	if len(errorStrs) > 0 {
		return nil, errors.New(strings.Join(errorStrs, "\n"))
	}

	builtCRD, err := crBuilder.Map()
	if err != nil {
		return nil, err
	}

	return builtCRD, nil
}
