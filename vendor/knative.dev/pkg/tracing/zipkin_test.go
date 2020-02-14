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

	"github.com/google/go-cmp/cmp"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinreporter "github.com/openzipkin/zipkin-go/reporter"
	reporterrecorder "github.com/openzipkin/zipkin-go/reporter/recorder"
	. "knative.dev/pkg/tracing"
	"knative.dev/pkg/tracing/config"
)

func TestOpenCensusTracerApplyConfig(t *testing.T) {
	tcs := []struct {
		name          string
		cfg           config.Config
		expect        *config.Config
		reporterError bool
	}{{
		name: "Disabled config",
		cfg: config.Config{
			Backend: config.None,
		},
		expect: nil,
	}, {
		name: "Endpoint specified",
		cfg: config.Config{
			Backend:        config.Zipkin,
			ZipkinEndpoint: "test-endpoint:1234",
		},
		expect: &config.Config{
			Backend:        config.Zipkin,
			ZipkinEndpoint: "test-endpoint:1234",
		},
	}}

	endpoint, _ := openzipkin.NewEndpoint("test", "localhost:1234")
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var gotCfg *config.Config
			reporter := reporterrecorder.NewReporter()
			oct := NewOpenCensusTracer(WithZipkinExporter(func(cfg *config.Config) (zipkinreporter.Reporter, error) {
				gotCfg = cfg
				if tc.reporterError {
					return nil, errors.New("Induced reporter factory error")
				}
				return reporter, nil
			}, endpoint))

			if err := oct.ApplyConfig(&tc.cfg); (err == nil) == tc.reporterError {
				t.Errorf("Failed to apply config: %v", err)
			}
			if diff := cmp.Diff(gotCfg, tc.expect); diff != "" {
				t.Errorf("Got tracer config (-want, +got) = %v", diff)
			}

			oct.Finish()
		})
	}
}
