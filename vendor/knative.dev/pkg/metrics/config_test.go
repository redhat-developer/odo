/*
Copyright 2018 The Knative Authors

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
package metrics

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"

	. "knative.dev/pkg/logging/testing"
	"knative.dev/pkg/metrics/metricstest"

	corev1 "k8s.io/api/core/v1"
)

// TODO UTs should move to eventing and serving, as appropriate.
// 	See https://github.com/knative/pkg/issues/608

const (
	servingDomain          = "knative.dev/serving"
	internalServingDomain  = "knative.dev/internal/serving"
	eventingDomain         = "knative.dev/eventing"
	internalEventingDomain = "knative.dev/internal/eventing"
	customSubDomain        = "test.domain"
	testComponent          = "testComponent"
	testProj               = "test-project"
	anotherProj            = "another-project"
)

var (
	errorTests = []struct {
		name        string
		ops         ExporterOptions
		expectedErr string
	}{{
		name: "empty config",
		ops: ExporterOptions{
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "metrics config map cannot be empty",
	}, {
		name: "unsupportedBackend",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination":    "unsupported",
				"metrics.stackdriver-project-id": testProj,
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "unsupported metrics backend value \"unsupported\"",
	}, {
		name: "emptyDomain",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination": "prometheus",
			},
			Domain:    "",
			Component: testComponent,
		},
		expectedErr: "metrics domain cannot be empty",
	}, {
		name: "invalidComponent",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination": "opencensus",
			},
			Domain:    servingDomain,
			Component: "",
		},
		expectedErr: "metrics component name cannot be empty",
	}, {
		name: "invalidReportingPeriod",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination":      "opencensus",
				"metrics.reporting-period-seconds": "test",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "invalid metrics.reporting-period-seconds value \"test\"",
	}, {
		name: "invalidOpenCensusSecuritySetting",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination":    "opencensus",
				"metrics.opencensus-require-tls": "yep",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "invalid metrics.opencensus-require-tls value \"yep\"",
	}, {
		name: "invalidAllowStackdriverCustomMetrics",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination":              "stackdriver",
				"metrics.allow-stackdriver-custom-metrics": "test",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "invalid metrics.allow-stackdriver-custom-metrics value \"test\"",
	}, {
		name: "tooSmallPrometheusPort",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination": "prometheus",
			},
			Domain:         servingDomain,
			Component:      testComponent,
			PrometheusPort: 1023,
		},
		expectedErr: "invalid port 1023, should between 1024 and 65535",
	}, {
		name: "tooBigPrometheusPort",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				"metrics.backend-destination": "prometheus",
			},
			Domain:         servingDomain,
			Component:      testComponent,
			PrometheusPort: 65536,
		},
		expectedErr: "invalid port 65536, should between 1024 and 65535",
	}}
	successTests = []struct {
		name                string
		ops                 ExporterOptions
		expectedConfig      metricsConfig
		expectedNewExporter bool // Whether the config requires a new exporter compared to previous test case
	}{
		// Note the first unit test is skipped in TestUpdateExporterFromConfigMap since
		// unit test does not have application default credentials.
		{
			name: "stackdriverProjectIDMissing",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination": "stackdriver",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
			},
			expectedNewExporter: true,
		}, {
			name: "backendKeyMissing",
			ops: ExporterOptions{
				ConfigMap: map[string]string{},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: Prometheus,
				reportingPeriod:    5 * time.Second,
				prometheusPort:     defaultPrometheusPort,
			},
			expectedNewExporter: true,
		}, {
			name: "validStackdriver",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "stackdriver",
					"metrics.stackdriver-project-id":   anotherProj,
					"metrics.stackdriver-gcp-location": "us-west1",
					"metrics.stackdriver-cluster-name": "cluster",
					"metrics.stackdriver-use-secret":   "true",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID:   anotherProj,
					GCPLocation: "us-west1",
					ClusterName: "cluster",
					UseSecret:   true,
				},
			},
			expectedNewExporter: true,
		}, {
			name: "validPartialStackdriver",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "stackdriver",
					"metrics.stackdriver-project-id":   anotherProj,
					"metrics.stackdriver-cluster-name": "cluster",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID:   anotherProj,
					ClusterName: "cluster",
				},
			},
			expectedNewExporter: true,
		}, {
			name: "validOpenCensusSettings",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":    "opencensus",
					"metrics.opencensus-address":     "external-svc:55678",
					"metrics.opencensus-require-tls": "true",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: OpenCensus,
				collectorAddress:   "external-svc:55678",
				requireSecure:      true,
			},
			expectedNewExporter: true,
		}, {
			name: "validPrometheus",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination": "prometheus",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: Prometheus,
				reportingPeriod:    5 * time.Second,
				prometheusPort:     defaultPrometheusPort,
			},
			expectedNewExporter: true,
		}, {
			name: "validCapitalStackdriver",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":    "Stackdriver",
					"metrics.stackdriver-project-id": testProj,
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID: testProj,
				},
			},
			expectedNewExporter: true,
		}, {
			name: "overriddenReportingPeriodPrometheus",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "prometheus",
					"metrics.reporting-period-seconds": "12",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: Prometheus,
				reportingPeriod:    12 * time.Second,
				prometheusPort:     defaultPrometheusPort,
			},
			expectedNewExporter: true,
		}, {
			name: "overriddenReportingPeriodStackdriver",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "stackdriver",
					"metrics.stackdriver-project-id":   "test2",
					"metrics.reporting-period-seconds": "7",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   7 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID: "test2",
				},
			},
			expectedNewExporter: true,
		}, {
			name: "overriddenReportingPeriodStackdriver2",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "stackdriver",
					"metrics.stackdriver-project-id":   "test2",
					"metrics.reporting-period-seconds": "3",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   3 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID: "test2",
				},
			},
		}, {
			name: "emptyReportingPeriodPrometheus",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "prometheus",
					"metrics.reporting-period-seconds": "",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: Prometheus,
				reportingPeriod:    5 * time.Second,
				prometheusPort:     defaultPrometheusPort,
			},
			expectedNewExporter: true,
		}, {
			name: "emptyReportingPeriodStackdriver",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":      "stackdriver",
					"metrics.stackdriver-project-id":   "test2",
					"metrics.reporting-period-seconds": "",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID: "test2",
				},
			},
			expectedNewExporter: true,
		}, {
			name: "allowStackdriverCustomMetric",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":              "stackdriver",
					"metrics.stackdriver-project-id":           "test2",
					"metrics.reporting-period-seconds":         "",
					"metrics.allow-stackdriver-custom-metrics": "true",
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID: "test2",
				},
			},
		}, {
			name: "allowStackdriverCustomMetric with subdomain",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination":                  "stackdriver",
					"metrics.stackdriver-project-id":               "test2",
					"metrics.reporting-period-seconds":             "",
					"metrics.stackdriver-custom-metrics-subdomain": customSubDomain,
				},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, customSubDomain, testComponent),
				stackdriverClientConfig: StackdriverClientConfig{
					ProjectID: "test2",
				},
			},
		}, {
			name: "overridePrometheusPort",
			ops: ExporterOptions{
				ConfigMap: map[string]string{
					"metrics.backend-destination": "prometheus",
				},
				Domain:         servingDomain,
				Component:      testComponent,
				PrometheusPort: 9091,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: Prometheus,
				reportingPeriod:    5 * time.Second,
				prometheusPort:     9091,
			},
			expectedNewExporter: true,
		}}
	envTests = []struct {
		name           string
		ops            ExporterOptions
		expectedConfig metricsConfig
	}{
		{
			name: "stackdriverFromEnv",
			ops: ExporterOptions{
				ConfigMap: map[string]string{},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:                            servingDomain,
				component:                         testComponent,
				backendDestination:                Stackdriver,
				reportingPeriod:                   60 * time.Second,
				isStackdriverBackend:              true,
				stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
				stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
			},
		}, {
			name: "validPrometheus",
			ops: ExporterOptions{
				ConfigMap: map[string]string{"metrics.backend-destination": "prometheus"},
				Domain:    servingDomain,
				Component: testComponent,
			},
			expectedConfig: metricsConfig{
				domain:             servingDomain,
				component:          testComponent,
				backendDestination: Prometheus,
				reportingPeriod:    5 * time.Second,
				prometheusPort:     defaultPrometheusPort,
			},
		}}
)

func TestGetMetricsConfig(t *testing.T) {
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			_, err := createMetricsConfig(test.ops, TestLogger(t))
			if err == nil || err.Error() != test.expectedErr {
				t.Errorf("Wanted err: %v, got: %v", test.expectedErr, err)
			}
		})
	}

	for _, test := range successTests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			mc, err := createMetricsConfig(test.ops, TestLogger(t))
			if err != nil {
				t.Errorf("Wanted valid config %v, got error %v", test.expectedConfig, err)
			}
			if diff := cmp.Diff(test.expectedConfig, *mc, cmp.AllowUnexported(*mc), cmpopts.IgnoreTypes(mc.recorder)); diff != "" {
				t.Errorf("Invalid config (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetMetricsConfig_fromEnv(t *testing.T) {
	os.Setenv(defaultBackendEnvName, "stackdriver")
	for _, test := range envTests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			mc, err := createMetricsConfig(test.ops, TestLogger(t))
			if err != nil {
				t.Errorf("Wanted valid config %v, got error %v", test.expectedConfig, err)
			}
			if diff := cmp.Diff(test.expectedConfig, *mc, cmp.AllowUnexported(*mc), cmpopts.IgnoreTypes(mc.recorder)); diff != "" {
				t.Errorf("Invalid config (-want +got):\n%s", diff)
			}
		})
	}
	os.Unsetenv(defaultBackendEnvName)
}

func TestIsNewExporterRequiredFromNilConfig(t *testing.T) {
	setCurMetricsConfig(nil)
	for _, test := range successTests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			mc, err := createMetricsConfig(test.ops, TestLogger(t))
			if err != nil {
				t.Errorf("Wanted valid config %v, got error %v", test.expectedConfig, err)
			}
			changed := isNewExporterRequired(mc)
			if changed != test.expectedNewExporter {
				t.Errorf("isMetricsConfigChanged=%v wanted %v", changed, test.expectedNewExporter)
			}
			setCurMetricsConfig(mc)
		})
	}
}

func TestIsNewExporterRequired(t *testing.T) {
	tests := []struct {
		name                string
		oldConfig           metricsConfig
		newConfig           metricsConfig
		newExporterRequired bool
	}{{
		name: "backendPrometheusChangeStackdriverClientConfig",
		oldConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Prometheus,
		},
		newConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Prometheus,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID:   testProj,
				ClusterName: "cluster",
			},
		},
		newExporterRequired: false,
	}, {
		name: "changeMetricsBackend",
		oldConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                Stackdriver,
			reportingPeriod:                   60 * time.Second,
			isStackdriverBackend:              true,
			stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
			stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
		},
		newConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                Prometheus,
			reportingPeriod:                   60 * time.Second,
			isStackdriverBackend:              true,
			stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
			stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
		},
		newExporterRequired: true,
	}, {
		name: "changeComponent",
		oldConfig: metricsConfig{
			domain:    servingDomain,
			component: "component1",
		},
		newConfig: metricsConfig{
			domain:    servingDomain,
			component: "component2",
		},
		newExporterRequired: false,
	}, {
		name: "backendStackdriverChangeProjectID",
		oldConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Stackdriver,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID: "proj1",
			},
		},
		newConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Stackdriver,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID: "proj2",
			},
		},
		newExporterRequired: true,
	}, {
		name: "backendStackdriverChangeStackdriverClientConfig",
		oldConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Stackdriver,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID:   testProj,
				ClusterName: "cluster1",
			},
		},
		newConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Stackdriver,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID:   testProj,
				ClusterName: "cluster2",
			},
		},
		newExporterRequired: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setCurMetricsConfig(&test.oldConfig)
			actualNewExporterRequired := isNewExporterRequired(&test.newConfig)
			if test.newExporterRequired != actualNewExporterRequired {
				t.Errorf("isNewExporterRequired returned incorrect value. Expected: [%v], Got: [%v]. Old config: [%v], New config: [%v]", test.newExporterRequired, actualNewExporterRequired, test.oldConfig, test.newConfig)
			}
		})
	}
}

func TestUpdateExporter(t *testing.T) {
	setCurMetricsConfig(nil)
	oldConfig := getCurMetricsConfig()
	for _, test := range successTests[1:] {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			UpdateExporter(test.ops, TestLogger(t))
			mConfig := getCurMetricsConfig()
			if mConfig == oldConfig {
				t.Error("Expected metrics config change")
			}
			if diff := cmp.Diff(test.expectedConfig, *mConfig, cmp.AllowUnexported(*mConfig), cmpopts.IgnoreTypes(mConfig.recorder)); diff != "" {
				t.Errorf("Invalid config (-want +got):\n%s", diff)
			}
			oldConfig = mConfig
		})
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			UpdateExporter(test.ops, TestLogger(t))
			mConfig := getCurMetricsConfig()
			if mConfig != oldConfig {
				t.Error("mConfig should not change")
			}
		})
	}
}

func TestUpdateExporterFromConfigMapWithOpts(t *testing.T) {
	setCurMetricsConfig(nil)
	oldConfig := getCurMetricsConfig()
	for _, test := range successTests[1:] {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			opts := ExporterOptions{
				Component:      test.ops.Component,
				Domain:         test.ops.Domain,
				PrometheusPort: test.ops.PrometheusPort,
			}
			updateFunc, err := UpdateExporterFromConfigMapWithOpts(opts, TestLogger(t))
			if err != nil {
				t.Errorf("failed to call UpdateExporterFromConfigMapWithOpts: %v", err)
			}
			updateFunc(&corev1.ConfigMap{Data: test.ops.ConfigMap})
			mConfig := getCurMetricsConfig()
			if mConfig == oldConfig {
				t.Error("Expected metrics config change")
			}
			if diff := cmp.Diff(test.expectedConfig, *mConfig, cmp.AllowUnexported(*mConfig), cmpopts.IgnoreTypes(mConfig.recorder)); diff != "" {
				t.Errorf("Invalid config (-want +got):\n%s", diff)
			}
			oldConfig = mConfig
		})
	}

	t.Run("ConfigMapSetErr", func(t *testing.T) {
		defer ClearAll()
		opts := ExporterOptions{
			Component:      testComponent,
			Domain:         servingDomain,
			PrometheusPort: defaultPrometheusPort,
			ConfigMap:      map[string]string{"some": "data"},
		}
		_, err := UpdateExporterFromConfigMapWithOpts(opts, TestLogger(t))
		if err == nil {
			t.Error("got err=nil want err")
		}
	})

	t.Run("MissingComponentErr", func(t *testing.T) {
		defer ClearAll()
		opts := ExporterOptions{
			Component:      "",
			Domain:         servingDomain,
			PrometheusPort: defaultPrometheusPort,
		}
		_, err := UpdateExporterFromConfigMapWithOpts(opts, TestLogger(t))
		if err == nil {
			t.Error("got err=nil want err")
		}
	})
}

func TestUpdateExporter_doesNotCreateExporter(t *testing.T) {
	setCurMetricsConfig(nil)
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			UpdateExporter(test.ops, TestLogger(t))
			mConfig := getCurMetricsConfig()
			if mConfig != nil {
				t.Error("mConfig should not be created")
			}
		})
	}
}

func TestMetricsOptions(t *testing.T) {
	testCases := map[string]struct {
		opts    *ExporterOptions
		want    string
		wantErr string
	}{
		"nil": {
			opts:    nil,
			want:    "",
			wantErr: "json options string is empty",
		},
		"happy": {
			opts: &ExporterOptions{
				Domain:         "domain",
				Component:      "component",
				PrometheusPort: 9090,
				ConfigMap: map[string]string{
					"foo":   "bar",
					"boosh": "kakow",
				},
			},
			want: `{"Domain":"domain","Component":"component","PrometheusPort":9090,"ConfigMap":{"boosh":"kakow","foo":"bar"}}`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			jsonOpts, err := MetricsOptionsToJson(tc.opts)
			if err != nil {
				t.Errorf("error while converting metrics config to json: %v", err)
			}
			// Test to json.
			{
				want := tc.want
				got := jsonOpts
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected (-want, +got) = %v", diff)
					t.Log(got)
				}
			}
			// Test to options.
			{
				want := tc.opts
				got, gotErr := JsonToMetricsOptions(jsonOpts)

				if gotErr != nil {
					if diff := cmp.Diff(tc.wantErr, gotErr.Error()); diff != "" {
						t.Errorf("unexpected err (-want, +got) = %v", diff)
					}
				} else if tc.wantErr != "" {
					t.Errorf("expected err %v", tc.wantErr)
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected (-want, +got) = %v", diff)
					t.Log(got)
				}
			}
		})
	}
}

func TestNewStackdriverConfigFromMap(t *testing.T) {
	tests := []struct {
		name           string
		stringMap      map[string]string
		expectedConfig StackdriverClientConfig
	}{{
		name: "fullSdConfig",
		stringMap: map[string]string{
			"metrics.stackdriver-project-id":   "project",
			"metrics.stackdriver-gcp-location": "us-west1",
			"metrics.stackdriver-cluster-name": "cluster",
			"metrics.stackdriver-use-secret":   "true",
		},
		expectedConfig: StackdriverClientConfig{
			ProjectID:   "project",
			GCPLocation: "us-west1",
			ClusterName: "cluster",
			UseSecret:   true,
		},
	}, {
		name:           "emptySdConfig",
		stringMap:      map[string]string{},
		expectedConfig: StackdriverClientConfig{},
	}, {
		name: "partialSdConfig",
		stringMap: map[string]string{
			"metrics.stackdriver-project-id":   "project",
			"metrics.stackdriver-gcp-location": "us-west1",
			"metrics.stackdriver-cluster-name": "cluster",
		},
		expectedConfig: StackdriverClientConfig{
			ProjectID:   "project",
			GCPLocation: "us-west1",
			ClusterName: "cluster",
			UseSecret:   false,
		},
	}, {
		name:           "nil",
		stringMap:      nil,
		expectedConfig: StackdriverClientConfig{},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewStackdriverClientConfigFromMap(test.stringMap)
			if test.expectedConfig != *c {
				t.Errorf("Incorrect stackdriver config. Expected: [%v], Got: [%v]", test.expectedConfig, *c)
			}
		})
	}
}

// TODO(evankanderson): Move the Stackdriver / Record patching out of config.go
func TestStackdriverRecord(t *testing.T) {
	testCases := map[string]struct {
		opts          map[string]string
		servedCounter int64
		statCounter   int64
	}{
		"non-stackdriver": {
			opts: map[string]string{
				"metrics.backend-destination": "prometheus",
			},
			servedCounter: 1,
			statCounter:   1,
		},
		"stackdriver with custom metrics": {
			opts: map[string]string{
				"metrics.backend-destination":              "stackdriver",
				"metrics.allow-stackdriver-custom-metrics": "true",
			},
			servedCounter: 1,
			statCounter:   1,
		},
		"stackdriver no custom metrics": {
			opts: map[string]string{
				"metrics.backend-destination": "stackdriver",
			},
			servedCounter: 1,
			statCounter:   0,
		},
	}

	servedCount := stats.Int64("request_count", "Number of requests", stats.UnitNone)
	statCount := stats.Int64("stat_errors", "Number of errors calling stat", stats.UnitNone)
	emptyTags := map[string]string{}

	for name, data := range testCases {
		t.Run(name, func(t *testing.T) {
			defer ClearAll()
			opts := ExporterOptions{
				ConfigMap: data.opts,
				Domain:    "knative.dev/internal/serving",
				Component: "activator",
			}
			mc, err := createMetricsConfig(opts, TestLogger(t))
			if err != nil {
				t.Errorf("Expected valid config %+v, got error: %v\n", opts, err)
			}
			setCurMetricsConfig(mc)
			ctx := context.Background()
			v := []*view.View{
				{Measure: servedCount, Aggregation: view.Count()},
				{Measure: statCount, Aggregation: view.Count()},
			}
			err = view.Register(v...)
			if err != nil {
				t.Errorf("Failed to register %+v in stats backend: %v", v, err)
			}
			defer view.Unregister(v...)

			// Try recording each metric and checking the result.
			Record(ctx, servedCount.M(1))
			metricstest.CheckCountData(t, servedCount.Name(), emptyTags, data.servedCounter)

			Record(ctx, statCount.M(1))
			if data.statCounter != 0 {
				metricstest.CheckCountData(t, statCount.Name(), emptyTags, data.statCounter)
			} else {
				metricstest.CheckStatsNotReported(t, statCount.Name())
			}
		})
	}
}
