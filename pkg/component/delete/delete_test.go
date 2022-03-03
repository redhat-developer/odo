package delete

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/kclient"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestDeleteComponentClient_ListResourcesToDelete(t *testing.T) {

	res1 := unstructured.Unstructured{}
	res1.SetAPIVersion("v1")
	res1.SetKind("deployment")
	res1.SetName("dep1")
	res2 := unstructured.Unstructured{}
	res2.SetAPIVersion("v1")
	res2.SetKind("service")
	res2.SetName("svc1")

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
					selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/part-of=app"
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
					selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/part-of=app"
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
					selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/part-of=app"
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
			got, err := do.ListResourcesToDelete(tt.args.componentName, tt.args.namespace)
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

	res1 := unstructured.Unstructured{}
	res1.SetAPIVersion("v1")
	res1.SetKind("deployment")
	res1.SetName("dep1")
	res2 := unstructured.Unstructured{}
	res2.SetAPIVersion("v1")
	res2.SetKind("service")
	res2.SetName("svc1")

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
					client.EXPECT().DeleteDynamicResource(res1.GetName(), "", "v1", res1.GetKind())
					client.EXPECT().DeleteDynamicResource(res2.GetName(), "", "v1", res2.GetKind())
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
					client.EXPECT().DeleteDynamicResource(res2.GetName(), "", "v1", res2.GetKind())
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
					client.EXPECT().DeleteDynamicResource(res1.GetName(), "", "v1", res1.GetKind()).Return(errors.New("some error"))
					client.EXPECT().DeleteDynamicResource(res2.GetName(), "", "v1", res2.GetKind())
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
			if got := do.DeleteResources(tt.args.resources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteComponentClient.DeleteResources() = %v, want %v", got, tt.want)
			}
		})
	}
}
