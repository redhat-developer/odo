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

package tracing

import (
	"crypto/rand"
	"testing"

	"go.opencensus.io/trace"
	"knative.dev/pkg/tracing/config"
)

func TestCreateOCTConfig(t *testing.T) {
	tcs := []struct {
		name   string
		cfg    config.Config
		expect trace.Config
	}{{
		name: "Default",
		cfg:  config.Config{},
		expect: trace.Config{
			DefaultSampler: trace.NeverSample(),
		},
	}, {
		name: "Disabled",
		cfg: config.Config{
			Backend: config.None,
		},
		expect: trace.Config{
			DefaultSampler: trace.NeverSample(),
		},
	}, {
		name: "Debug",
		cfg: config.Config{
			Backend: config.Stackdriver,
			Debug:   true,
		},
		expect: trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		},
	}, {
		name: "Debug disabled",
		cfg: config.Config{
			Backend: config.None,
			Debug:   true,
		},
		expect: trace.Config{
			DefaultSampler: trace.NeverSample(),
		},
	}, {
		name: "percent sampler",
		cfg: config.Config{
			Backend:    config.Zipkin,
			Debug:      false,
			SampleRate: 0.5,
		},
		expect: trace.Config{
			DefaultSampler: trace.ProbabilitySampler(0.5),
		},
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			octCfg := createOCTConfig(&tc.cfg)

			// Create 100 traceIDs and make sure our expected sampler samples the same as what we get
			for i := 0; i < 100; i++ {
				spanID := make([]byte, 8)
				rand.Read(spanID)
				param := trace.SamplingParameters{}
				copy(param.SpanID[:], spanID)
				if tc.expect.DefaultSampler(param).Sample != octCfg.DefaultSampler(param).Sample {
					t.Errorf("Sampler for config did not match expected sample value for trace.")
				}
			}
		})
	}
}
