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
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"

	. "knative.dev/pkg/logging/testing"
	"knative.dev/pkg/metrics/metricstest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				BackendDestinationKey:   "unsupported",
				stackdriverProjectIDKey: testProj,
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: `unsupported metrics backend value "unsupported"`,
	}, {
		name: "emptyDomain",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(prometheus),
			},
			Domain:    "",
			Component: testComponent,
		},
		expectedErr: "metrics domain cannot be empty",
	}, {
		name: "invalidComponent",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(openCensus),
			},
			Domain:    servingDomain,
			Component: "",
		},
		expectedErr: "metrics component name cannot be empty",
	}, {
		name: "invalidReportingPeriod",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(openCensus),
				reportingPeriodKey:    "test",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "invalid " + reportingPeriodKey + ` value "test"`,
	}, {
		name: "invalidOpenCensusSecuritySetting",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(openCensus),
				collectorSecureKey:    "yep",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "invalid " + collectorSecureKey + ` value "yep"`,
	}, {
		name: "invalidAllowStackdriverCustomMetrics",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey:            string(stackdriver),
				allowStackdriverCustomMetricsKey: "test",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErr: "invalid " + allowStackdriverCustomMetricsKey + ` value "test"`,
	}, {
		name: "tooSmallPrometheusPort",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(prometheus),
			},
			Domain:         servingDomain,
			Component:      testComponent,
			PrometheusPort: 1023,
		},
		expectedErr: "invalid port 1023, should be between 1024 and 65535",
	}, {
		name: "tooBigPrometheusPort",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(prometheus),
			},
			Domain:         servingDomain,
			Component:      testComponent,
			PrometheusPort: 65536,
		},
		expectedErr: "invalid port 65536, should be between 1024 and 65535",
	}}

	successTests = []struct {
		name                string
		ops                 ExporterOptions
		expectedConfig      metricsConfig
		expectedNewExporter bool // Whether the config requires a new exporter compared to previous test case
	}{{
		name: "stackdriverProjectIDMissing",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(stackdriver),
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
			backendDestination: prometheus,
			reportingPeriod:    5 * time.Second,
			prometheusPort:     defaultPrometheusPort,
		},
		expectedNewExporter: true,
	}, {
		name: "validStackdriver",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey:     string(stackdriver),
				stackdriverProjectIDKey:   anotherProj,
				stackdriverGCPLocationKey: "us-west1",
				stackdriverClusterNameKey: "cluster",
				stackdriverUseSecretKey:   "true",
			},
			Domain:    servingDomain,
			Component: testComponent,
			Secrets: fakeSecretList(corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      StackdriverSecretNameDefault,
					Namespace: StackdriverSecretNamespaceDefault,
				},
			}).Get,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      StackdriverSecretNameDefault,
					Namespace: StackdriverSecretNamespaceDefault,
				},
			},
		},
		expectedNewExporter: true,
	}, {
		name: "validPartialStackdriver",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey:     string(stackdriver),
				stackdriverProjectIDKey:   anotherProj,
				stackdriverClusterNameKey: "cluster",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey: string(openCensus),
				collectorAddressKey:   "external-svc:55678",
				collectorSecureKey:    "true",
			},
			Domain:    servingDomain,
			Component: testComponent,
			Secrets: fakeSecretList(corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "opencensus",
				},
				Data: map[string][]byte{
					"client-cert.pem": {},
					"client-key.pem":  {},
				},
			}).Get,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: openCensus,
			collectorAddress:   "external-svc:55678",
			requireSecure:      true,
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "opencensus",
				},
				Data: map[string][]byte{
					"client-cert.pem": {},
					"client-key.pem":  {},
				},
			},
		},
		expectedNewExporter: true,
	}, {
		name: "validPrometheus",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey: string(prometheus),
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
			reportingPeriod:    5 * time.Second,
			prometheusPort:     defaultPrometheusPort,
		},
		expectedNewExporter: true,
	}, {
		name: "validCapitalStackdriver",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey:   "Stackdriver",
				stackdriverProjectIDKey: testProj,
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey: string(prometheus),
				reportingPeriodKey:    "12",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
			reportingPeriod:    12 * time.Second,
			prometheusPort:     defaultPrometheusPort,
		},
		expectedNewExporter: true,
	}, {
		name: "overriddenReportingPeriodStackdriver",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey:   string(stackdriver),
				stackdriverProjectIDKey: "test2",
				reportingPeriodKey:      "7",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey:   string(stackdriver),
				stackdriverProjectIDKey: "test2",
				reportingPeriodKey:      "3",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey: string(prometheus),
				reportingPeriodKey:    "",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
			reportingPeriod:    5 * time.Second,
			prometheusPort:     defaultPrometheusPort,
		},
		expectedNewExporter: true,
	}, {
		name: "emptyReportingPeriodStackdriver",
		ops: ExporterOptions{
			ConfigMap: map[string]string{
				BackendDestinationKey:   string(stackdriver),
				stackdriverProjectIDKey: "test2",
				reportingPeriodKey:      "",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey:            string(stackdriver),
				stackdriverProjectIDKey:          "test2",
				reportingPeriodKey:               "",
				allowStackdriverCustomMetricsKey: "true",
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey:               string(stackdriver),
				stackdriverProjectIDKey:             "test2",
				reportingPeriodKey:                  "",
				stackdriverCustomMetricSubDomainKey: customSubDomain,
			},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
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
				BackendDestinationKey: string(prometheus),
			},
			Domain:         servingDomain,
			Component:      testComponent,
			PrometheusPort: 9091,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
			reportingPeriod:    5 * time.Second,
			prometheusPort:     9091,
		},
		expectedNewExporter: true,
	}}
)

