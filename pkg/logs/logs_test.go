package logs

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/kclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func fakePod(name string) unstructured.Unstructured {
	return unstructured.Unstructured{map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": fmt.Sprintf("pod-%s", name),
			"uid":  fmt.Sprintf("pod-%s", name),
		},
		"spec": map[string]interface{}{
			"containers": map[string]interface{}{
				"name":  fmt.Sprintf("%s-1", name),
				"image": "image",
			},
		},
	}}
}

func fakeDeployment(name string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("deployment-%s", name),
				"uid":  fmt.Sprintf("deployment-%s", name),
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "test",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"spec": fakePod(name),
				},
			},
		},
	}
}

func generateOwnerRefernce(object unstructured.Unstructured) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: object.GetAPIVersion(),
		Kind:       object.GetKind(),
		Name:       object.GetName(),
		UID:        object.GetUID(),
	}
}

func TestLogsClient_matchOwnerReferenceWithResources_PodsWithOwnerInResources(t *testing.T) {
	type args struct {
		owner     metav1.OwnerReference
		resources func() []unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Case 1: pod owned by a deployment",
			args: args{
				resources: func() []unstructured.Unstructured {
					pod := fakePod("pod")
					deployment := fakeDeployment("deployment")
					deployOwnerRef := generateOwnerRefernce(deployment)
					pod.SetOwnerReferences([]metav1.OwnerReference{deployOwnerRef})
					return []unstructured.Unstructured{pod, deployment}
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubernetesClient := kclient.NewMockClientInterface(ctrl)
			o := &LogsClient{
				kubernetesClient: kubernetesClient,
			}

			got, err := o.matchOwnerReferenceWithResources(tt.args.resources()[0].GetOwnerReferences()[0], tt.args.resources())
			if (err != nil) != tt.wantErr {
				t.Errorf("matchOwnerReferenceWithResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("matchOwnerReferenceWithResources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogsClient_matchOwnerReferenceWithResources_PodsWithNoOwnerInResources(t *testing.T) {
	// pod and deployment that are not a part of args.resources
	independentDeploy := fakeDeployment("independent-deploy")
	independentPod := fakePod("independent-pod")
	independentPod.SetOwnerReferences([]metav1.OwnerReference{generateOwnerRefernce(independentDeploy)})

	type args struct {
		owner     metav1.OwnerReference
		resources func() []unstructured.Unstructured
	}
	tests := []struct {
		name     string
		args     args
		gvk      *meta.RESTMapping
		resource *unstructured.Unstructured
		want     bool
		wantErr  bool
	}{
		{
			name: "Case 1: Pod not owned by anything in `resources` slice",
			args: args{
				resources: func() []unstructured.Unstructured {
					pod := fakePod("pod")
					deployment := fakeDeployment("deployment")
					deployOwnerRef := generateOwnerRefernce(deployment)
					pod.SetOwnerReferences([]metav1.OwnerReference{deployOwnerRef})
					return []unstructured.Unstructured{pod, deployment}
				},
			},
			gvk: &meta.RESTMapping{
				Resource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			resource: &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"namespace": independentDeploy.GetNamespace(),
					"name":      independentDeploy.GetName(),
					"uid":       independentDeploy.GetUID(),
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "test",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
						"spec": fakePod(independentDeploy.GetName()),
					},
				},
			}},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubernetesClient := kclient.NewMockClientInterface(ctrl)
			kubernetesClient.EXPECT().GetRestMappingFromGVK(
				schema.FromAPIVersionAndKind(independentDeploy.GetAPIVersion(), independentDeploy.GetKind())).Return(tt.gvk, nil).AnyTimes()
			kubernetesClient.EXPECT().GetDynamicResource(tt.gvk.Resource, independentDeploy.GetName()).Return(tt.resource, nil).AnyTimes()

			o := &LogsClient{
				kubernetesClient: kubernetesClient,
			}
			got, err := o.matchOwnerReferenceWithResources(independentPod.GetOwnerReferences()[0], tt.args.resources())
			if (err != nil) != tt.wantErr {
				t.Errorf("matchOwnerReferenceWithResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("matchOwnerReferenceWithResources() got = %v, want %v", got, tt.want)
			}
		})
	}
}
