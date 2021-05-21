package service

import (
	"encoding/json"
	"fmt"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

// CRSpecBuilder provides all the functionalities to validate and build operands (operators) spec
// based on schema available for them.
type CRSpecBuilder struct {
	descriptors []olm.SpecDescriptor

	builtJsonStr string
	params       map[string]string
}

func NewCRSpecBuilder(descriptors []olm.SpecDescriptor) *CRSpecBuilder {
	return &CRSpecBuilder{
		params:      make(map[string]string),
		descriptors: descriptors,
	}
}

// set sets the param. The param is provided in json path format. e.g. "first.name"
func (pb *CRSpecBuilder) set(param string, value string) error {
	pb.params[param] = value
	tJsonStr, err := sjson.Set(pb.builtJsonStr, param, value)
	if err != nil {
		return errors.Wrap(err, "error while setting param value for operand")
	}
	pb.builtJsonStr = tJsonStr
	return nil
}

// SetAndValidate validates if a param is part of the operand schema and then sets it.
func (pb *CRSpecBuilder) SetAndValidate(param string, value string) error {
	if pb.hasParam(param) {
		pb.set(param, value)
		return nil
	}
	return fmt.Errorf("the parameter %s is not present in the Operand Schema", param)
}

func (pb *CRSpecBuilder) hasParam(param string) bool {
	for _, desc := range pb.descriptors {
		if desc.Path == param {
			return true
		}
	}
	return false
}

// Map returns the final map
func (pb *CRSpecBuilder) Map() (map[string]interface{}, error) {
	var out map[string]interface{}

	err := json.Unmarshal([]byte(pb.builtJsonStr), &out)
	return out, err
}
