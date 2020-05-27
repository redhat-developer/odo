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

package clustermanager

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"knative.dev/pkg/test/gke"
	"knative.dev/pkg/testutils/clustermanager/e2e-tests/common"
)

var (
	fakeProj    = "b"
	fakeCluster = "d"
)

func TestSetup(t *testing.T) {
	minNodesOverride := int64(2)
	maxNodesOverride := int64(4)
	nodeTypeOverride := "foonode"
	regionOverride := "fooregion"
	zoneOverride := "foozone"
	boskosResTypeOverride := "customResType"
	fakeAddons := "fake-addon"
	fakeBuildID := "1234"
	type env struct {
		isProw          bool
		regionEnv       string
		backupRegionEnv string
	}
	tests := []struct {
		name string
		arg  GKERequest
		env  env
		want *GKECluster
	}{
		{
			name: "Defaults, not running in Prow",
			arg:  GKERequest{},
			env:  env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Defaults, running in Prow",
			arg:  GKERequest{},
			env:  env{true, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Custom Boskos Resource type, running in Prow",
			arg:  GKERequest{ResourceType: boskosResTypeOverride},
			env:  env{true, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  boskosResTypeOverride,
				},
			},
		}, {
			name: "Project provided, not running in Prow",
			arg: GKERequest{
				Request: gke.Request{
					Project:    fakeProj,
					GKEVersion: "1.2.3",
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						Project:     "b",
						GKEVersion:  "1.2.3",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
				Project:      fakeProj,
				asyncCleanup: true,
			},
		}, {
			name: "Project provided, running in Prow",
			arg: GKERequest{
				Request: gke.Request{
					Project: fakeProj,
				},
			},
			env: env{true, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						Project:     "b",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
				Project:      fakeProj,
				asyncCleanup: true,
			},
		}, {
			name: "Cluster name provided, not running in Prow",
			arg: GKERequest{
				Request: gke.Request{
					ClusterName: "predefined-cluster-name",
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "predefined-cluster-name",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Cluster name provided, running in Prow",
			arg: GKERequest{
				Request: gke.Request{
					ClusterName: "predefined-cluster-name",
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "predefined-cluster-name",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Override other parts",
			arg: GKERequest{
				Request: gke.Request{
					MinNodes: minNodesOverride,
					MaxNodes: maxNodesOverride,
					NodeType: nodeTypeOverride,
					Region:   regionOverride,
					Zone:     zoneOverride,
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    2,
						MaxNodes:    4,
						NodeType:    "foonode",
						Region:      "fooregion",
						Zone:        "foozone",
						Addons:      nil,
					},
					BackupRegions: []string{},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Override other parts but not zone",
			arg: GKERequest{
				Request: gke.Request{
					MinNodes: minNodesOverride,
					MaxNodes: maxNodesOverride,
					NodeType: nodeTypeOverride,
					Region:   regionOverride,
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    2,
						MaxNodes:    4,
						NodeType:    "foonode",
						Region:      "fooregion",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: nil,
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Min Nodes > Max Nodes",
			arg: GKERequest{
				Request: gke.Request{
					MinNodes: 10,
					NodeType: nodeTypeOverride,
					Region:   regionOverride,
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    10,
						MaxNodes:    10,
						NodeType:    "foonode",
						Region:      "fooregion",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: nil,
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Set env Region",
			arg:  GKERequest{},
			env:  env{false, "customregion", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "customregion",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Set env backupzone",
			arg:  GKERequest{},
			env:  env{false, "", "backupregion1 backupregion2"},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      nil,
					},
					BackupRegions: []string{"backupregion1", "backupregion2"},
					ResourceType:  defaultResourceType,
				},
			},
		}, {
			name: "Set addons",
			arg: GKERequest{
				Request: gke.Request{
					Addons: []string{fakeAddons},
				},
			},
			env: env{false, "", ""},
			want: &GKECluster{
				Request: &GKERequest{
					Request: gke.Request{
						ClusterName: "",
						MinNodes:    1,
						MaxNodes:    3,
						NodeType:    "e2-standard-4",
						Region:      "us-central1",
						Zone:        "",
						Addons:      []string{fakeAddons},
					},
					BackupRegions: []string{"us-west1", "us-east1"},
					ResourceType:  defaultResourceType,
				},
			},
		},
	}

	// mock GetOSEnv for testing
	oldEnvFunc := common.GetOSEnv
	oldExecFunc := common.StandardExec
	oldDefaultCred := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	tf, _ := ioutil.TempFile("", "foo")
	tf.WriteString(`{"type": "service_account"}`)
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tf.Name())
	defer func() {
		// restore
		common.GetOSEnv = oldEnvFunc
		common.StandardExec = oldExecFunc
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", oldDefaultCred)
		os.Remove(tf.Name())
	}()
	// mock as kubectl not set and gcloud set as "b", so check environment
	// return project as "b"
	common.StandardExec = func(name string, args ...string) ([]byte, error) {
		var out []byte
		var err error
		switch name {
		case "gcloud":
			out = []byte("b")
			err = nil
		case "kubectl":
			out = []byte("")
			err = fmt.Errorf("kubectl not set")
		default:
			out, err = oldExecFunc(name, args...)
		}
		return out, err
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.GetOSEnv = func(s string) string {
				switch s {
				case "E2E_CLUSTER_REGION":
					return tt.env.regionEnv
				case "E2E_CLUSTER_BACKUP_REGIONS":
					return tt.env.backupRegionEnv
				case "BUILD_NUMBER":
					return fakeBuildID
				case "PROW_JOB_ID": // needed to mock IsProw()
					if tt.env.isProw {
						return "fake_job_id"
					}
					return ""
				}
				return oldEnvFunc(s)
			}
			c := GKEClient{}
			co := c.Setup(tt.arg)
			errMsg := fmt.Sprintf("testing setup with:\n\t%+v\n\tregionEnv: %v\n\tbackupRegionEnv: %v",
				tt.arg, tt.env.regionEnv, tt.env.backupRegionEnv)
			gotCo := co.(*GKECluster)
			// mock for easier comparison
			gotCo.boskosOps = nil
			if dif := cmp.Diff(gotCo.Request, tt.want.Request); dif != "" {
				t.Errorf("%s\nRequest got(+) is different from wanted(-)\n%v", errMsg, dif)
			}
		})
	}
}
