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

package tracecontextb3

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	ocpropagation "go.opencensus.io/trace/propagation"

	_ "knative.dev/pkg/metrics/testing"
)

var (
	sampled trace.TraceOptions = 1

	traceID = trace.TraceID{99, 108, 97, 101, 114, 115, 105, 103, 104, 116, 101, 100, 110, 101, 115, 115}
	spanID  = trace.SpanID{107, 110, 97, 116, 105, 118, 101, 0}

	spanContext = trace.SpanContext{
		TraceID:      traceID,
		SpanID:       spanID,
		TraceOptions: sampled,
	}
)

func TestTraceContextB3Egress_Ingress(t *testing.T) {
	testCases := map[string]struct {
		ingress ocpropagation.HTTPFormat
	}{
		"traceContext": {
			ingress: &tracecontext.HTTPFormat{},
		},
		"b3": {
			ingress: &b3.HTTPFormat{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			tc.ingress.SpanContextToRequest(spanContext, r)

			assertFormatReadsSpanContext(t, r, TraceContextB3Egress)
		})
	}
}

func TestTraceContextB3Egress_Egress(t *testing.T) {
	testCases := map[string]struct {
		egress ocpropagation.HTTPFormat
	}{
		"traceContext": {
			egress: &tracecontext.HTTPFormat{},
		},
		"b3": {
			egress: &b3.HTTPFormat{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			TraceContextB3Egress.SpanContextToRequest(spanContext, r)

			assertFormatReadsSpanContext(t, r, tc.egress)
		})
	}
}

func TestTraceContextEgress_Ingress(t *testing.T) {
	testCases := map[string]struct {
		ingress ocpropagation.HTTPFormat
	}{
		"traceContext": {
			ingress: &tracecontext.HTTPFormat{},
		},
		"b3": {
			ingress: &b3.HTTPFormat{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			tc.ingress.SpanContextToRequest(spanContext, r)

			assertFormatReadsSpanContext(t, r, TraceContextEgress)
		})
	}
}

func TestTraceContextEgress_Egress(t *testing.T) {
	testCases := map[string]struct {
		egress ocpropagation.HTTPFormat
	}{
		"traceContext": {
			egress: &tracecontext.HTTPFormat{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			TraceContextEgress.SpanContextToRequest(spanContext, r)

			assertFormatReadsSpanContext(t, r, tc.egress)
		})
	}
}

func TestB3Egress_Ingress(t *testing.T) {
	testCases := map[string]struct {
		ingress ocpropagation.HTTPFormat
	}{
		"traceContext": {
			ingress: &tracecontext.HTTPFormat{},
		},
		"b3": {
			ingress: &b3.HTTPFormat{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			tc.ingress.SpanContextToRequest(spanContext, r)

			assertFormatReadsSpanContext(t, r, B3Egress)
		})
	}
}

func TestB3Egress_Egress(t *testing.T) {
	testCases := map[string]struct {
		egress ocpropagation.HTTPFormat
	}{
		"b3": {
			egress: &b3.HTTPFormat{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			B3Egress.SpanContextToRequest(spanContext, r)

			assertFormatReadsSpanContext(t, r, tc.egress)
		})
	}
}

func assertFormatReadsSpanContext(t *testing.T, r *http.Request, format ocpropagation.HTTPFormat) {
	sc, ok := format.SpanContextFromRequest(r)
	if !ok {
		t.Error("Expected to get the SpanContext")
	}
	if diff := cmp.Diff(spanContext, sc); diff != "" {
		t.Errorf("Unexpected SpanContext (-want +got): %s", diff)
	}
}
