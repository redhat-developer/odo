//
//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha2

import (
	"errors"
	"reflect"
)

func visitUnion(union interface{}, visitor interface{}) (err error) {
	visitorValue := reflect.ValueOf(visitor)
	unionValue := reflect.ValueOf(union)
	oneMemberPresent := false
	typeOfVisitor := visitorValue.Type()
	for i := 0; i < visitorValue.NumField(); i++ {
		unionMemberToRead := typeOfVisitor.Field(i).Name
		unionMember := unionValue.FieldByName(unionMemberToRead)
		if !unionMember.IsZero() {
			if oneMemberPresent {
				err = errors.New("Only one element should be set in union: " + unionValue.Type().Name())
				return
			}
			oneMemberPresent = true
			visitorFunction := visitorValue.Field(i)
			if visitorFunction.IsNil() {
				return
			}
			results := visitorFunction.Call([]reflect.Value{unionMember})
			if !results[0].IsNil() {
				err = results[0].Interface().(error)
			}
			return
		}
	}
	return
}

func simplifyUnion(union Union, visitorType reflect.Type) {
	_ = normalizeUnion(union, visitorType)
	*union.discriminator() = ""
}

func normalizeUnion(union Union, visitorType reflect.Type) error {
	err := updateDiscriminator(union, visitorType)
	if err != nil {
		return err
	}

	err = cleanupValues(union, visitorType)
	if err != nil {
		return err
	}
	return nil
}

func updateDiscriminator(union Union, visitorType reflect.Type) error {
	unionValue := reflect.ValueOf(union)

	if union.discriminator() == nil {
		return errors.New("Discriminator should not be 'nil' in union: " + unionValue.Type().Name())
	}

	if *union.discriminator() != "" {
		// Nothing to do
		return nil
	}

	oneMemberPresent := false
	for i := 0; i < visitorType.NumField(); i++ {
		unionMemberToRead := visitorType.Field(i).Name
		unionMember := unionValue.Elem().FieldByName(unionMemberToRead)
		if !unionMember.IsZero() {
			if oneMemberPresent {
				return errors.New("Discriminator cannot be deduced from 2 values in union: " + unionValue.Type().Name())
			}
			oneMemberPresent = true
			*(union.discriminator()) = unionMemberToRead
		}
	}
	return nil
}

func cleanupValues(union Union, visitorType reflect.Type) error {
	unionValue := reflect.ValueOf(union)

	if union.discriminator() == nil {
		return errors.New("Discriminator should not be 'nil' in union: " + unionValue.Type().Name())
	}

	if *union.discriminator() == "" {
		// Nothing to do
		return errors.New("Values cannot be cleaned up without a discriminator in union: " + unionValue.Type().Name())
	}

	for i := 0; i < visitorType.NumField(); i++ {
		unionMemberToRead := visitorType.Field(i).Name
		unionMember := unionValue.Elem().FieldByName(unionMemberToRead)
		if !unionMember.IsZero() {
			if unionMemberToRead != *union.discriminator() {
				unionMember.Set(reflect.Zero(unionMember.Type()))
			}
		}
	}
	return nil
}
