/*
Copyright 2019 The Knative Authors

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

package apis_test

import (
	"context"
	"strings"
	"testing"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
	. "knative.dev/pkg/testing"
)

func TestCheckDeprecated(t *testing.T) {

	testCases := map[string]struct {
		strict   bool
		obj      interface{}
		wantErrs []string
	}{
		"create strict, string": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedString: "an error",
			},
			wantErrs: []string{
				"must not set",
				"string",
			},
		},
		"create strict, stringptr": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedStringPtr: ptr.String("test string"),
			},
			wantErrs: []string{
				"must not set",
				"stringPtr",
			},
		},
		"create strict, int": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedInt: 42,
			},
			wantErrs: []string{
				"must not set",
				"int",
			},
		},
		"create strict, intptr": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedIntPtr: ptr.Int64(42),
			},
			wantErrs: []string{
				"must not set",
				"intPtr",
			},
		},
		"create strict, map": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedMap: map[string]string{"hello": "failure"},
			},
			wantErrs: []string{
				"must not set",
				"map",
			},
		},
		"create strict, slice": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedSlice: []string{"hello", "failure"},
			},
			wantErrs: []string{
				"must not set",
				"slice",
			},
		},
		"create strict, struct": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedStruct: InnerDefaultStruct{FieldAsString: "not ok"},
			},
			wantErrs: []string{
				"must not set",
				"struct",
			},
		},
		"create strict, structptr": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedStructPtr: &InnerDefaultStruct{
					FieldAsString: "fail",
				},
			},
			wantErrs: []string{
				"must not set",
				"structPtr",
			},
		},
		"create strict, not json": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedNotJson: "fail",
			},
			wantErrs: []string{
				"must not set",
				"DeprecatedNotJson",
			},
		},
		"create strict, inlined": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				InlinedStruct: InlinedStruct{
					DeprecatedField: "fail",
				},
			},
			wantErrs: []string{
				"must not set",
				"fieldA",
			},
		},
		"create strict, inlined ptr": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				InlinedPtrStruct: &InlinedPtrStruct{
					DeprecatedField: "fail",
				},
			},
			wantErrs: []string{
				"must not set",
				"fieldB",
			},
		},
		"create strict, inlined nested": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				InlinedStruct: InlinedStruct{
					InlinedPtrStruct: &InlinedPtrStruct{
						DeprecatedField: "fail",
					},
				},
			},
			wantErrs: []string{
				"must not set",
				"fieldB",
			},
		},
		"create strict, all errors": {
			strict: true,
			obj: &InnerDefaultSubSpec{
				DeprecatedString:    "an error",
				DeprecatedStringPtr: ptr.String("test string"),
				DeprecatedInt:       42,
				DeprecatedIntPtr:    ptr.Int64(42),
				DeprecatedMap:       map[string]string{"hello": "failure"},
				DeprecatedSlice:     []string{"hello", "failure"},
				DeprecatedStruct:    InnerDefaultStruct{FieldAsString: "not ok"},
				DeprecatedStructPtr: &InnerDefaultStruct{
					FieldAsString: "fail",
				},
				InlinedStruct: InlinedStruct{
					DeprecatedField: "fail",
					InlinedPtrStruct: &InlinedPtrStruct{
						DeprecatedField: "fail",
					},
				},
			},
			wantErrs: []string{
				"string",
				"stringPtr",
				"int",
				"intPtr",
				"map",
				"slice",
				"struct",
				"structPtr",
				"fieldA",
				"fieldB",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ctx := context.Background()
			if tc.strict {
				ctx = apis.DisallowDeprecated(ctx)
			}
			resp := apis.CheckDeprecated(ctx, tc.obj)

			if len(tc.wantErrs) > 0 {
				for _, err := range tc.wantErrs {
					var gotErr string
					if resp != nil {
						gotErr = resp.Error()
					}
					if !strings.Contains(gotErr, err) {
						t.Errorf("Expected failure containing %q got %q", err, gotErr)
					}
				}
			} else if resp != nil {
				t.Errorf("Expected no error, got %q", resp.Error())
			}
		})
	}
}

// This test makes sure that errors will flatten the duped error for fieldB.
// It comes in on obj.InlinedStruct.InlinedPtrStruct.DeprecatedField and
// obj.InlinedPtrStruct.DeprecatedField.
func TestCheckDeprecated_Dedupe(t *testing.T) {

	obj := &InnerDefaultSubSpec{
		InlinedStruct: InlinedStruct{
			DeprecatedField: "fail",
			InlinedPtrStruct: &InlinedPtrStruct{
				DeprecatedField: "fail",
			},
		},
		InlinedPtrStruct: &InlinedPtrStruct{
			DeprecatedField: "fail",
		},
	}
	wantErr := "must not set the field(s): fieldA, fieldB"

	ctx := apis.DisallowDeprecated(context.Background())
	resp := apis.CheckDeprecated(ctx, obj)

	gotErr := resp.Error()
	if gotErr != wantErr {
		t.Errorf("Expected failure %q got %q", wantErr, gotErr)
	}
}

func TestCheckDeprecatedUpdate(t *testing.T) {

	testCases := map[string]struct {
		strict   bool
		obj      interface{}
		org      interface{}
		wantErrs []string
	}{
		"update strict, intptr": {
			strict: true,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedIntPtr: ptr.Int64(42),
			},
			wantErrs: []string{
				"must not set",
				"intPtr",
			},
		},
		"update strict, map": {
			strict: true,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedMap: map[string]string{"hello": "failure"},
			},

			wantErrs: []string{
				"must not set",
				"map",
			},
		},
		"update strict, slice": {
			strict: true,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedSlice: []string{"hello", "failure"},
			},
			wantErrs: []string{
				"must not set",
				"slice",
			},
		},
		"update strict, struct": {
			strict: true,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedStruct: InnerDefaultStruct{
					FieldAsString: "fail",
				},
			},
			wantErrs: []string{
				"must not set",
				"struct",
			},
		},
		"update strict, structptr": {
			strict: true,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedStructPtr: &InnerDefaultStruct{
					FieldAsString: "fail",
				},
			},
			wantErrs: []string{
				"must not set",
				"structPtr",
			},
		},
		"update strict, all errors": {
			strict: true,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedString:    "an error",
				DeprecatedStringPtr: ptr.String("test string"),
				DeprecatedInt:       42,
				DeprecatedIntPtr:    ptr.Int64(42),
				DeprecatedMap:       map[string]string{"hello": "failure"},
				DeprecatedSlice:     []string{"hello", "failure"},
				DeprecatedStruct:    InnerDefaultStruct{FieldAsString: "not ok"},
				DeprecatedStructPtr: &InnerDefaultStruct{
					FieldAsString: "fail",
				},
				InlinedStruct: InlinedStruct{
					DeprecatedField: "fail",
					InlinedPtrStruct: &InlinedPtrStruct{
						DeprecatedField: "fail",
					},
				},
			},
			wantErrs: []string{
				"string",
				"stringPtr",
				"int",
				"intPtr",
				"map",
				"slice",
				"struct",
				"structPtr",
				"fieldA",
				"fieldB",
			},
		},

		"overwrite strict, string": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedString: "original setting.",
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedString: "fail setting.",
			},
			wantErrs: []string{
				"must not update",
				"string",
			},
		},
		"overwrite strict, stringptr": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedStringPtr: ptr.String("original string"),
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedStringPtr: ptr.String("fail string"),
			},
			wantErrs: []string{
				"must not update",
				"stringPtr",
			},
		},
		"overwrite strict, int": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedInt: 10,
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedInt: 42,
			},
			wantErrs: []string{
				"must not update",
				"int",
			},
		},
		"overwrite strict, intptr": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedIntPtr: ptr.Int64(10),
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedIntPtr: ptr.Int64(42),
			},
			wantErrs: []string{
				"must not update",
				"intPtr",
			},
		},
		"overwrite strict, map": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedMap: map[string]string{"goodbye": "existing"},
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedMap: map[string]string{"hello": "failure"},
			},
			wantErrs: []string{
				"must not update",
				"map",
			},
		},
		"overwrite strict, slice": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedSlice: []string{"hello", "existing"},
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedSlice: []string{"hello", "failure"},
			},
			wantErrs: []string{
				"must not update",
				"slice",
			},
		},
		"overwrite strict, struct": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedStruct: InnerDefaultStruct{
					FieldAsString: "original",
				},
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedStruct: InnerDefaultStruct{
					FieldAsString: "fail",
				},
			},
			wantErrs: []string{
				"must not update",
				"struct",
			},
		},
		"overwrite strict, structptr": {
			strict: true,
			org: &InnerDefaultSubSpec{
				DeprecatedStructPtr: &InnerDefaultStruct{
					FieldAsString: "original",
				},
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedStructPtr: &InnerDefaultStruct{
					FieldAsString: "fail",
				},
			},
			wantErrs: []string{
				"must not update",
				"structPtr",
			},
		},

		"create, not strict": {
			strict: false,
			obj: &InnerDefaultSubSpec{
				DeprecatedString: "fail setting.",
			},
		},
		"update, not strict": {
			strict: false,
			org:    &InnerDefaultSubSpec{},
			obj: &InnerDefaultSubSpec{
				DeprecatedString: "it's k",
			},
		},
		"overwrite, not strict": {
			strict: false,
			org: &InnerDefaultSubSpec{
				DeprecatedString: "org",
			},
			obj: &InnerDefaultSubSpec{
				DeprecatedString: "it's k",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ctx := context.Background()
			if tc.strict {
				ctx = apis.DisallowDeprecated(ctx)
			}
			resp := apis.CheckDeprecatedUpdate(ctx, tc.obj, tc.org)

			if len(tc.wantErrs) > 0 {
				for _, err := range tc.wantErrs {
					var gotErr string
					if resp != nil {
						gotErr = resp.Error()
					}
					if !strings.Contains(gotErr, err) {
						t.Errorf("Expected failure containing %q got %q", err, gotErr)
					}
				}
			} else if resp != nil {
				t.Errorf("Expected no error, got %q", resp.Error())
			}
		})
	}
}
