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

package tracing_test

import (
	"errors"
	"testing"

	. "knative.dev/pkg/tracing"
	"knative.dev/pkg/tracing/config"
	. "knative.dev/pkg/tracing/testing"
)

func TestOpenCensusTracerGlobalLifecycle(t *testing.T) {
	reporter, co := FakeZipkinExporter()
	defer reporter.Close()
	oct := NewOpenCensusTracer(co)
	// Apply a config to make us the global OCT
	if err := oct.ApplyConfig(&config.Config{}); err != nil {
		t.Fatalf("Failed to ApplyConfig on tracer: %v", err)
	}

	otherOCT := NewOpenCensusTracer(co)
	if err := otherOCT.ApplyConfig(&config.Config{}); err == nil {
		t.Fatalf("Expected error when applying config to second OCT.")
	}

	if err := oct.Finish(); err != nil {
		t.Fatalf("Failed to finish OCT: %v", err)
	}

	if err := otherOCT.ApplyConfig(&config.Config{}); err != nil {
		t.Fatalf("Failed to ApplyConfig on OtherOCT after finishing OCT: %v", err)
	}
	otherOCT.Finish()
}

func TestOpenCensusTraceApplyConfigFailingConfigOption(t *testing.T) {
	coErr := errors.New("configOption error")
	oct := NewOpenCensusTracer(func(c *config.Config) error {
		if c != nil {
			return coErr
		}
		return nil
	})
	if err := oct.ApplyConfig(&config.Config{}); err != coErr {
		t.Errorf("Expected error not seen. Got %q. Want %q", err, coErr)
	}
	if err := oct.Finish(); err != nil {
		t.Errorf("Unexpected error Finishing: %q", err)
	}
}

func TestOpenCensusTraceFinishFailingConfigOption(t *testing.T) {
	coErr := errors.New("configOption error")
	errToReturn := coErr
	oct := NewOpenCensusTracer(func(c *config.Config) error {
		if c == nil {
			// We need finish to work on the second try, otherwise we have mutated global state. So,
			// make sure that next run through, the returned error is nil.
			e := errToReturn
			errToReturn = nil
			return e
		}
		return nil
	})
	if err := oct.ApplyConfig(&config.Config{}); err != nil {
		t.Errorf("Unexpected error Applying Config: %q", err)
	}
	if err := oct.Finish(); err != coErr {
		t.Errorf("Expected error not seen. Got %q. Want %q", err, coErr)
	}
	if err := oct.Finish(); err != nil {
		t.Errorf("Unexpected error on second Finish (global state mutated, other tests may fail oddly): %q", err)
	}
}
