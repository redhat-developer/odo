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

package propagation

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	_ "knative.dev/pkg/metrics/testing"
)

var (
	sampled    trace.TraceOptions = 1
	notSampled trace.TraceOptions = 1

	traceID = trace.TraceID{99, 108, 97, 101, 114, 115, 105, 103, 104, 116, 101, 100, 110, 101, 115, 115}
	spanID  = trace.SpanID{107, 110, 97, 116, 105, 118, 101, 0}

	tracePropagators = []propagation.HTTPFormat{
		&b3.HTTPFormat{},
		&tracecontext.HTTPFormat{},
		both,
	}
)

var both = &HTTPFormatSequence{
	Ingress: []propagation.HTTPFormat{
		&tracecontext.HTTPFormat{},
		&b3.HTTPFormat{},
	},
	Egress: []propagation.HTTPFormat{
		&tracecontext.HTTPFormat{},
		&b3.HTTPFormat{},
	},
}

func TestSpanContextFromRequest(t *testing.T) {
	testCases := map[string]struct {
		sc *trace.SpanContext
	}{
		"no incoming trace": {},
		"not sampled": {
			sc: &trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: notSampled,
			},
		},
		"sampled": {
			sc: &trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: sampled,
			},
		},
	}

	for _, tracePropagator := range tracePropagators {
		for n, tc := range testCases {
			t.Run(fmt.Sprintf("%T-%s", tracePropagator, n), func(t *testing.T) {
				r := &http.Request{}
				r.Header = http.Header{}
				if tc.sc != nil {
					tracePropagator.SpanContextToRequest(*tc.sc, r)
				}
				// Check we extract the correct SpanContext with both the original and the
				// 'both' propagators.
				for _, extractFormat := range []propagation.HTTPFormat{tracePropagator, both} {
					actual, ok := extractFormat.SpanContextFromRequest(r)
					if tc.sc == nil {
						if ok {
							t.Errorf("Expected no span context using %T, found %v", extractFormat, actual)
						}
						continue
					}
					if diff := cmp.Diff(*tc.sc, actual); diff != "" {
						t.Errorf("Unexpected span context using %T (-want +got): %s", extractFormat, diff)
					}
				}
			})
		}
	}
}

func TestSpanContextToRequest(t *testing.T) {
	testCases := map[string]struct {
		sc *trace.SpanContext
	}{
		"no incoming trace": {},
		"not sampled": {
			sc: &trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: notSampled,
			},
		},
		"sampled": {
			sc: &trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: sampled,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			if tc.sc != nil {
				// Apply using the TraceContextB3 propagator.
				both.SpanContextToRequest(*tc.sc, r)
			}
			// Verify that we extract the correct SpanContext with all three formats.
			for _, tracePropagator := range tracePropagators {
				actual, ok := tracePropagator.SpanContextFromRequest(r)
				if tc.sc == nil {
					if ok {
						t.Errorf("Expected no span context using %T, found %v", tracePropagator, actual)
					}
					continue
				}
				if diff := cmp.Diff(*tc.sc, actual); diff != "" {
					t.Errorf("Unexpected span context using %T (-want +got): %s", tracePropagator, diff)
				}
			}
		})
	}
}
