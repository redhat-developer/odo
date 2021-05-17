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
	crb.cr["apiVersion"] = crb.crd.Version
	crb.cr["kind"] = crb.crd.Kind
	specMap, err := crb.CRSpecBuilder.Map()
	if err != nil {
		return nil, err
	}
	crb.cr["spec"] = specMap
	return crb.cr, nil
}