func successTestsInit() {
	SetStackdriverSecretLocation(StackdriverSecretNameDefault, StackdriverSecretNamespaceDefault)
}

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

	successTestsInit()
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
	successTests := []struct {
		name           string
		varName        string
		varValue       string
		ops            ExporterOptions
		expectedConfig metricsConfig
	}{{
		name:     "Stackdriver backend from env, no config",
		varName:  defaultBackendEnvName,
		varValue: string(stackdriver),
		ops: ExporterOptions{
			ConfigMap: map[string]string{},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                stackdriver,
			reportingPeriod:                   60 * time.Second,
			isStackdriverBackend:              true,
			stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
			stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
		},
	}, {
		name:     "Stackdriver backend from env, Prometheus backend from config",
		varName:  defaultBackendEnvName,
		varValue: string(stackdriver),
		ops: ExporterOptions{
			ConfigMap: map[string]string{BackendDestinationKey: string(prometheus)},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
			reportingPeriod:    5 * time.Second,
			prometheusPort:     defaultPrometheusPort,
		},
	}, {
		name:     "PrometheusPort from env",
		varName:  prometheusPortEnvName,
		varValue: "9999",
		ops: ExporterOptions{
			ConfigMap: map[string]string{},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
			reportingPeriod:    5 * time.Second,
			prometheusPort:     9999,
		},
	}}

	failureTests := []struct {
		name                string
		varName             string
		varValue            string
		ops                 ExporterOptions
		expectedErrContains string
	}{{
		name:     "Invalid PrometheusPort from env",
		varName:  prometheusPortEnvName,
		varValue: strconv.Itoa(math.MaxUint16 + 1),
		ops: ExporterOptions{
			ConfigMap: map[string]string{},
			Domain:    servingDomain,
			Component: testComponent,
		},
		expectedErrContains: "value out of range",
	}}

	for _, test := range successTests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(test.varName, test.varValue)
			defer os.Unsetenv(test.varName)

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

	for _, test := range failureTests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(test.varName, test.varValue)
			defer os.Unsetenv(test.varName)

			defer ClearAll()
			mc, err := createMetricsConfig(test.ops, TestLogger(t))
			if mc != nil {
				t.Errorf("Wanted no config, got %v", mc)
			}
			if err == nil || !strings.Contains(err.Error(), test.expectedErrContains) {
				t.Errorf("Wanted err to contain: %q, got: %v", test.expectedErrContains, err)
			}
		})
	}
}

func TestIsNewExporterRequiredFromNilConfig(t *testing.T) {
	setCurMetricsConfig(nil)
	successTestsInit()
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
			backendDestination: prometheus,
		},
		newConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: prometheus,
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
			backendDestination:                stackdriver,
			reportingPeriod:                   60 * time.Second,
			isStackdriverBackend:              true,
			stackdriverMetricTypePrefix:       path.Join(servingDomain, testComponent),
			stackdriverCustomMetricTypePrefix: path.Join(customMetricTypePrefix, defaultCustomMetricSubDomain, testComponent),
		},
		newConfig: metricsConfig{
			domain:                            servingDomain,
			component:                         testComponent,
			backendDestination:                prometheus,
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
			backendDestination: stackdriver,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID: "proj1",
			},
		},
		newConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: stackdriver,
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
			backendDestination: stackdriver,
			stackdriverClientConfig: StackdriverClientConfig{
				ProjectID:   testProj,
				ClusterName: "cluster1",
			},
		},
		newConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: stackdriver,
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
	successTestsInit()
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
	successTestsInit()
	for _, test := range successTests[1:] {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			opts := ExporterOptions{
				Component:      test.ops.Component,
				Domain:         test.ops.Domain,
				PrometheusPort: test.ops.PrometheusPort,
				Secrets:        test.ops.Secrets,
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
			stackdriverProjectIDKey:   "project",
			stackdriverGCPLocationKey: "us-west1",
			stackdriverClusterNameKey: "cluster",
			stackdriverUseSecretKey:   "true",
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
			stackdriverProjectIDKey:   "project",
			stackdriverGCPLocationKey: "us-west1",
			stackdriverClusterNameKey: "cluster",
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
				BackendDestinationKey: string(prometheus),
			},
			servedCounter: 1,
			statCounter:   1,
		},
		"stackdriver with custom metrics": {
			opts: map[string]string{
				BackendDestinationKey:            string(stackdriver),
				allowStackdriverCustomMetricsKey: "true",
			},
			servedCounter: 1,
			statCounter:   1,
		},
		"stackdriver no custom metrics": {
			opts: map[string]string{
				BackendDestinationKey: string(stackdriver),
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
			err = RegisterResourceView(v...)
			if err != nil {
				t.Errorf("Failed to register %+v in stats backend: %v", v, err)
			}
			defer UnregisterResourceView(v...)

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
