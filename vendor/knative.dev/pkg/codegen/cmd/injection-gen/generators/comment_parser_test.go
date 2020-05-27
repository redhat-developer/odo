/*
Copyright 2020 The Knative Authors.

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
package generators

import "testing"

func TestParseComments(t *testing.T) {
	comment := []string{
		"This is an example comment to parse",
		"",
		" +foo",
		"+bar",
		"+with:option",
		"+pair:key=value",
		"+manypairs:key1=value1,key2=value2",
	}

	extracted := ExtractCommentTags("+", comment)

	if val, ok := extracted["foo"]; !ok || val != nil {
		t.Errorf("Failed to extract single key got=%t,%v want=true,nil", ok, val)
	}

	if val, ok := extracted["bar"]; !ok || val != nil {
		t.Errorf("Failed to extract single key got=%t,%v want=true,nil", ok, val)
	}

	if val, ok := extracted["with"]; !ok || val["option"] != "" {
		t.Errorf(`Failed to extract single key got=%t,%v want=true,{"option":""}`, ok, val)
	}

	if val, ok := extracted["pair"]; !ok || val["key"] != "value" {
		t.Errorf(`Failed to extract single key got=%t,%v want=true,{"key":"value"}`, ok, val)
	}

	if val, ok := extracted["manypairs"]; !ok || val["key1"] != "value1" || val["key2"] != "value2" {
		t.Errorf(`Failed to extract single key got=%t,%v want=true,{"key":"value"}`, ok, val)
	}
}

func TestMergeDuplicates(t *testing.T) {
	comment := []string{
		"This is an example comment to parse",
		"",
		"+foo",
		" +foo",
		"+bar:key=value",
		"+bar",
		"+manypairs:key1=value1",
		"+manypairs:key2=value2",
		"+oops:,,,",
	}

	extracted := ExtractCommentTags("+", comment)

	if val, ok := extracted["foo"]; !ok || val != nil {
		t.Errorf("Failed to extract single key got=%t,%v want=true,nil", ok, val)
	}

	if val, ok := extracted["bar"]; !ok || val["key"] != "value" {
		t.Errorf(`Failed to extract single key got=%t,%v want=true,{"key":"value"}`, ok, val)
	}

	if val, ok := extracted["manypairs"]; !ok || val["key1"] != "value1" || val["key2"] != "value2" {
		t.Errorf(`Failed to extract single key got=%t,%v want=true,{"key":"value"}`, ok, val)
	}

	if val, ok := extracted["oops"]; !ok || val != nil {
		t.Errorf(`Failed to extract single key got=%t,%v want=true,{"oops":nil}`, ok, val)
	}
}
