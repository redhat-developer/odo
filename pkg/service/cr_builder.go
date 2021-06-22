package service

import (
	"strings"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
)

type CRBuilder struct {
	CRSpecBuilder *CRSpecBuilder
	crd           *olm.CRDDescription
	cr            map[string]interface{}
}

func NewCRBuilder(crd *olm.CRDDescription) *CRBuilder {
	return &CRBuilder{
		CRSpecBuilder: NewCRSpecBuilder(crd.SpecDescriptors),
		crd:           crd,
		cr:            make(map[string]interface{}),
	}
}

func (crb *CRBuilder) SetAndValidate(param string, value string) error {
	return crb.CRSpecBuilder.SetAndValidate(param, value)
}

func (crb *CRBuilder) Map() (map[string]interface{}, error) {
	group, version, _, err := GetGVRFromCR(crb.crd)
	if err != nil {
		return nil, err
	}
	crb.cr["apiVersion"] = group + "/" + version
	crb.cr["kind"] = crb.crd.Kind
	crb.cr["metadata"] = make(map[string]interface{})
	specMap, err := crb.CRSpecBuilder.Map()
	if err != nil {
		return nil, err
	}
	crb.cr["spec"] = specMap
	return crb.cr, nil
}

// BuildCRDFromParams iterates over the parameter maps provided by the user and builds the CR
func BuildCRDFromParams(cr *olm.CRDDescription, paramMap map[string]string) (map[string]interface{}, error) {

	crBuilder := NewCRBuilder(cr)
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
