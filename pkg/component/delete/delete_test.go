package delete

import (
	"errors"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/golang/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"
)

func TestDeleteComponentClient_ListClusterResourcesToDelete(t *testing.T) {
	res1 := getUnstructured("dep1", "deployment", "v1", "")
	res2 := getUnstructured("svc1", "service", "v1", "")

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		componentName string
		namespace     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "no resource found",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/managed-by=odo,app.kubernetes.io/part-of=app"
					client.EXPECT().GetAllResourcesFromSelector(selector, "my-ns").Return(nil, nil)
					return client
				},
			},
			args: args{
				componentName: "my-component",
				namespace:     "my-ns",
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "2 unrelated resources found",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res1, res2)
					client := kclient.NewMockClientInterface(ctrl)
					selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/managed-by=odo,app.kubernetes.io/part-of=app"
					client.EXPECT().GetAllResourcesFromSelector(selector, "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				componentName: "my-component",
				namespace:     "my-ns",
			},
			wantErr: false,
			want:    []unstructured.Unstructured{res1, res2},
		},
		{
			name: "2 resources found, one owned by the other",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					res1.SetOwnerReferences([]metav1.OwnerReference{
						{
							APIVersion: res2.GetAPIVersion(),
							Kind:       res2.GetKind(),
							Name:       res2.GetName(),
						},
					})
					resources = append(resources, res1, res2)
					client := kclient.NewMockClientInterface(ctrl)
					selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/managed-by=odo,app.kubernetes.io/part-of=app"
					client.EXPECT().GetAllResourcesFromSelector(selector, "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				componentName: "my-component",
				namespace:     "my-ns",
			},
			wantErr: false,
			want:    []unstructured.Unstructured{res2},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			do := &DeleteComponentClient{
				kubeClient: tt.fields.kubeClient(ctrl),
			}
			got, err := do.ListClusterResourcesToDelete(tt.args.componentName, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteComponentClient.ListResourcesToDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteComponentClient.ListResourcesToDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteComponentClient_DeleteResources(t *testing.T) {
	res1 := getUnstructured("dep1", "deployment", "v1", "")
	res2 := getUnstructured("svc1", "service", "v1", "")

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		resources []unstructured.Unstructured
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []unstructured.Unstructured
	}{
		{
			name: "2 resources deleted succesfully",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetRestMappingFromUnstructured(res1).Return(&meta.RESTMapping{
						Resource: schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: res1.GetKind(),
						},
					}, nil)
					client.EXPECT().GetRestMappingFromUnstructured(res2).Return(&meta.RESTMapping{
						Resource: schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: res2.GetKind(),
						},
					}, nil)
					client.EXPECT().DeleteDynamicResource(res1.GetName(), getGVR("", "v1", res1.GetKind()), false)
					client.EXPECT().DeleteDynamicResource(res2.GetName(), getGVR("", "v1", res2.GetKind()), false)
					return client
				},
			},
			args: args{
				resources: []unstructured.Unstructured{res1, res2},
			},
			want: nil,
		},
		{
			name: "2 resources, 1 deleted succesfully, 1 failed during restmapping",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetRestMappingFromUnstructured(res1).Return(nil, errors.New("some restmapping error"))
					client.EXPECT().GetRestMappingFromUnstructured(res2).Return(&meta.RESTMapping{
						Resource: schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: res2.GetKind(),
						},
					}, nil)
					client.EXPECT().DeleteDynamicResource(res2.GetName(), getGVR("", "v1", res2.GetKind()), false)
					return client
				},
			},
			args: args{
				resources: []unstructured.Unstructured{res1, res2},
			},
			want: []unstructured.Unstructured{res1},
		},
		{
			name: "2 resources, 1 deleted succesfully, 1 failed during deletion",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetRestMappingFromUnstructured(res1).Return(&meta.RESTMapping{
						Resource: schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: res1.GetKind(),
						},
					}, nil)
					client.EXPECT().GetRestMappingFromUnstructured(res2).Return(&meta.RESTMapping{
						Resource: schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: res2.GetKind(),
						},
					}, nil)
					client.EXPECT().DeleteDynamicResource(res1.GetName(), getGVR("", "v1", res1.GetKind()), false).Return(errors.New("some error"))
					client.EXPECT().DeleteDynamicResource(res2.GetName(), getGVR("", "v1", res2.GetKind()), false)
					return client
				},
			},
			args: args{
				resources: []unstructured.Unstructured{res1, res2},
			},
			want: []unstructured.Unstructured{res1},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			do := &DeleteComponentClient{
				kubeClient: tt.fields.kubeClient(ctrl),
			}
			if got := do.DeleteResources(tt.args.resources, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteComponentClient.DeleteResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteComponentClient_ListResourcesToDeleteFromDevfile(t *testing.T) {
	const appName = "app"
	const projectName = "default"
	const compName = "nodejs-prj1-api-abhz"
	innerLoopDeploymentName, _ := util.NamespaceKubernetesObject(compName, appName)
	deployment := odoTestingUtil.CreateFakeDeployment(compName)

	deployedInnerLoopResource, e := kclient.ConvertK8sResourceToUnstructured(deployment)
	if e != nil {
		t.Errorf("unable to convert deployment to unstructured")
	}

	outerLoopResource := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "my-component",
			},
			"spec": map[string]interface{}{
				"replicas": float64(1),
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "node-app",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "node-app",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": "quay.io/unknown-account/myimage",
								"name":  "main",
								"resources": map[string]interface{}{
									"limits": map[string]interface{}{
										"cpu":    "500m",
										"memory": "128Mi",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	outerLoopGVR := meta.RESTMapping{
		Resource: schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: outerLoopResource.GetKind(),
		},
	}
	deployedOuterLoopResource := getUnstructured("my-component", "Deployment", "apps/v1", projectName)

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		devfileObj parser.DevfileObj
		appName    string
	}
	tests := []struct {
		name                    string
		fields                  fields
		args                    args
		wantIsInnerLoopDeployed bool
		wantResources           []unstructured.Unstructured
		wantErr                 bool
	}{
		{
			name: "both innerloop and outerloop resources are pushed to the cluster",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopDeploymentName).Return(deployment, nil)

					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResource).Return(&outerLoopGVR, nil)
					kubeClient.EXPECT().
						GetDynamicResource(outerLoopGVR.Resource, outerLoopResource.GetName()).
						Return(&deployedOuterLoopResource, nil)

					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{deployedInnerLoopResource, deployedOuterLoopResource},
			wantErr:                 false,
		},
		{
			name: "no outerloop resources are present in the devfile",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopDeploymentName).Return(deployment, nil)
					return kubeClient
				},
			},
			args: args{
				devfileObj: func() parser.DevfileObj {
					obj := odoTestingUtil.GetTestDevfileObjFromFile("devfile.yaml")
					// change the metadata name to the desired one since devfile.yaml has a different name
					metadata := obj.Data.GetMetadata()
					metadata.Name = compName
					obj.Data.SetMetadata(metadata)
					return obj
				}(),
				appName: appName,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{deployedInnerLoopResource},
			wantErr:                 false,
		},
		{
			name: "only outerloop resources are deployed; innerloop resource is not found",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopDeploymentName).
						Return(&appsv1.Deployment{}, kerrors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "Deployments"}, innerLoopDeploymentName))
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResource).Return(&outerLoopGVR, nil)
					kubeClient.EXPECT().
						GetDynamicResource(outerLoopGVR.Resource, outerLoopResource.GetName()).
						Return(&deployedOuterLoopResource, nil)
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
			},
			wantIsInnerLoopDeployed: false,
			wantResources:           []unstructured.Unstructured{deployedOuterLoopResource},
			wantErr:                 false,
		},
		{
			name: "fetching inner loop resource failed due to some error(!NotFoundError)",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopDeploymentName).Return(&appsv1.Deployment{}, errors.New("some error"))
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
			},
			wantIsInnerLoopDeployed: false,
			wantResources:           nil,
			wantErr:                 true,
		},
		{
			name: "failed to add outerloop resource to the list because kubeclient.GetRestMappingFromUnstructured() failed",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopDeploymentName).Return(deployment, nil)
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResource).Return(nil, errors.New("some error"))
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{deployedInnerLoopResource},
			wantErr:                 false,
		},
		{
			name: "failed to add outerloop resource to the list because kubeclient.GetDynamicResource() failed",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopDeploymentName).Return(deployment, nil)
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResource).Return(&outerLoopGVR, nil)
					kubeClient.EXPECT().
						GetDynamicResource(outerLoopGVR.Resource, outerLoopResource.GetName()).
						Return(nil, errors.New("some error"))
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{deployedInnerLoopResource},
			wantErr:                 false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			do := DeleteComponentClient{
				kubeClient: tt.fields.kubeClient(ctrl),
			}
			gotIsInnerLoopDeployed, gotResources, err := do.ListResourcesToDeleteFromDevfile(tt.args.devfileObj, tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListResourcesToDeleteFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsInnerLoopDeployed != tt.wantIsInnerLoopDeployed {
				t.Errorf("ListResourcesToDeleteFromDevfile() gotIsInnerLoopDeployed = %v, want %v", gotIsInnerLoopDeployed, tt.wantIsInnerLoopDeployed)
			}
			if !reflect.DeepEqual(gotResources, tt.wantResources) {
				t.Errorf("ListResourcesToDeleteFromDevfile() gotResources = %v, want %v", gotResources, tt.wantResources)
			}
		})
	}
}

