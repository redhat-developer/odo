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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	container "google.golang.org/api/container/v1beta1"
	boskoscommon "sigs.k8s.io/boskos/common"

	"knative.dev/pkg/test/gke"
	gkeFake "knative.dev/pkg/test/gke/fake"
	boskosFake "knative.dev/pkg/testutils/clustermanager/e2e-tests/boskos/fake"
	"knative.dev/pkg/testutils/clustermanager/e2e-tests/common"

	"github.com/google/go-cmp/cmp"
)

func setupFakeGKECluster() GKECluster {
	return GKECluster{
		Request:    &GKERequest{},
		operations: gkeFake.NewGKESDKClient(),
		boskosOps:  &boskosFake.FakeBoskosClient{},
	}
}

func TestGKECheckEnvironment(t *testing.T) {
	datas := []struct {
		kubectlOut         string
		kubectlErr         error
		gcloudOut          string
		gcloudErr          error
		clusterExist       bool
		requestClusterName string
		requestProject     string
		expProj            string
		expCluster         *string
		expErr             error
	}{
		{
			// Base condition, kubectl shouldn't return empty string if there is no error
			"", nil, "", nil, false, "", "", "", nil, nil,
		}, {
			// kubeconfig not set and gcloud not set
			"", fmt.Errorf("kubectl not set"), "", nil, false, "", "", "", nil, nil,
		}, {
			// kubeconfig failed
			"failed", fmt.Errorf("kubectl other err"), "", nil, false, "", "", "", nil, fmt.Errorf(`failed running kubectl config current-context: "failed"`),
		}, {
			// kubeconfig returned something other than "gke_PROJECT_REGION_CLUSTER"
			"gke_b_c", nil, "", nil, false, "", "", "", nil, nil,
		}, {
			// kubeconfig returned something other than "gke_PROJECT_REGION_CLUSTER"
			"gke_b_c_d_e", nil, "", nil, false, "", "", "", nil, nil,
		}, {
			// kubeconfig correctly set and cluster exist
			"gke_b_c_d", nil, "", nil, true, "d", "b", fakeProj, &fakeCluster, nil,
		}, {
			// kubeconfig correctly set and cluster exist, project wasn't requested
			"gke_b_c_d", nil, "", nil, true, "d", "", fakeProj, &fakeCluster, nil,
		}, {
			// kubeconfig correctly set and cluster exist, project doesn't match
			"gke_b_c_d", nil, "", nil, true, "d", "doesntexist", "", nil, nil,
		}, {
			// kubeconfig correctly set and cluster exist, cluster wasn't requested
			"gke_b_c_d", nil, "", nil, true, "", "b", fakeProj, &fakeCluster, nil,
		}, {
			// kubeconfig correctly set and cluster exist, cluster doesn't match
			"gke_b_c_d", nil, "", nil, true, "doesntexist", "b", "", nil, nil,
		}, {
			// kubeconfig correctly set and cluster exist, none of project/cluster requested
			"gke_b_c_d", nil, "", nil, true, "", "", fakeProj, &fakeCluster, nil,
		}, {
			// kubeconfig correctly set, but cluster doesn't exist
			"gke_b_c_d", nil, "", nil, false, "d", "", "", nil, fmt.Errorf("couldn't find cluster d in b in c, does it exist? cluster not found"),
		}, {
			// kubeconfig not set and gcloud failed
			"", fmt.Errorf("kubectl not set"), "", fmt.Errorf("gcloud failed"), false, "", "", "", nil, fmt.Errorf("failed getting gcloud project: gcloud failed"),
		}, {
			// kubeconfig not set and gcloud not set
			"", fmt.Errorf("kubectl not set"), "", nil, false, "", "", "", nil, nil,
		}, {
			// kubeconfig not set and gcloud set
			"", fmt.Errorf("kubectl not set"), "b", nil, false, "", "", fakeProj, nil, nil,
		},
	}

	oldFunc := common.StandardExec
	defer func() {
		common.StandardExec = oldFunc
	}()

	for _, data := range datas {
		fgc := setupFakeGKECluster()
		if data.clusterExist {
			parts := strings.Split(data.kubectlOut, "_")
			fgc.operations.CreateClusterAsync(parts[1], parts[2], "", &container.CreateClusterRequest{
				Cluster: &container.Cluster{
					Name: parts[3],
				},
				ProjectId: parts[1],
			})
		}
		fgc.Request.ClusterName = data.requestClusterName
		fgc.Request.Project = data.requestProject
		// mock for testing
		common.StandardExec = func(name string, args ...string) ([]byte, error) {
			var out []byte
			var err error
			switch name {
			case "gcloud":
				out = []byte(data.gcloudOut)
				err = data.gcloudErr
			case "kubectl":
				out = []byte(data.kubectlOut)
				err = data.kubectlErr
			}
			return out, err
		}

		err := fgc.checkEnvironment()
		var gotCluster *string
		if fgc.Cluster != nil {
			gotCluster = &fgc.Cluster.Name
		}

		errMsg := fmt.Sprintf("check environment with:\n\tkubectl output: %q\n\t\terror: %v\n\tgcloud output: %q\n\t\t"+
			"error: '%v'\n\t\tclustername requested: %q\n\t\tproject requested: %q",
			data.kubectlOut, data.kubectlErr, data.gcloudOut, data.gcloudErr, data.requestClusterName, data.requestProject)

		if (data.expErr != nil) != (err != nil) {
			t.Errorf("Error got = %v, want: %v", err, data.expErr)
		} else if err != nil && err.Error() != data.expErr.Error() {
			t.Errorf("Error got = %v, want: %v", err, data.expErr)
		}
		if dif := cmp.Diff(data.expCluster, gotCluster); dif != "" {
			t.Errorf("%s\nCluster got(+) is different from wanted(-)\n%v", errMsg, dif)
		}
		if dif := cmp.Diff(data.expProj, fgc.Project); dif != "" {
			t.Errorf("%s\nProject got(+) is different from wanted(-)\n%v", errMsg, dif)
		}
	}
}

