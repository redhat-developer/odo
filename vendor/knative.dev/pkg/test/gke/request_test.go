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

package gke

import "testing"

func TestNewCreateClusterRequest(t *testing.T) {
	datas := []struct {
		req           *Request
		errorExpected bool
	}{
		{
			req: &Request{
				Project:     "project-a",
				ClusterName: "name-a",
				MinNodes:    1,
				MaxNodes:    1,
				NodeType:    "n1-standard-4",
				Addons:      []string{"Istio"},
			},
			errorExpected: false,
		}, {
			req: &Request{
				Project:     "project-b",
				ClusterName: "name-b",
				MinNodes:    10,
				MaxNodes:    10,
				NodeType:    "n1-standard-8",
				Addons:      []string{"HorizontalPodAutoscaling", "HttpLoadBalancing", "CloudRun"},
			},
			errorExpected: false,
		},
		{
			req: &Request{
				Project:        "project-b",
				ClusterName:    "name-b",
				MinNodes:       10,
				MaxNodes:       10,
				NodeType:       "n1-standard-8",
				Addons:         []string{"HorizontalPodAutoscaling", "HttpLoadBalancing", "CloudRun"},
				ReleaseChannel: "rapid",
			},
			errorExpected: false,
		},
		{
			req: &Request{
				Project:        "project-b",
				ClusterName:    "name-b",
				GKEVersion:     "1-2-3",
				MinNodes:       10,
				MaxNodes:       10,
				NodeType:       "n1-standard-8",
				Addons:         []string{"HorizontalPodAutoscaling", "HttpLoadBalancing", "CloudRun"},
				ReleaseChannel: "rapid",
			},
			errorExpected: true,
		},
		{
			req: &Request{
				Project:    "project-c",
				GKEVersion: "1-2-3",
				MinNodes:   1,
				MaxNodes:   1,
				NodeType:   "n1-standard-4",
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:     "project-d",
				GKEVersion:  "1-2-3",
				ClusterName: "name-d",
				MinNodes:    0,
				MaxNodes:    1,
				NodeType:    "n1-standard-4",
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:     "project-e",
				GKEVersion:  "1-2-3",
				ClusterName: "name-e",
				MinNodes:    10,
				MaxNodes:    1,
				NodeType:    "n1-standard-4",
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:     "project-f",
				GKEVersion:  "1-2-3",
				ClusterName: "name-f",
				MinNodes:    1,
				MaxNodes:    1,
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:     "project-d",
				GKEVersion:  "1-2-3",
				ClusterName: "name-d",
				MinNodes:    0,
				MaxNodes:    1,
				NodeType:    "n1-standard-4",
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:     "project-e",
				GKEVersion:  "1-2-3",
				ClusterName: "name-e",
				MinNodes:    10,
				MaxNodes:    1,
				NodeType:    "n1-standard-4",
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:     "project-f",
				GKEVersion:  "1-2-3",
				ClusterName: "name-f",
				MinNodes:    1,
				MaxNodes:    1,
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:                "project-g",
				GKEVersion:             "1-2-3",
				ClusterName:            "name-g",
				MinNodes:               1,
				MaxNodes:               1,
				NodeType:               "n1-standard-4",
				EnableWorkloadIdentity: true,
			},
			errorExpected: false,
		}, {
			req: &Request{
				GKEVersion:             "1-2-3",
				ClusterName:            "name-h",
				MinNodes:               3,
				MaxNodes:               3,
				NodeType:               "n1-standard-4",
				EnableWorkloadIdentity: true,
			},
			errorExpected: true,
		}, {
			req: &Request{
				Project:        "project-i",
				GKEVersion:     "1-2-3",
				ClusterName:    "name-i",
				MinNodes:       3,
				MaxNodes:       3,
				NodeType:       "n1-standard-4",
				ServiceAccount: "sa-i",
			},
			errorExpected: false,
		}}
	for _, data := range datas {
		createReq, err := NewCreateClusterRequest(data.req)
		if data.errorExpected {
			if err == nil {
				t.Errorf("Expected error from request '%v', but got nil", data.req)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error from request '%v', but got '%v'", data.req, err)
			}
			if createReq == nil {
				t.Error("Expected a valid request, but got nil")
			}
		}
	}
}
