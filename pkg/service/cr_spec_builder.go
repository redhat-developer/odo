package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

// CRSpecBuilder provides all the functionalities to validate and build operands (operators) spec
// based on schema available for them.
type CRSpecBuilder struct {
	descriptors []olm.SpecDescriptor

	builtJsonStr string
	params       map[string]interface{}
}

func NewCRSpecBuilder(descriptors []olm.SpecDescriptor) *CRSpecBuilder {
	return &CRSpecBuilder{
		params:      make(map[string]interface{}),
		descriptors: descriptors,
	}
}

// set sets the param. The param is provided in json path format. e.g. "first.name".
// It is also responsible for parsing the values from string to an appropriate type.
func (pb *CRSpecBuilder) set(param string, value string) error {
	parsedValue := pb.convertType(value)
	pb.params[param] = parsedValue
	tJsonStr, err := sjson.Set(pb.builtJsonStr, param, parsedValue)
	if err != nil {
		return errors.Wrap(err, "error while setting param value for operand")
	}
	pb.builtJsonStr = tJsonStr
	return nil
}

func (pb *CRSpecBuilder) convertType(value string) interface{} {
	intv, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return int64(intv)
	}
	floatv, err := strconv.ParseFloat(value, 64)

	if err == nil {
		return floatv
	}

	boolv, err := strconv.ParseBool(value)
	if err == nil {
		return boolv
	}
	// if there were errors for everything else we return the value
	return value
}

// ValidateAndSet validates if a param is part of the operand schema and then sets it.
func (pb *CRSpecBuilder) ValidateAndSet(param string, value string) error {
	if pb.hasParam(param) {
		return pb.set(param, value)
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

// JSON returns the final json string
func (pb *CRSpecBuilder) JSON() string {
	return pb.builtJsonStr
}
