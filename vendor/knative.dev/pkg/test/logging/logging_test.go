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

package logging

import (
	"errors"
	"reflect"
	"testing"
)

type abcf struct {
	A int
	b string
	C *de
	F func()
}

type de struct {
	D string
	e float64
}

var someStruct abcf

func init() {
	someStruct = abcf{
		A: 42,
		b: "some string",
		C: &de{
			D: "hello world",
			e: 72.3,
		},
		F: InitializeLogger,
	}
}

func TestTLogger(legacy *testing.T) {
	verbosity = 5
	InitializeLogger()
	t, cancel := NewTLogger(legacy)
	defer cancel()

	var blank interface{} = &someStruct

	t.V(6).Info("Should not be printed")
	t.V(4).Info("Should be printed!")
	t.Run("A-Nice-Subtest", func(ts *TLogger) {
		ts.V(0).Info("This is pretty important; everyone needs to see it!",
			"some pointer", blank,
			"some number", 42.0)
		t.Run("A-Nested-Subtest", func(ts *TLogger) {
			ts.Parallel()
			ts.V(1).Info("I am visible!")
			ts.V(6).Info("I am invisible!")
			t.Collect("collected_non_error", "I'm not an error!")
		})
		t.Run("A-2nd-Nested-Subtest", func(ts *TLogger) {
			ts.Parallel()
			ts.V(1).Info("I am visible!")
			ts.V(6).Info("I am also invisible!")
		})
	})
	t.Run("Skipped", func(ts *TLogger) {
		ts.SkipNow()
	})
	t.ErrorIfErr(nil, "I won't fail because no error!")
	t.FatalIfErr(nil, "I won't fail because no error!")
	t = t.WithName("LongerName")
	t = t.WithValues("persistentKey", "persistentValue")
	t.Logf("Sadly still have to support %s", "LogF")
}

func TestTLoggerFailing(legacy *testing.T) {
	t, cancel := NewTLogger(legacy)
	defer cancel()
	t.dontFail = true

	t.Run("Failing1", func(ts *TLogger) {
		t.V(0).Info("dontFail values", "t", t.dontFail, "ts", ts.dontFail)
		t.Collect("collected_error", errors.New("collected"))
		ts.Error("I am an error", "hello", "world")
	})
	t.Run("Failing2", func(ts *TLogger) {
		ts.Fatal("I am a fatal error", "hello", "world")
	})
}

type errorWithRuntimeCheckValues struct {
	testName           string
	inputs             []interface{}
	expectedError      error
	expectedString     string
	expectedInterfaces []interface{}
}

func TestTLoggerInternals(legacy *testing.T) {
	verbosity = 2
	InitializeLogger()
	t, cancel := NewTLogger(legacy)
	defer cancel()

	tests := []errorWithRuntimeCheckValues{
		{"empty", nil, nil, "", nil},
		{"string with valid single key-value pair", []interface{}{"greetings!", "hello", "world"}, nil, "greetings!", []interface{}{"hello", "world"}},
		{"junk inputs", []interface{}{42}, nil, "unstructured error", t.interfacesToFields(42)},
	}

	for _, tt := range tests {
		t.Run("errorWithRuntimeCheck "+tt.testName, func(t *TLogger) {
			e, s, i := t.errorWithRuntimeCheck(tt.inputs...)
			if e != tt.expectedError {
				t.Error("error did not match", "got", e, "want", tt.expectedError)
			}
			if s != tt.expectedString {
				t.Error("string did not match", "got", s, "want", tt.expectedString)
			}
			if !reflect.DeepEqual(i, tt.expectedInterfaces) {
				t.Error("interfaces did not match", "got", i, "want", tt.expectedInterfaces)
			}
		})
	}

	if validateKeysAndValues(42, "whoops not string key before") {
		t.Error("Should not have accepted non-string key")
	}

	if !validateKeysAndValues("we like string keys", "any value is fine") {
		t.Error("Should have accepted string key")
	}

	input := []interface{}{4, 5, 6}
	things := t.interfacesToFields(input...)
	expected := []interface{}{"arg 0", 4, "arg 1", 5, "arg 2", 6}
	if !reflect.DeepEqual(things, expected) {
		t.Error("interfacesToFields() didn't give expected output", "input", input, "want", expected, "got", things)
	}

	t.Helper() // Doesn't do anything
}

func TestStructuredError(legacy *testing.T) {
	verbosity = 5
	InitializeLogger()
	t, cancel := NewTLogger(legacy)
	defer cancel()
	err := Error("Hello World", "key", "value", "current function", TestStructuredError, "deep struct", someStruct, "z", 4, "y", 3, "x", 2, "w", 1)
	t.Log(err.Error())
}
