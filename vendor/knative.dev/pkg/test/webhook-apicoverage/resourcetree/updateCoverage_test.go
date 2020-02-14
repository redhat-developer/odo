/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resourcetree

import (
	"reflect"
	"testing"
)

func TestSimpleStructValue(t *testing.T) {
	tree := getTestTree(basicTypeName, reflect.TypeOf(baseType{}))
	tree.UpdateCoverage(reflect.ValueOf(getBaseTypeValue()))
	if err := verifyBaseTypeValue("", tree.Root); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCoverage(t *testing.T) {
	datas := []struct {
		TestName string
		name     string
		typeI    interface{}
		value    interface{}
		f        func(NodeInterface) error
	}{{
		"TestPtrValueAllCovered", ptrTypeName, ptrType{}, getPtrTypeValueAllCovered(), verifyPtrValueAllCovered,
	}, {
		"TestPtrValueSomeCovered", ptrTypeName, ptrType{}, getPtrTypeValueSomeCovered(), verifyPtrValueSomeCovered,
	}, {
		"TestArrValueAllCovered", arrayTypeName, arrayType{}, getArrValueAllCovered(), verifyArryValueAllCovered,
	}, {
		"TestArrValueSomeCovered", arrayTypeName, arrayType{}, getArrValueSomeCovered(), verifyArrValueSomeCovered,
	}, {
		"TestOtherValue", otherTypeName, otherType{}, getOtherTypeValue(), verifyOtherTypeValue,
	}}

	for _, data := range datas {
		t.Run(data.TestName, func(t *testing.T) {
			tree := getTestTree(data.name, reflect.TypeOf(data.typeI))
			tree.UpdateCoverage(reflect.ValueOf(data.value))
			if err := data.f(tree.Root); err != nil {
				t.Fatal(err)
			}
		})

	}
}
