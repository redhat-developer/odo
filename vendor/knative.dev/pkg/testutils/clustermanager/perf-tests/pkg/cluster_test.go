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
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	container "google.golang.org/api/container/v1beta1"

	"knative.dev/pkg/test/gke"
	gkeFake "knative.dev/pkg/test/gke/fake"
)

const (
	fakeProject       = "p"
	fakeRepository    = "r"
	testBenchmarkRoot = "testdir"
)

func setupFakeGKEClient() gkeClient {
	return gkeClient{
		ops: gkeFake.NewGKESDKClient(),
	}
}

func TestRecreateClusters(t *testing.T) {
	allExpectedClusters := map[string]ClusterConfig{
		clusterNameForBenchmark("test-benchmark1", fakeRepository): clusterConfigForBenchmark("test-benchmark1", testBenchmarkRoot),
		clusterNameForBenchmark("test-benchmark2", fakeRepository): clusterConfigForBenchmark("test-benchmark2", testBenchmarkRoot),
		clusterNameForBenchmark("test-benchmark3", fakeRepository): clusterConfigForBenchmark("test-benchmark3", testBenchmarkRoot),
		clusterNameForBenchmark("test-benchmark4", fakeRepository): clusterConfigForBenchmark("test-benchmark4", testBenchmarkRoot),
	}
	testCases := []struct {
		testName           string
		benchmarkRoot      string
		precreatedClusters map[string]ClusterConfig
		expectedClusters   map[string]ClusterConfig
	}{
		// all clusters will be created if there is no cluster at the beginning
		{
			testName:           "all clusters will be created if there is no cluster at the beginning",
			benchmarkRoot:      testBenchmarkRoot,
			precreatedClusters: make(map[string]ClusterConfig),
			expectedClusters:   allExpectedClusters,
		},
		// clusters that do not belong to this repo will not be touched
		{
			testName:      "clusters that do not belong to this repo will not be touched",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				"unrelated-cluster": {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: combineClusterMaps(map[string]ClusterConfig{
				"unrelated-cluster": {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			}, allExpectedClusters),
		},
		// clusters that belong to this repo, but have no corresponding benchmark, will be deleted
		{
			testName:      "clusters that belong to this repo, but have no corresponding benchmark, will be deleted",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("random-cluster", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: allExpectedClusters,
		},
		// clusters that belong to this repo, and have corresponding benchmark, will be recreated with the new config
		{
			testName:      "clusters that belong to this repo, and have corresponding benchmark, will be recreated with the new config",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("test-benchmark1", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 2,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: allExpectedClusters,
		},
		// multiple different clusters can be all handled in one single function call
		{
			testName:      "multiple different clusters can be all handled in one single function call",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("test-benchmark1", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 2,
					NodeType:  "n1-standard-4",
				},
				clusterNameForBenchmark("test-benchmark2", fakeRepository): {
					Location:  "us-west1",
					NodeCount: 2,
					NodeType:  "n1-standard-8",
				},
			},
			expectedClusters: allExpectedClusters,
		},
	}

	for _, tc := range testCases {
		client := setupFakeGKEClient()
		for name, config := range tc.precreatedClusters {
			region, zone := gke.RegionZoneFromLoc(config.Location)
			var addons []string
			if strings.TrimSpace(config.Addons) != "" {
				addons = strings.Split(config.Addons, ",")
			}
			req := &gke.Request{
				ClusterName: name,
				MinNodes:    config.NodeCount,
				MaxNodes:    config.NodeCount,
				NodeType:    config.NodeType,
				Addons:      addons,
			}
			creq, _ := gke.NewCreateClusterRequest(req)
			client.ops.CreateCluster(fakeProject, region, zone, creq)
		}
		err := client.RecreateClusters(fakeProject, fakeRepository, testBenchmarkRoot)
		fmt.Println(err)

		clusters, _ := client.ops.ListClustersInProject(fakeProject)
		actual := make(map[string]ClusterConfig)
		for _, cluster := range clusters {
			actual[cluster.Name] = ClusterConfig{
				Location:  cluster.Location,
				NodeCount: cluster.NodePools[0].Autoscaling.MaxNodeCount,
				NodeType:  cluster.NodePools[0].Config.MachineType,
				Addons:    getAddonsForCluster(cluster),
			}
		}

		if diff := cmp.Diff(tc.expectedClusters, actual); diff != "" {
			t.Fatalf("Test %q fails, RecreateClusters(%q, %q, %q) returns wrong result (-want +got):\n%s",
				tc.testName, fakeProject, fakeRepository, tc.benchmarkRoot, diff)
		}
	}
}

func TestReconcileClusters(t *testing.T) {
	reconciledClusters := map[string]ClusterConfig{
		clusterNameForBenchmark("test-benchmark1", fakeRepository): clusterConfigForBenchmark("test-benchmark1", testBenchmarkRoot),
		clusterNameForBenchmark("test-benchmark2", fakeRepository): clusterConfigForBenchmark("test-benchmark2", testBenchmarkRoot),
		clusterNameForBenchmark("test-benchmark3", fakeRepository): clusterConfigForBenchmark("test-benchmark3", testBenchmarkRoot),
		clusterNameForBenchmark("test-benchmark4", fakeRepository): clusterConfigForBenchmark("test-benchmark4", testBenchmarkRoot),
	}

	testCases := []struct {
		testName           string
		benchmarkRoot      string
		precreatedClusters map[string]ClusterConfig
		expectedClusters   map[string]ClusterConfig
	}{
		// all clusters will be created if there is no cluster at the beginning
		{
			testName:           "all clusters will be created if there is no cluster at the beginning",
			benchmarkRoot:      testBenchmarkRoot,
			precreatedClusters: make(map[string]ClusterConfig),
			expectedClusters:   reconciledClusters,
		},
		// clusters that do not belong to this repo will not be touched
		{
			testName:      "clusters that do not belong to this repo will not be touched",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				"unrelated-cluster": {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: combineClusterMaps(map[string]ClusterConfig{
				"unrelated-cluster": {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			}, reconciledClusters),
		},
		// clusters that belong to this repo, but have no corresponding benchmark, will be deleted
		{
			testName:      "clusters that belong to this repo, but have no corresponding benchmark, will be deleted",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("random-cluster", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: reconciledClusters,
		},
		// clusters that belong to this repo, and have corresponding benchmark, will be recreated with the new config
		{
			testName:      "clusters that belong to this repo, and have corresponding benchmark, will be recreated with the new config",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("test-benchmark1", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 2,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: reconciledClusters,
		},
		// multiple different clusters can be all handled in one single function call
		{
			testName:      "multiple different clusters can be all handled in one single function call",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("test-benchmark1", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 2,
					NodeType:  "n1-standard-4",
				},
				clusterNameForBenchmark("test-benchmark2", fakeRepository): {
					Location:  "us-west1",
					NodeCount: 2,
					NodeType:  "n1-standard-8",
				},
				clusterNameForBenchmark("random-cluster", fakeRepository): {
					Location:  "us-west1",
					NodeCount: 2,
					NodeType:  "n1-standard-8",
				},
			},
			expectedClusters: reconciledClusters,
		},
	}

	for _, tc := range testCases {
		client := setupFakeGKEClient()
		for name, config := range tc.precreatedClusters {
			region, zone := gke.RegionZoneFromLoc(config.Location)
			var addons []string
			if strings.TrimSpace(config.Addons) != "" {
				addons = strings.Split(config.Addons, ",")
			}
			req := &gke.Request{
				ClusterName: name,
				MinNodes:    config.NodeCount,
				MaxNodes:    config.NodeCount,
				NodeType:    config.NodeType,
				Addons:      addons,
			}
			creq, _ := gke.NewCreateClusterRequest(req)
			client.ops.CreateCluster(fakeProject, region, zone, creq)
		}
		err := client.ReconcileClusters(fakeProject, fakeRepository, testBenchmarkRoot)
		fmt.Println(err)

		clusters, _ := client.ops.ListClustersInProject(fakeProject)
		actual := make(map[string]ClusterConfig)
		for _, cluster := range clusters {
			actual[cluster.Name] = ClusterConfig{
				Location:  cluster.Location,
				NodeCount: cluster.NodePools[0].Autoscaling.MaxNodeCount,
				NodeType:  cluster.NodePools[0].Config.MachineType,
				Addons:    getAddonsForCluster(cluster),
			}
		}

		if diff := cmp.Diff(tc.expectedClusters, actual); diff != "" {
			t.Fatalf("Test %q fails, ReconcileClusters(%q, %q, %q) returns wrong result (-want +got):\n%s",
				tc.testName, fakeProject, fakeRepository, tc.benchmarkRoot, diff)
		}
	}
}

func TestDeleteClusters(t *testing.T) {
	testCases := []struct {
		testName           string
		benchmarkRoot      string
		precreatedClusters map[string]ClusterConfig
		expectedClusters   map[string]ClusterConfig
	}{
		// nothing will be done if there is no cluster at the beginning
		{
			testName:           "all related clusters will be deleted",
			benchmarkRoot:      testBenchmarkRoot,
			precreatedClusters: make(map[string]ClusterConfig),
			expectedClusters:   make(map[string]ClusterConfig),
		},
		// all clusters will be created if there is no cluster at the beginning
		{
			testName:      "all related clusters will be deleted",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("test-benchmark1", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 2,
					NodeType:  "n1-standard-4",
				},
				clusterNameForBenchmark("test-benchmark2", fakeRepository): {
					Location:  "us-west1",
					NodeCount: 2,
					NodeType:  "n1-standard-8",
				},
			},
			expectedClusters: make(map[string]ClusterConfig),
		},
		// clusters that do not belong to this repo will not be touched
		{
			testName:      "clusters that do not belong to this repo will not be touched",
			benchmarkRoot: testBenchmarkRoot,
			precreatedClusters: map[string]ClusterConfig{
				clusterNameForBenchmark("test-benchmark1", fakeRepository): {
					Location:  "us-central1",
					NodeCount: 2,
					NodeType:  "n1-standard-4",
				},
				"unrelated-cluster": {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			},
			expectedClusters: map[string]ClusterConfig{
				"unrelated-cluster": {
					Location:  "us-central1",
					NodeCount: 3,
					NodeType:  "n1-standard-4",
				},
			},
		},
	}

	for _, tc := range testCases {
		client := setupFakeGKEClient()
		for name, config := range tc.precreatedClusters {
			region, zone := gke.RegionZoneFromLoc(config.Location)
			var addons []string
			if strings.TrimSpace(config.Addons) != "" {
				addons = strings.Split(config.Addons, ",")
			}
			req := &gke.Request{
				ClusterName: name,
				MinNodes:    config.NodeCount,
				MaxNodes:    config.NodeCount,
				NodeType:    config.NodeType,
				Addons:      addons,
			}
			creq, _ := gke.NewCreateClusterRequest(req)
			client.ops.CreateCluster(fakeProject, region, zone, creq)
		}
		err := client.DeleteClusters(fakeProject, fakeRepository, testBenchmarkRoot)
		fmt.Println(err)

		clusters, _ := client.ops.ListClustersInProject(fakeProject)
		actual := make(map[string]ClusterConfig)
		for _, cluster := range clusters {
			actual[cluster.Name] = ClusterConfig{
				Location:  cluster.Location,
				NodeCount: cluster.NodePools[0].Autoscaling.MaxNodeCount,
				NodeType:  cluster.NodePools[0].Config.MachineType,
				Addons:    getAddonsForCluster(cluster),
			}
		}

		if diff := cmp.Diff(tc.expectedClusters, actual); diff != "" {
			t.Fatalf("Test %q fails, DeleteClusters(%q, %q, %q) returns wrong result (-want +got):\n%s",
				tc.testName, fakeProject, fakeRepository, tc.benchmarkRoot, diff)
		}
	}
}

// Return addons as a string slice for the given cluster.
// In this test we only use istio so only checking istio is enough here.
func getAddonsForCluster(cluster *container.Cluster) string {
	addons := make([]string, 0)
	if cluster.AddonsConfig.IstioConfig != nil && !cluster.AddonsConfig.IstioConfig.Disabled {
		addons = append(addons, "istio")
	}

	return strings.Join(addons, ",")
}

func combineClusterMaps(m1 map[string]ClusterConfig, m2 map[string]ClusterConfig) map[string]ClusterConfig {
	for name, config := range m2 {
		m1[name] = config
	}
	return m1
}
