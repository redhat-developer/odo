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

func TestSimpleStructType(t *testing.T) {
	tree := getTestTree(basicTypeName, reflect.TypeOf(baseType{}))
	if err := verifyBaseTypeNode("", tree.Root.GetData()); err != nil {
		t.Fatal(err)
	}
}

func TestPtrType(t *testing.T) {
	tree := getTestTree(ptrTypeName, reflect.TypeOf(ptrType{}))
	if err := verifyPtrNode(tree.Root.GetData()); err != nil {
		t.Fatal(err)
	}
}

func TestArrayType(t *testing.T) {
	tree := getTestTree(arrayTypeName, reflect.TypeOf(arrayType{}))
	if err := verifyArrayNode(tree.Root.GetData()); err != nil {
		t.Fatal(err)
	}
}

func TestOtherType(t *testing.T) {
	tree := getTestTree(otherTypeName, reflect.TypeOf(otherType{}))
	if err := verifyOtherTypeNode(tree.Root.GetData()); err != nil {
		t.Fatal(err)
	}
}

func TestCombinedType(t *testing.T) {
	tree := getTestTree(combinedTypeName, reflect.TypeOf(combinedNodeType{}))
	if err := verifyResourceForest(tree.Forest); err != nil {
		t.Fatal(err)
	}
}