func TestAcquire(t *testing.T) {
	predefinedClusterName := "predefined-cluster-name"
	fakeBoskosProj := "fake-boskos-proj-0"
	fakeBuildID := "1234"
	type wantResult struct {
		expCluster *container.Cluster
		expErr     error
		expPanic   bool
	}
	type request struct {
		clusterName string
		addons      []string
	}
	type testdata struct {
		request      request
		isProw       bool
		project      string
		existCluster *container.Cluster
		nextOpStatus []string
		boskosProjs  []string
		skipCreation bool
	}
	tests := []struct {
		name string
		td   testdata
		want wantResult
	}{
		{
			name: "cluster not exist, running in Prow and boskos not available",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, nextOpStatus: []string{}, boskosProjs: []string{}, skipCreation: false},
			want: wantResult{expCluster: nil, expErr: fmt.Errorf("failed acquiring boskos project: '%w'", errors.New("no GKE project available")), expPanic: false},
		}, {
			name: "cluster not exist, running in Prow and boskos available",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj},
				skipCreation: false},
			want: wantResult{expCluster: &container.Cluster{
				Name:         predefinedClusterName,
				Location:     "us-central1",
				Status:       "RUNNING",
				AddonsConfig: &container.AddonsConfig{},
				NodePools: []*container.NodePool{
					{
						Name:             "default-pool",
						InitialNodeCount: defaultGKEMinNodes,
						Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
						Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
					},
				},
				MasterAuth: &container.MasterAuth{
					Username: "admin",
				},
			}, expErr: nil, expPanic: false},
		}, {
			name: "cluster not exist, project not set, running in Prow and boskos not available",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, nextOpStatus: []string{}, boskosProjs: []string{}, skipCreation: false},
			want: wantResult{expCluster: nil, expErr: fmt.Errorf("failed acquiring boskos project: '%w'", errors.New("no GKE project available")), expPanic: false},
		},
		{
			name: "cluster not exist, project not set, running in Prow and boskos available",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         predefinedClusterName,
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				},
				nil,
				false,
			},
		},
		{
			name: "cluster name not defined in Acquire gets default cluster",
			td: testdata{
				request: request{clusterName: "", addons: []string{}},
				isProw:  true, nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         "kpkg-e2e-cls-1234",
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				},
				nil,
				false,
			},
		},
		{
			name: "project not set, not in Prow and boskos not available",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  false, nextOpStatus: []string{}, boskosProjs: []string{}, skipCreation: false},
			want: wantResult{nil, fmt.Errorf("GCP project must be set"), false},
		}, {
			name: "project not set, not in Prow and boskos available",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  false, nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{nil, fmt.Errorf("GCP project must be set"), false},
		}, {
			name: "cluster exists, project set, running in Prow",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj,
				existCluster: &container.Cluster{Name: "customcluster", Location: "us-central1"},
				nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         "customcluster",
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil, false,
			},
		}, {
			name: "cluster exists, project set and not running in Prow",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  false, project: fakeProj,
				existCluster: &container.Cluster{Name: "customcluster", Location: "us-central1"},
				nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         "customcluster",
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				},
				nil,
				false,
			},
		}, {
			name: "cluster exist, not running in Prow and skip creation",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  false, project: fakeProj,
				existCluster: &container.Cluster{Name: "customcluster", Location: "us-central1"},
				nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         "customcluster",
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil,
				false,
			},
		}, {
			name: "cluster exist, running in Prow and skip creation",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj,
				existCluster: &container.Cluster{Name: "customcluster", Location: "us-central1"},
				nextOpStatus: []string{}, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         "customcluster",
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil, false,
			},
		}, {
			name: "cluster not exist, not running in Prow and skip creation",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  false, project: fakeProj, nextOpStatus: []string{},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: true},
			want: wantResult{nil, fmt.Errorf("failing acquiring an existing cluster"), false},
		}, {
			name: "cluster not exist, running in Prow and skip creation",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: true},
			want: wantResult{nil, fmt.Errorf("failing acquiring an existing cluster"), false},
		}, {
			name: "skipped cluster creation as SkipCreation is requested",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         predefinedClusterName,
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil, false,
			},
		},
		{
			name: "cluster creation succeeded with addon",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{"istio"}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:     predefinedClusterName,
					Location: "us-central1",
					Status:   "RUNNING",
					AddonsConfig: &container.AddonsConfig{
						IstioConfig: &container.IstioConfig{Disabled: false},
					},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil, false,
			},
		},
		{
			name: "cluster creation succeeded without defined cluster name",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         predefinedClusterName,
					Location:     "us-central1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil, false,
			},
		}, {
			name: "cluster creation succeeded retry",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{"PENDING"},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{
				&container.Cluster{
					Name:         predefinedClusterName,
					Location:     "us-west1",
					Status:       "RUNNING",
					AddonsConfig: &container.AddonsConfig{},
					NodePools: []*container.NodePool{
						{
							Name:             "default-pool",
							InitialNodeCount: defaultGKEMinNodes,
							Config:           &container.NodeConfig{MachineType: "e2-standard-4", OauthScopes: []string{container.CloudPlatformScope}},
							Autoscaling:      &container.NodePoolAutoscaling{Enabled: true, MaxNodeCount: 3, MinNodeCount: 1},
						},
					},
					MasterAuth: &container.MasterAuth{
						Username: "admin",
					},
				}, nil, false,
			},
		}, {
			name: "cluster creation failed all retry",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{"PENDING", "PENDING", "PENDING"},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{nil, fmt.Errorf("timed out waiting"), false},
		}, {
			name: "cluster creation went bad state",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{}},
				isProw:  true, project: fakeProj, nextOpStatus: []string{"BAD", "BAD", "BAD"},
				boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{nil, fmt.Errorf("unexpected operation status: %q", "BAD"), false},
		}, {
			name: "bad addon, should get a panic",
			td: testdata{
				request: request{clusterName: predefinedClusterName, addons: []string{"bad_addon"}},
				isProw:  true, project: fakeProj, boskosProjs: []string{fakeBoskosProj}, skipCreation: false},
			want: wantResult{nil, nil, true},
		},
	}

	oldEnvFunc := common.GetOSEnv
	oldExecFunc := common.StandardExec
	// mock timeout so it doesn't run forever
	oldCreationTimeout := gkeFake.CreationTimeout
	// wait function polls every 500ms, give it 1000 to avoid random timeout
	gkeFake.CreationTimeout = 1000 * time.Millisecond
	defer func() {
		// restore
		common.GetOSEnv = oldEnvFunc
		common.StandardExec = oldExecFunc
		gkeFake.CreationTimeout = oldCreationTimeout
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.td
			defer func() {
				if r := recover(); r != nil && !tt.want.expPanic {
					t.Errorf("got unexpected panic: '%v'", r)
				}
			}()
			// mock for testing
			common.StandardExec = func(name string, args ...string) ([]byte, error) {
				var out []byte
				var err error
				switch name {
				case "gcloud":
					out = []byte("")
					err = nil
					if data.project != "" {
						out = []byte(data.project)
						err = nil
					}
				case "kubectl":
					out = []byte("")
					err = fmt.Errorf("kubectl not set")
					if data.existCluster != nil {
						context := fmt.Sprintf("gke_%s_%s_%s", data.project, data.existCluster.Location, data.existCluster.Name)
						out = []byte(context)
						err = nil
					}
				default:
					out, err = oldExecFunc(name, args...)
				}
				return out, err
			}
			common.GetOSEnv = func(key string) string {
				switch key {
				case "BUILD_NUMBER":
					return fakeBuildID
				case "PROW_JOB_ID": // needed to mock IsProw()
					if data.isProw {
						return "fake_job_id"
					}
					return ""
				}
				return oldEnvFunc(key)
			}
			fgc := setupFakeGKECluster()
			// Set up fake boskos
			for _, bos := range data.boskosProjs {
				fgc.boskosOps.(*boskosFake.FakeBoskosClient).NewGKEProject(bos)
			}
			fgc.Request = &GKERequest{
				Request: gke.Request{
					ClusterName: tt.td.request.clusterName,
					MinNodes:    defaultGKEMinNodes,
					MaxNodes:    defaultGKEMaxNodes,
					NodeType:    defaultGKENodeType,
					Region:      defaultGKERegion,
					Zone:        "",
					Addons:      tt.td.request.addons,
				},
				BackupRegions: defaultGKEBackupRegions,
				ResourceType:  defaultResourceType,
			}
			opCount := 0
			if data.existCluster != nil {
				opCount++
				fgc.Request.ClusterName = data.existCluster.Name
				rb, _ := gke.NewCreateClusterRequest(&fgc.Request.Request)
				fgc.operations.CreateClusterAsync(data.project, data.existCluster.Location, "", rb)
				fgc.Cluster, _ = fgc.operations.GetCluster(data.project, data.existCluster.Location, "", data.existCluster.Name)
			}

			fgc.Project = data.project
			for i, status := range data.nextOpStatus {
				fgc.operations.(*gkeFake.GKESDKClient).OpStatus[strconv.Itoa(opCount+i)] = status
			}

			if data.isProw && data.project == "" {
				fgc.isBoskos = true
			}
			if data.skipCreation {
				fgc.Request.SkipCreation = true
			}
			// Set asyncCleanup to false for easier testing, as it launches a
			// goroutine
			fgc.asyncCleanup = false
			err := fgc.Acquire()
			errMsg := fmt.Sprintf("testing acquiring cluster, with:\n\tisProw: '%v'\n\tproject: '%v'\n\texisting cluster: '%+v'\n\tSkip creation: '%+v'\n\t"+
				"next operations outcomes: '%v'\n\taddons: '%v'\n\tboskos projects: '%v'",
				data.isProw, data.project, data.existCluster, data.skipCreation, data.nextOpStatus, tt.td.request.addons, data.boskosProjs)
			if !reflect.DeepEqual(err, tt.want.expErr) {
				t.Errorf("%s\nerror got: '%v'\nerror want: '%v'", errMsg, err, tt.want.expErr)
			}
			if dif := cmp.Diff(tt.want.expCluster, fgc.Cluster); dif != "" {
				t.Errorf("%s\nCluster got(+) is different from wanted(-)\n%v", errMsg, dif)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type testdata struct {
		isProw      bool
		isBoskos    bool
		boskosState []*boskoscommon.Resource
		cluster     *container.Cluster
	}
	type wantResult struct {
		Boskos  []*boskoscommon.Resource
		Cluster *container.Cluster
		Err     error
	}
	tests := []struct {
		name string
		td   testdata
		want wantResult
	}{
		{
			name: "Not in prow",
			td: testdata{
				isProw:      false,
				boskosState: []*boskoscommon.Resource{},
				cluster: &container.Cluster{
					Name:     "customcluster",
					Location: "us-central1",
				},
			},
			want: wantResult{
				nil,
				nil,
				nil,
			},
		},
		{
			name: "Not in prow, but cluster doesn't exist",
			td: testdata{
				isProw:      false,
				boskosState: []*boskoscommon.Resource{},
				cluster:     nil,
			},
			want: wantResult{
				nil,
				nil,
				fmt.Errorf("cluster doesn't exist"),
			},
		},
		{
			name: "In prow",
			td: testdata{
				isProw:   true,
				isBoskos: true,
				boskosState: []*boskoscommon.Resource{{
					Name: fakeProj,
				}},
				cluster: &container.Cluster{
					Name:     "customcluster",
					Location: "us-central1",
				},
			},
			want: wantResult{
				[]*boskoscommon.Resource{{
					Type:  "gke-project",
					Name:  fakeProj,
					State: boskoscommon.Free,
				}},
				nil,
				nil,
			},
		},
	}

	oldEnvFunc := common.GetOSEnv
	oldExecFunc := common.StandardExec
	defer func() {
		// restore
		common.GetOSEnv = oldEnvFunc
		common.StandardExec = oldExecFunc
	}()

	// Mocked StandardExec so it does not actually run kubectl, gcloud commands.
	// Override so checkEnvironment returns nil all the time and each test use
	// the provided testdata.
	common.StandardExec = func(name string, args ...string) ([]byte, error) {
		var out []byte
		var err error
		switch name {
		case "gcloud":
			out = []byte("")
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
			data := tt.td
			common.GetOSEnv = func(key string) string {
				switch key {
				case "PROW_JOB_ID": // needed to mock IsProw()
					if data.isProw {
						return "fake_job_id"
					}
					return ""
				}
				return oldEnvFunc(key)
			}
			fgc := setupFakeGKECluster()
			fgc.Project = fakeProj
			fgc.isBoskos = data.isBoskos
			fgc.Request = &GKERequest{
				Request: gke.Request{
					MinNodes: defaultGKEMinNodes,
					MaxNodes: defaultGKEMaxNodes,
					NodeType: defaultGKENodeType,
					Region:   defaultGKERegion,
					Zone:     "",
				},
			}
			if data.cluster != nil {
				fgc.Request.ClusterName = data.cluster.Name
				rb, _ := gke.NewCreateClusterRequest(&fgc.Request.Request)
				fgc.operations.CreateClusterAsync(fakeProj, data.cluster.Location, "", rb)
				fgc.Cluster, _ = fgc.operations.GetCluster(fakeProj, data.cluster.Location, "", data.cluster.Name)
			}
			// Set up fake boskos
			for _, bos := range data.boskosState {
				fgc.boskosOps.(*boskosFake.FakeBoskosClient).NewGKEProject(bos.Name)
				// Acquire with default user
				fgc.boskosOps.(*boskosFake.FakeBoskosClient).AcquireGKEProject(defaultResourceType)
			}

			err := fgc.Delete()
			var gotCluster *container.Cluster
			if data.cluster != nil {
				gotCluster, _ = fgc.operations.GetCluster(fakeProj, data.cluster.Location, "", data.cluster.Name)
			}
			gotBoskos := fgc.boskosOps.(*boskosFake.FakeBoskosClient).GetResources()
			errMsg := fmt.Sprintf("testing deleting cluster, with:\n\tIs Prow: '%v'\n\tIs Boskos: '%v'\n\t"+
				"existing cluster: '%v'\n\tboskos state: '%v'",
				data.isProw, data.isBoskos, data.cluster, data.boskosState)
			if !reflect.DeepEqual(err, tt.want.Err) {
				t.Errorf("%s\nerror got: '%v'\nerror want: '%v'", errMsg, err, tt.want.Err)
			}
			if dif := cmp.Diff(tt.want.Cluster, gotCluster); dif != "" {
				t.Errorf("%s\nCluster got(+) is different from wanted(-)\n%v", errMsg, dif)
			}
			if dif := cmp.Diff(tt.want.Boskos, gotBoskos); dif != "" {
				t.Errorf("%s\nBoskos got(+) is different from wanted(-)\n%v", errMsg, dif)
			}
		})
	}
}
