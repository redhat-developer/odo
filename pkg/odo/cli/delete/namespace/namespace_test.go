package namespace

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	_delete "github.com/redhat-developer/odo/pkg/project"
)

func TestCmdNamespaceDelete(t *testing.T) {
	type fields struct {
		commandName           string
		namespaceName         string
		forceFlag             bool
		deleteNamespaceClient func(ctrl *gomock.Controller) _delete.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "No namespace found",
			fields: fields{
				commandName:   "namespace",
				namespaceName: "my-namespace",
				forceFlag:     false,
				deleteNamespaceClient: func(ctrl *gomock.Controller) _delete.Client {
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().Exists("my-namespace").Return(false, nil)
					client.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(0)
					return client
				},
			},
			wantErr: true,
		},
		{
			name: "Delete namespace",
			fields: fields{
				commandName:   "namespace",
				namespaceName: "my-namespace",
				forceFlag:     true,
				deleteNamespaceClient: func(ctrl *gomock.Controller) _delete.Client {
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().Exists("my-namespace").Return(true, nil)
					client.EXPECT().Delete("my-namespace", false).Return(nil).Times(1)
					return client
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			do := &DeleteOptions{
				commandName:   tt.fields.commandName,
				namespaceName: tt.fields.namespaceName,
				forceFlag:     tt.fields.forceFlag,
				clientset: &clientset.Clientset{
					ProjectClient: tt.fields.deleteNamespaceClient(ctrl),
				},
			}
			if err := do.Run(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
