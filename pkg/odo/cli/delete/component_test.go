package delete

import (
	"testing"

	"github.com/golang/mock/gomock"
	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestComponentOptions_deleteNamedComponent(t *testing.T) {
	type fields struct {
		name                  string
		namespace             string
		forceFlag             bool
		kubernetesClient      func(ctrl *gomock.Controller) kclient.ClientInterface
		deleteComponentClient func(ctrl *gomock.Controller) _delete.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "No resource found",
			fields: fields{
				name:      "my-component",
				namespace: "",
				forceFlag: false,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetCurrentNamespace().Return("my-namespace")
					return client
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListResourcesToDelete("my-component", "my-namespace").Return(nil, nil)
					client.EXPECT().DeleteResources(gomock.Any()).Times(0)
					return client
				},
			},
			wantErr: false,
		},
		{
			name: "2 resources to delete",
			fields: fields{
				name:      "my-component",
				namespace: "",
				forceFlag: true,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetCurrentNamespace().Return("my-namespace")
					return client
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					var resources []unstructured.Unstructured
					res1 := unstructured.Unstructured{}
					res1.SetAPIVersion("v1")
					res1.SetKind("deployment")
					res1.SetName("dep1")
					res2 := unstructured.Unstructured{}
					res2.SetAPIVersion("v1")
					res2.SetKind("service")
					res2.SetName("svc1")
					resources = append(resources, res1, res2)
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListResourcesToDelete("my-component", "my-namespace").Return(resources, nil)
					client.EXPECT().DeleteResources([]unstructured.Unstructured{res1, res2}).Times(1)
					return client
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &ComponentOptions{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				forceFlag: tt.fields.forceFlag,
				clientset: &clientset.Clientset{
					KubernetesClient: tt.fields.kubernetesClient(ctrl),
					DeleteClient:     tt.fields.deleteComponentClient(ctrl),
				},
			}
			if err := o.deleteNamedComponent(); (err != nil) != tt.wantErr {
				t.Errorf("ComponentOptions.deleteNamedComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
