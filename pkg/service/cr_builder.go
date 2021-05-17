package service

import (
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

type CRBuilder struct {
	*CRSpecBuilder
	crd *olm.CRDDescription
	cr  map[string]interface{}
}

func NewCRBuilder(crd *olm.CRDDescription) *CRBuilder {
	return &CRBuilder{
		CRSpecBuilder: NewCRSpecBuilder(crd.SpecDescriptors),
		crd:           crd,
		cr:            make(map[string]interface{}),
	}
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