func TestDeleteComponentClient_ExecutePreStopEvents(t *testing.T) {
	const componentName = "nodejs-prj1-api-abhz"
	const appName = "app"
	fs := filesystem.NewFakeFs()

	devfileObjWithPreStopEvents := odoTestingUtil.GetTestDevfileObjWithPreStopEvents(fs, "runtime", "echo \"Hello World!\"")
	metadata := devfileObjWithPreStopEvents.Data.GetMetadata()
	metadata.Name = componentName
	devfileObjWithPreStopEvents.Data.SetMetadata(metadata)

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		devfileObj parser.DevfileObj
		appName    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "no preStop event to execute",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					return kclient.NewMockClientInterface(ctrl)
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
			},
			wantErr: false,
		},
		{
			name: "did not execute preStop event because pod was not found",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)

					selector := componentlabels.GetSelector(componentName, "app")
					client.EXPECT().GetOnePodFromSelector(selector).Return(&corev1.Pod{}, &kclient.PodNotFoundError{Selector: selector})
					return client
				},
			},
			args: args{
				devfileObj: devfileObjWithPreStopEvents,
				appName:    appName,
			},
			wantErr: false,
		},
		{
			name: "failed to execute preStop event because of an un-ignorable error",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)

					selector := componentlabels.GetSelector(componentName, "app")
					client.EXPECT().GetOnePodFromSelector(selector).Return(nil, errors.New("some un-ignorable error"))
					return client
				},
			},
			args: args{
				devfileObj: devfileObjWithPreStopEvents,
				appName:    appName,
			},
			wantErr: true,
		},
		{
			name: "successfully executed preStop events in the running pod",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)

					selector := componentlabels.GetSelector(componentName, "app")
					client.EXPECT().GetOnePodFromSelector(selector).Return(odoTestingUtil.CreateFakePod(componentName, "runtime"), nil)

					cmd := []string{"/bin/sh", "-c", "cd /projects/nodejs-starter && echo \"Hello World!\""}
					client.EXPECT().ExecCMDInContainer("runtime", "runtime", cmd, gomock.Any(), gomock.Any(), nil, false).Return(nil)

					return client
				},
			},
			args: args{
				devfileObj: devfileObjWithPreStopEvents,
				appName:    appName,
			},
			wantErr: false,
		},
		{
			name: "did not execute PreStopEvents because the pod is not in the running state",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)

					selector := componentlabels.GetSelector(componentName, "app")
					pod := odoTestingUtil.CreateFakePod(componentName, "runtime")
					pod.Status.Phase = corev1.PodFailed
					client.EXPECT().GetOnePodFromSelector(selector).Return(pod, nil)
					return client
				},
			},
			args: args{
				devfileObj: devfileObjWithPreStopEvents,
				appName:    appName,
			},
			wantErr: false,
		},
		{
			name: "failed to execute PreStopEvents because it failed to execute the command inside the container, but no error returned",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)

					selector := componentlabels.GetSelector(componentName, "app")
					client.EXPECT().GetOnePodFromSelector(selector).Return(odoTestingUtil.CreateFakePod(componentName, "runtime"), nil)

					cmd := []string{"/bin/sh", "-c", "cd /projects/nodejs-starter && echo \"Hello World!\""}
					client.EXPECT().ExecCMDInContainer("runtime", "runtime", cmd, gomock.Any(), gomock.Any(), nil, false).Return(errors.New("some error"))

					return client
				},
			},
			args: args{
				devfileObj: devfileObjWithPreStopEvents,
				appName:    appName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			do := &DeleteComponentClient{
				kubeClient: tt.fields.kubeClient(ctrl),
			}
			if err := do.ExecutePreStopEvents(tt.args.devfileObj, tt.args.appName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// getUnstructured returns an unstructured.Unstructured object
func getUnstructured(name, kind, apiVersion, namespace string) (u unstructured.Unstructured) {
	u.SetName(name)
	u.SetKind(kind)
	u.SetAPIVersion(apiVersion)
	u.SetNamespace(namespace)
	return
}

func getGVR(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}
