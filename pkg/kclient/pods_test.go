package kclient

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ktesting "k8s.io/client-go/testing"
)

func TestGetOnePodFromSelector(t *testing.T) {
	fakePod := FakePodStatus(corev1.PodRunning, "nodejs")
	fakePod.Labels["component"] = "nodejs"

	fakePodWithDeletionTimeStamp := FakePodStatus(corev1.PodRunning, "nodejs")
	fakePodWithDeletionTimeStamp.Labels["component"] = "nodejs"
	currentTime := metav1.NewTime(time.Now())
	fakePodWithDeletionTimeStamp.DeletionTimestamp = &currentTime

	type args struct {
		selector string
	}
	tests := []struct {
		name         string
		args         args
		returnedPods *corev1.PodList
		want         *corev1.Pod
		wantErr      bool
	}{
		{
			name: "valid number of pods",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{
					*fakePod,
				},
			},
			want:    fakePod,
			wantErr: false,
		},
		{
			name: "zero pods",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
		{
			name: "mutiple pods",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{
					*fakePod,
					*fakePod,
				},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
		{
			name: "pod is in the deletion state",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{
					*fakePodWithDeletionTimeStamp,
				},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()

			fkclientset.Kubernetes.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if action.(ktesting.ListAction).GetListRestrictions().Labels.String() != fmt.Sprintf("component=%s", "nodejs") {
					t.Errorf("list called with different selector want:%s, got:%s", fmt.Sprintf("component=%s", "nodejs"), action.(ktesting.ListAction).GetListRestrictions().Labels.String())
				}
				return true, tt.returnedPods, nil
			})

			got, err := fkclient.GetRunningPodFromSelector(tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOnePodFromSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr && err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOnePodFromSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPodUsingComponentName(t *testing.T) {
	fakePod := FakePodStatus(corev1.PodRunning, "nodejs")
	fakePod.Labels["component"] = "nodejs"

	type args struct {
		componentName string
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Pod
		wantErr bool
	}{
		{
			name: "list called with same component name",
			args: args{
				componentName: "nodejs",
			},
			want:    fakePod,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.Kubernetes.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if action.(ktesting.ListAction).GetListRestrictions().Labels.String() != fmt.Sprintf("component=%s", tt.args.componentName) {
					t.Errorf("list called with different selector want:%s, got:%s", fmt.Sprintf("component=%s", tt.args.componentName), action.(ktesting.ListAction).GetListRestrictions().Labels.String())
				}
				return true, &corev1.PodList{
					Items: []corev1.Pod{
						*fakePod,
					},
				}, nil
			})

			got, err := fkclient.GetPodUsingComponentName(tt.args.componentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodUsingComponentName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodUsingComponentName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateOwnerReference(object unstructured.Unstructured) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: object.GetAPIVersion(),
		Kind:       object.GetKind(),
		Name:       object.GetName(),
		UID:        object.GetUID(),
	}
}

func Test_matchOwnerReferenceWithResources_PodsWithOwnerInResources(t *testing.T) {
	type args struct {
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
					deployOwnerRef := generateOwnerReference(deployment)
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
			kubernetesClient := NewMockClientInterface(ctrl)

			got, err := matchOwnerReferenceWithResources(
				kubernetesClient,
				tt.args.resources()[0].GetOwnerReferences()[0],
				tt.args.resources(),
			)
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

func fakePod(name string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]interface{}{
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

func Test_matchOwnerReferenceWithResources_PodsWithNoOwnerInResources(t *testing.T) {
	// pod and deployment that are not a part of args.resources
	independentDeploy := fakeDeployment("independent-deploy")
	independentPod := fakePod("independent-pod")
	independentPod.SetOwnerReferences([]metav1.OwnerReference{generateOwnerReference(independentDeploy)})

	type args struct {
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
					deployOwnerRef := generateOwnerReference(deployment)
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
			kubernetesClient := NewMockClientInterface(ctrl)
			kubernetesClient.EXPECT().GetRestMappingFromGVK(
				schema.FromAPIVersionAndKind(independentDeploy.GetAPIVersion(), independentDeploy.GetKind())).Return(tt.gvk, nil).AnyTimes()
			kubernetesClient.EXPECT().GetDynamicResource(tt.gvk.Resource, independentDeploy.GetName()).Return(tt.resource, nil).AnyTimes()

			got, err := matchOwnerReferenceWithResources(kubernetesClient, independentPod.GetOwnerReferences()[0], tt.args.resources())
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
