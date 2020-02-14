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

package clustermanager

import (
	"log"

	"knative.dev/pkg/test/gke"
)

// This is not a real test, it's for documenting purpose, showcasing the usage
// of entire clustermanager package
// Important: DO NOT add `// Output` comment inside this function as it will
// cause `go test` execute this function. See here: https://blog.golang.org/examples
func Example() {
	var (
		minNodes int64 = 1
		maxNodes int64 = 3
		nodeType       = "e2-standard-8"
		region         = "us-east1"
		zone           = "a"
		project        = "myGKEproject"
		addons         = []string{"istio"}
	)
	gkeClient := GKEClient{}
	clusterOps := gkeClient.Setup(GKERequest{
		Request: gke.Request{
			MinNodes: minNodes,
			MaxNodes: maxNodes,
			NodeType: nodeType,
			Region:   region,
			Zone:     zone,
			Project:  project,
			Addons:   addons,
		}})
	// Cast to GKEOperation
	gkeOps := clusterOps.(*GKECluster)
	if err := gkeOps.Acquire(); err != nil {
		log.Fatalf("failed acquire cluster: '%v'", err)
	}
	log.Printf("GKE project is: %q", gkeOps.Project)
	log.Printf("GKE cluster is: %v", gkeOps.Cluster)
}
