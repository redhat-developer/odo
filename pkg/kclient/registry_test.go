package kclient

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/api"
	"k8s.io/client-go/dynamic"
)

func TestClient_GetRegistryList(t *testing.T) {
	type fields struct {
		Namespace     string
		DynamicClient func() dynamic.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []api.Registry
		wantErr bool
	}{
		//{
		//	name: "TODO",
		//	fields: fields{
		//		DynamicClient: func() dynamic.Interface {
		//			scheme := runtime.NewScheme()
		//			client := fake.NewSimpleDynamicClient(scheme)
		//			return client
		//		},
		//	},
		//},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Namespace:     tt.fields.Namespace,
				DynamicClient: tt.fields.DynamicClient(),
			}
			got, err := c.GetRegistryList()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetRegistryList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetRegistryList() = %v, want %v", got, tt.want)
			}
		})
	}
}
