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

package pkg

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBenchmarkNames(t *testing.T) {
	testCases := []struct {
		benchmarkRoot      string
		expectedBenchmarks []string
		expectedError      bool
	}{{
		benchmarkRoot:      "testdir",
		expectedBenchmarks: []string{"test-benchmark1", "test-benchmark2", "test-benchmark3", "test-benchmark4"},
		expectedError:      false,
	}, {
		benchmarkRoot:      "non-existed-dir",
		expectedBenchmarks: []string{},
		expectedError:      true,
	}}

	for _, tc := range testCases {
		benchmarks, err := benchmarkNames(tc.benchmarkRoot)
		if diff := cmp.Diff(tc.expectedBenchmarks, benchmarks); diff != "" {
			t.Fatalf("benchmarkNames(%q) returns wrong result (-want +got):\n%s",
				tc.benchmarkRoot, diff)
		}

		if tc.expectedError && err == nil {
			t.Fatalf("expected to get error for getting benchmarks under %q, but got nil", tc.benchmarkRoot)
		}
		if !tc.expectedError && err != nil {
			t.Fatalf("expected to get no error for getting benchmarks under %q, but got %v", tc.benchmarkRoot, err)
		}
	}
}

func TestClusterConfigForBenchmark(t *testing.T) {
	testCases := []struct {
		benchmarkRoot         string
		benchmarkName         string
		expectedClusterConfig ClusterConfig
	}{{
		benchmarkRoot: "testdir",
		benchmarkName: "test-benchmark1",
		expectedClusterConfig: ClusterConfig{
			Location: "us-west1", NodeCount: 4, NodeType: "e2-standard-8", Addons: "istio"},
	}, {
		benchmarkRoot: "testdir",
		benchmarkName: "test-benchmark2",
		expectedClusterConfig: ClusterConfig{
			Location: defaultLocation, NodeCount: defaultNodeCount, NodeType: defaultNodeType, Addons: defaultAddons},
	}, {
		benchmarkRoot: "testdir",
		benchmarkName: "test-benchmark3",
		expectedClusterConfig: ClusterConfig{
			Location: defaultLocation, NodeCount: 1, NodeType: defaultNodeType, Addons: "istio"},
	}, {
		benchmarkRoot: "testdir",
		benchmarkName: "test-benchmark4",
		expectedClusterConfig: ClusterConfig{
			Location: defaultLocation, NodeCount: defaultNodeCount, NodeType: defaultNodeType, Addons: defaultAddons},
	}}

	for _, tc := range testCases {
		clusterConfig := clusterConfigForBenchmark(tc.benchmarkName, tc.benchmarkRoot)
		if diff := cmp.Diff(tc.expectedClusterConfig, clusterConfig); diff != "" {
			t.Fatalf("clusterConfigForBenchmark(%q, %q) returns wrong result (-want +got):\n%s",
				tc.benchmarkName, tc.benchmarkRoot, diff)
		}
	}
}

func TestClusterNameForBenchmark(t *testing.T) {
	repo, benchmarkName, expectedName := "serving", "load-test", "serving--load-test"
	if gotName := clusterNameForBenchmark(benchmarkName, repo); gotName != expectedName {
		t.Fatalf(
			"expected to get cluster name %q for benchmark %q under repo %q, but got %q",
			expectedName, benchmarkName, repo, gotName,
		)
	}
}

func TestBenchmarkNameForCluster(t *testing.T) {
	testCases := []struct {
		clusterName           string
		repo                  string
		expectedBenchmarkName string
	}{{
		clusterName:           "serving--load-test",
		repo:                  "serving",
		expectedBenchmarkName: "load-test",
	}, {
		clusterName:           "serving--load-test",
		repo:                  "eventing",
		expectedBenchmarkName: "",
	}, {
		clusterName:           "serving---load-test",
		repo:                  "serving",
		expectedBenchmarkName: "-load-test",
	}, {
		clusterName:           "eventing--broker-imc",
		repo:                  "eventing",
		expectedBenchmarkName: "broker-imc",
	}, {
		clusterName:           "eventing-contrib--broker-natss",
		repo:                  "eventing-contrib",
		expectedBenchmarkName: "broker-natss",
	}, {
		clusterName:           "eventing-contrib--broker-kafka",
		repo:                  "eventing",
		expectedBenchmarkName: "",
	}}

	for _, tc := range testCases {
		if benchmarkName := benchmarkNameForCluster(tc.clusterName, tc.repo); benchmarkName != tc.expectedBenchmarkName {
			t.Fatalf(
				"expected to get benchmark name %q for cluster %q in repo %q, but got %q",
				tc.expectedBenchmarkName, tc.clusterName, tc.repo, benchmarkName,
			)
		}
	}
}
