/*
Copyright 2020 The Knative Authors

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

package configmap

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
)

type testConfig struct {
	str string
	boo bool
	i32 int32
	i64 int64
	u32 uint32
	f64 float64
	dur time.Duration
	set sets.String
	qua *resource.Quantity
}

func TestParse(t *testing.T) {
	fiveHundredM := resource.MustParse("500m")
	tests := []struct {
		name      string
		conf      testConfig
		data      map[string]string
		want      testConfig
		expectErr bool
	}{{
		name: "all good",
		data: map[string]string{
			"test-string":   "foo.bar",
			"test-bool":     "true",
			"test-int32":    "1",
			"test-int64":    "2",
			"test-uint32":   "3",
			"test-float64":  "1.0",
			"test-duration": "1m",
			"test-set":      "a,b,c",
			"test-quantity": "500m",
		},
		want: testConfig{
			str: "foo.bar",
			boo: true,
			i32: 1,
			i64: 2,
			u32: 3,
			f64: 1.0,
			dur: time.Minute,
			set: sets.NewString("a", "b", "c"),
			qua: &fiveHundredM,
		},
	}, {
		name: "respect defaults",
		conf: testConfig{
			str: "foo.bar",
			boo: true,
			i32: 1,
			i64: 2,
			f64: 1.0,
			dur: time.Minute,
			qua: &fiveHundredM,
		},
		want: testConfig{
			str: "foo.bar",
			boo: true,
			i32: 1,
			i64: 2,
			f64: 1.0,
			dur: time.Minute,
			qua: &fiveHundredM,
		},
	}, {
		name: "bool defaults to false",
		data: map[string]string{
			"test-bool": "foo",
		},
		want: testConfig{
			boo: false,
		},
	}, {
		name: "int32 error",
		data: map[string]string{
			"test-int32": "foo",
		},
		expectErr: true,
	}, {
		name: "int64 error",
		data: map[string]string{
			"test-int64": "foo",
		},
		expectErr: true,
	}, {
		name: "uint32 error",
		data: map[string]string{
			"test-uint32": "foo",
		},
		expectErr: true,
	}, {
		name: "float64 error",
		data: map[string]string{
			"test-float64": "foo",
		},
		expectErr: true,
	}, {
		name: "duration error",
		data: map[string]string{
			"test-duration": "foo",
		},
		expectErr: true,
	}, {
		name: "quantity error",
		data: map[string]string{
			"test-quantity": "foo",
		},
		expectErr: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Parse(test.data,
				AsString("test-string", &test.conf.str),
				AsBool("test-bool", &test.conf.boo),
				AsInt32("test-int32", &test.conf.i32),
				AsInt64("test-int64", &test.conf.i64),
				AsUint32("test-uint32", &test.conf.u32),
				AsFloat64("test-float64", &test.conf.f64),
				AsDuration("test-duration", &test.conf.dur),
				AsStringSet("test-set", &test.conf.set),
				AsQuantity("test-quantity", &test.conf.qua),
			); (err == nil) == test.expectErr {
				t.Fatal("Failed to parse data:", err)
			}

			if !cmp.Equal(test.conf, test.want, cmp.AllowUnexported(testConfig{})) {
				t.Fatalf("parsed = %v, want %v", test.conf, test.want)
			}
		})
	}
}
