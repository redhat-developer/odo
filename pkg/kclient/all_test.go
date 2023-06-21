package kclient

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
)

func TestClient_GetAllResourcesFromSelector(t *testing.T) {
	type args struct {
		selector string
		ns       string
	}
	tests := []struct {
		name        string
		args        args
		objects     func() []runtime.Object
		checkResult func([]unstructured.Unstructured)
		wantErr     bool
	}{
		{
			name: "a deployment exists, matching labels",
			args: args{
				selector: "key1=value1",
			},
			objects: func() []runtime.Object {
				dep1 := appsv1.Deployment{}
				dep1.SetName("deploy1")
				dep1.SetLabels(map[string]string{
					"key1": "value1",
					"key2": "value2",
				})
				return []runtime.Object{&dep1}
			},
			checkResult: func(u []unstructured.Unstructured) {
				if len(u) != 1 {
					t.Fatalf("len of result should be %d but is %d", 1, len(u))
				}
				if u[0].GetName() != "deploy1" {
					t.Errorf("Name of 1st result should be %q but is %q", "deploy1", u[0].GetName())
				}
			},
		},
		{
			name: "a deployment exists, not matching labels",
			args: args{
				selector: "key1=value1",
			},
			objects: func() []runtime.Object {
				dep1 := appsv1.Deployment{}
				dep1.SetName("deploy1")
				dep1.SetLabels(map[string]string{
					"key1": "value2",
					"key2": "value1",
				})
				return []runtime.Object{&dep1}
			},
			checkResult: func(u []unstructured.Unstructured) {
				if len(u) != 0 {
					t.Fatalf("len of result should be %d but is %d", 0, len(u))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			objects := []runtime.Object{}
			if tt.objects != nil {
				objects = tt.objects()
			}
			fkclient.SetDynamicClient(scheme.Scheme, objects...)

			fkclientset.Kubernetes.Fake.Resources = []*metav1.APIResourceList{
				{
					GroupVersion: "apps/v1",
					APIResources: []metav1.APIResource{
						{
							Group:        "apps",
							Version:      "v1",
							Kind:         "Deployment",
							Name:         "deployments",
							SingularName: "deployment",
							Namespaced:   true,
							Verbs:        []string{"list"},
						},
					},
				},
			}

			got, err := fkclient.GetAllResourcesFromSelector(tt.args.selector, tt.args.ns)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetAllResourcesFromSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkResult != nil {
				tt.checkResult(got)
			}
		})
	}
}
