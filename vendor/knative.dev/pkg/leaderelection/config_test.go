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

package leaderelection

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/kmeta"
)

func okConfig() *Config {
	return &Config{
		ResourceLock:      "leases",
		Buckets:           1,
		LeaseDuration:     15 * time.Second,
		RenewDeadline:     10 * time.Second,
		RetryPeriod:       2 * time.Second,
		EnabledComponents: sets.NewString(),
	}
}

func okData() map[string]string {
	return map[string]string{
		"resourceLock": "leases",
		"buckets":      "1",
		// values in this data come from the defaults suggested in the
		// code:
		// https://github.com/kubernetes/client-go/blob/kubernetes-1.16.0/tools/leaderelection/leaderelection.go
		"leaseDuration":     "15s",
		"renewDeadline":     "10s",
		"retryPeriod":       "2s",
		"enabledComponents": "controller",
	}
}

func TestNewConfigMapFromData(t *testing.T) {
	cases := []struct {
		name     string
		data     map[string]string
		expected *Config
		err      error
	}{{
		name: "disabled but OK config",
		data: func() map[string]string {
			data := okData()
			delete(data, "enabledComponents")
			return data
		}(),
		expected: okConfig(),
	}, {
		name: "OK config - controller enabled",
		data: okData(),
		expected: func() *Config {
			config := okConfig()
			config.EnabledComponents.Insert("controller")
			return config
		}(),
	}, {
		name: "OK config - controller enabled with multiple buckets",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"buckets": "5",
		}),
		expected: func() *Config {
			config := okConfig()
			config.EnabledComponents.Insert("controller")
			config.Buckets = 5
			return config
		}(),
	}, {
		name: "invalid resourceLock",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"resourceLock": "flarps",
		}),
		err: errors.New(`resourceLock: invalid value "flarps": valid values are "leases","configmaps","endpoints"`),
	}, {
		name: "invalid leaseDuration",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"leaseDuration": "flops",
		}),
		err: errors.New(`failed to parse "leaseDuration": time: invalid duration flops`),
	}, {
		name: "invalid renewDeadline",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"renewDeadline": "flops",
		}),
		err: errors.New(`failed to parse "renewDeadline": time: invalid duration flops`),
	}, {
		name: "invalid retryPeriod",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"retryPeriod": "flops",
		}),
		err: errors.New(`failed to parse "retryPeriod": time: invalid duration flops`),
	}, {
		name: "invalid buckets - not an int",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"buckets": "not-an-int",
		}),
		err: errors.New(`failed to parse "buckets": strconv.ParseUint: parsing "not-an-int": invalid syntax`),
	}, {
		name: "invalid buckets - too small",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"buckets": "0",
		}),
		err: fmt.Errorf("buckets: value must be between 1 <= 0 <= %d", MaxBuckets),
	}, {
		name: "invalid buckets - too large",
		data: kmeta.UnionMaps(okData(), map[string]string{
			"buckets": strconv.Itoa(int(MaxBuckets + 1)),
		}),
		err: fmt.Errorf(`buckets: value must be between 1 <= %d <= %d`, MaxBuckets+1, MaxBuckets),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualConfig, actualErr := NewConfigFromConfigMap(
				&corev1.ConfigMap{
					Data: tc.data,
				})

			if tc.err != nil && tc.err.Error() != actualErr.Error() {
				t.Fatalf("Error = %v, want: %v", actualErr, tc.err)
			}

			if got, want := actualConfig, tc.expected; !cmp.Equal(got, want) {
				t.Errorf("Config = %#v, want: %#v, diff(-want,+got):\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}

func TestGetComponentConfig(t *testing.T) {
	const expectedName = "the-component"
	cases := []struct {
		name     string
		config   Config
		expected ComponentConfig
	}{{
		name: "component enabled",
		config: Config{
			ResourceLock:      "leases",
			LeaseDuration:     15 * time.Second,
			RenewDeadline:     10 * time.Second,
			RetryPeriod:       2 * time.Second,
			EnabledComponents: sets.NewString(expectedName),
		},
		expected: ComponentConfig{
			Component:     expectedName,
			LeaderElect:   true,
			ResourceLock:  "leases",
			LeaseDuration: 15 * time.Second,
			RenewDeadline: 10 * time.Second,
			RetryPeriod:   2 * time.Second,
		},
	}, {
		name: "component disabled",
		config: Config{
			ResourceLock:      "leases",
			LeaseDuration:     15 * time.Second,
			RenewDeadline:     10 * time.Second,
			RetryPeriod:       2 * time.Second,
			EnabledComponents: sets.NewString("not-the-component"),
		},
		expected: ComponentConfig{
			Component:   expectedName,
			LeaderElect: false,
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.config.GetComponentConfig(expectedName)
			if got, want := actual, tc.expected; !cmp.Equal(got, want) {
				t.Errorf("Incorrect config: diff(-want,+got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
