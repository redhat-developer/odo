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

package test

import (
	"flag"
	"sync"
	"testing"
)

func TestE2eFlags(t *testing.T) {
	v := flag.Lookup("v")
	if v == nil {
		printFlags()
		t.Fatal("Could not find 'v' flag; klog was not init-ed")
	}
	if v.Value.String() != klogDefaultLogLevel {
		t.Fatalf("Either '%s' was passed in for v or v is not being initialized properly. The former is a test limitation, but the latter is an actual problem", v.Value.String())
	}

	flagsAreUninitialized := false
	flagsSetupOnce.Do(func() {
		flagsAreUninitialized = true
	})
	if !flagsAreUninitialized {
		t.Error("SetupLoggingFlags should not have been called at package initialization stage")
	}

	flagTests := []struct {
		testCase    string
		logVerbose  bool
		nonDefaultV string
		expectedV   string
	}{
		{
			testCase:    "Check default case with logverbose off",
			logVerbose:  false,
			nonDefaultV: "",
			expectedV:   klogDefaultLogLevel,
		},
		{
			testCase:    "Check default case with logverbose on",
			logVerbose:  true,
			nonDefaultV: "",
			expectedV:   "8",
		},
		{
			testCase:    "Check default case with logverbose on and v set to non-default value",
			logVerbose:  true,
			nonDefaultV: "5",
			expectedV:   "5",
		},
	}

	alsologtostderr := flag.Lookup("alsologtostderr").Value.String()
	if "true" != alsologtostderr {
		t.Errorf("alsologtostderr = '%s', want: 'true'\n", alsologtostderr)
	}

	for _, tc := range flagTests {
		t.Run(tc.testCase, func(t *testing.T) {
			flagsSetupOnce = &sync.Once{}
			Flags.LogVerbose = tc.logVerbose
			if tc.nonDefaultV != "" {
				flag.Set("v", tc.nonDefaultV)
			}
			SetupLoggingFlags()
			v := klogFlags.Lookup("v").Value.String()
			if tc.expectedV != v {
				t.Errorf("v = '%s', want: '%s'\n", v, tc.expectedV)
			}
		})
	}
}
