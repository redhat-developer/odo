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

	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"
)

const (
	appName = "app"
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
			kubeClient := tt.fields.kubeClient(ctrl)
			execClient := exec.NewExecClient(kubeClient)
			do := NewDeleteComponentClient(kubeClient, execClient)
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
			kubeClient := tt.fields.kubeClient(ctrl)
			execClient := exec.NewExecClient(kubeClient)
			do := NewDeleteComponentClient(kubeClient, execClient)
			if got := do.DeleteResources(tt.args.resources, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteComponentClient.DeleteResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteComponentClient_ListResourcesToDeleteFromDevfile(t *testing.T) {
	const compName = "nodejs-prj1-api-abhz"
	innerLoopCoreDeploymentName, _ := util.NamespaceKubernetesObject(compName, appName)

	// innerLoopCoreDeployment is the deployment created by odo dev for the component
	innerLoopCoreDeployment := odoTestingUtil.CreateFakeDeployment(compName, true)

	innerLoopCoreDeploymentUnstructured, e := kclient.ConvertK8sResourceToUnstructured(innerLoopCoreDeployment)
	if e != nil {
		t.Errorf("unable to convert deployment to unstructured")
	}

	// outerLoopResourceUnstructured is the deployment created by odo deploy
	outerLoopResourceUnstructured := unstructured.Unstructured{
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

	// labeledOuterloopResource is the deployment with labels set
	labeledOuterloopResource := *outerLoopResourceUnstructured.DeepCopy()
	labeledOuterloopResource.SetLabels(odolabels.GetLabels(compName, appName, "", odolabels.ComponentDeployMode, false))

	// innerLoopResourceUnstructured is the deployment that will be deployed by apply command with `odo dev`
	innerLoopResourceUnstructured := *outerLoopResourceUnstructured.DeepCopy()
	innerLoopResourceUnstructured.SetLabels(odolabels.GetLabels(compName, appName, "", odolabels.ComponentDevMode, false))

	deploymentRESTMapping := meta.RESTMapping{
		Resource: getGVR("apps", "v1", "Deployment"),
	}

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		devfileObj parser.DevfileObj
		appName    string
		mode       string
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
			name: "list innerloop core resource(deployment), and outerloop resources",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(innerLoopCoreDeployment, nil)

					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(&deploymentRESTMapping, nil)
					kubeClient.EXPECT().
						GetDynamicResource(deploymentRESTMapping.Resource, outerLoopResourceUnstructured.GetName()).
						Return(&labeledOuterloopResource, nil)

					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
				mode:       odolabels.ComponentAnyMode,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{innerLoopCoreDeploymentUnstructured, labeledOuterloopResource},
			wantErr:                 false,
		},
		{
			name: "list innerloop core resource(deployment) only",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(innerLoopCoreDeployment, nil)
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
				mode:    odolabels.ComponentDevMode,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{innerLoopCoreDeploymentUnstructured},
			wantErr:                 false,
		},
		{
			name: "list innerloop core resources(deployment), another innerloop resources",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(innerLoopCoreDeployment, nil)

					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(&deploymentRESTMapping, nil)
					kubeClient.EXPECT().
						GetDynamicResource(deploymentRESTMapping.Resource, outerLoopResourceUnstructured.GetName()).
						Return(&innerLoopResourceUnstructured, nil)

					return kubeClient
				},
			},
			args: args{
				devfileObj: func() parser.DevfileObj {
					obj := odoTestingUtil.GetTestDevfileObjFromFile("devfile-composite-apply-commands.yaml")
					// change the metadata name to the desired one since devfile.yaml has a different name
					metadata := obj.Data.GetMetadata()
					metadata.Name = compName
					obj.Data.SetMetadata(metadata)
					return obj
				}(),
				appName: appName,
				mode:    odolabels.ComponentDevMode,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{innerLoopCoreDeploymentUnstructured, innerLoopResourceUnstructured},
			wantErr:                 false,
		},
		{
			name: "list outerloop resources only",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).
						Return(&appsv1.Deployment{}, kerrors.NewNotFound(deploymentRESTMapping.Resource.GroupResource(), innerLoopCoreDeploymentName))
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(&deploymentRESTMapping, nil)
					kubeClient.EXPECT().
						GetDynamicResource(deploymentRESTMapping.Resource, outerLoopResourceUnstructured.GetName()).
						Return(&labeledOuterloopResource, nil)
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
				mode:       odolabels.ComponentAnyMode,
			},
			wantIsInnerLoopDeployed: false,
			wantResources:           []unstructured.Unstructured{labeledOuterloopResource},
			wantErr:                 false,
		},
		{
			name: "list uri-referenced outerloop resources",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).
						Return(&appsv1.Deployment{}, kerrors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "Deployments"}, innerLoopCoreDeploymentName))
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(&deploymentRESTMapping, nil)
					kubeClient.EXPECT().
						GetDynamicResource(deploymentRESTMapping.Resource, outerLoopResourceUnstructured.GetName()).
						Return(&labeledOuterloopResource, nil)
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy-with-k8s-uri.yaml"),
				appName:    appName,
				mode:       odolabels.ComponentAnyMode,
			},
			wantIsInnerLoopDeployed: false,
			wantResources:           []unstructured.Unstructured{labeledOuterloopResource},
			wantErr:                 false,
		},
		{
			name: "fetching inner loop resource failed due to some error(!NotFoundError)",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(&appsv1.Deployment{}, errors.New("some error"))
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
				mode:       odolabels.ComponentAnyMode,
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
					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(innerLoopCoreDeployment, nil)
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(nil, errors.New("some error"))
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
				mode:       odolabels.ComponentAnyMode,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{innerLoopCoreDeploymentUnstructured},
			wantErr:                 false,
		},
		{
			name: "failed to add outerloop resource to the list because kubeclient.GetDynamicResource() failed",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)
					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(innerLoopCoreDeployment, nil)
					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(&deploymentRESTMapping, nil)
					kubeClient.EXPECT().
						GetDynamicResource(deploymentRESTMapping.Resource, outerLoopResourceUnstructured.GetName()).
						Return(nil, errors.New("some error"))
					return kubeClient
				},
			},
			args: args{
				devfileObj: odoTestingUtil.GetTestDevfileObjFromFile("devfile-deploy.yaml"),
				appName:    appName,
				mode:       odolabels.ComponentAnyMode,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{innerLoopCoreDeploymentUnstructured},
			wantErr:                 false,
		},
		{
			name: "do not list outerloop resource if Dev mode is asked",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					kubeClient := kclient.NewMockClientInterface(ctrl)

					kubeClient.EXPECT().GetDeploymentByName(innerLoopCoreDeploymentName).Return(innerLoopCoreDeployment, nil)

					kubeClient.EXPECT().GetRestMappingFromUnstructured(outerLoopResourceUnstructured).Return(&deploymentRESTMapping, nil)
					kubeClient.EXPECT().
						GetDynamicResource(deploymentRESTMapping.Resource, outerLoopResourceUnstructured.GetName()).
						Return(&outerLoopResourceUnstructured, nil)

					return kubeClient
				},
			},
			args: args{
				devfileObj: func() parser.DevfileObj {
					obj := odoTestingUtil.GetTestDevfileObjFromFile("devfile-composite-apply-commands.yaml")
					// change the metadata name to the desired one since devfile.yaml has a different name
					metadata := obj.Data.GetMetadata()
					metadata.Name = compName
					obj.Data.SetMetadata(metadata)
					return obj
				}(),
				appName: appName,
				mode:    odolabels.ComponentDevMode,
			},
			wantIsInnerLoopDeployed: true,
			wantResources:           []unstructured.Unstructured{innerLoopCoreDeploymentUnstructured},
			wantErr:                 false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			do := DeleteComponentClient{
				kubeClient: tt.fields.kubeClient(ctrl),
			}
			gotIsInnerLoopDeployed, gotResources, err := do.ListResourcesToDeleteFromDevfile(tt.args.devfileObj, tt.args.appName, tt.args.devfileObj.GetMetadataName(), tt.args.mode)
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

					selector := odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode, false)
					client.EXPECT().GetRunningPodFromSelector(selector).Return(&corev1.Pod{}, &kclient.PodNotFoundError{Selector: selector})
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

					selector := odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode, false)
					client.EXPECT().GetRunningPodFromSelector(selector).Return(nil, errors.New("some un-ignorable error"))
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

					selector := odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode, false)
					client.EXPECT().GetRunningPodFromSelector(selector).Return(odoTestingUtil.CreateFakePod(componentName, "runtime"), nil)

					cmd := []string{"/bin/sh", "-c", "cd /projects/nodejs-starter && (echo \"Hello World!\") 1>>/proc/1/fd/1 2>>/proc/1/fd/2"}
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

					selector := odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode, false)
					pod := odoTestingUtil.CreateFakePod(componentName, "runtime")
					pod.Status.Phase = corev1.PodFailed
					client.EXPECT().GetRunningPodFromSelector(selector).Return(pod, nil)
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

					selector := odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode, false)
					fakePod := odoTestingUtil.CreateFakePod(componentName, "runtime")
					// Expecting this method to be called twice because if the command execution fails, we try to get the pod logs by calling GetOnePodFromSelector again.
					client.EXPECT().GetRunningPodFromSelector(selector).Return(fakePod, nil).Times(2)

					client.EXPECT().GetPodLogs(fakePod.Name, gomock.Any(), gomock.Any()).Return(nil, errors.New("an error"))

					cmd := []string{"/bin/sh", "-c", "cd /projects/nodejs-starter && (echo \"Hello World!\") 1>>/proc/1/fd/1 2>>/proc/1/fd/2"}
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
			kubeClient := tt.fields.kubeClient(ctrl)
			execClient := exec.NewExecClient(kubeClient)
			do := NewDeleteComponentClient(kubeClient, execClient)
			if err := do.ExecutePreStopEvents(tt.args.devfileObj, tt.args.appName, tt.args.devfileObj.GetMetadataName()); (err != nil) != tt.wantErr {
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
