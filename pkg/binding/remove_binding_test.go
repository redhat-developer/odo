package binding

import (
	"testing"

	"github.com/devfile/library/v2/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/google/go-cmp/cmp"

	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
)

func TestBindingClient_ValidateRemoveBinding(t *testing.T) {
	type args struct {
		flags map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "--name flag is passed",
			args:    args{flags: map[string]string{"name": "redis-my-node-app"}},
			wantErr: false,
		},
		{
			name:    "--name flag is not passed",
			args:    args{flags: map[string]string{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &BindingClient{}
			if err := o.ValidateRemoveBinding(tt.args.flags); (err != nil) != tt.wantErr {
				t.Errorf("ValidateRemoveBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBindingClient_RemoveBinding(t *testing.T) {
	type args struct {
		servicebindingName string
		obj                parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		want    parser.DevfileObj
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "removed the binding successfully when bound as files",
			args: args{
				servicebindingName: "my-nodejs-app-cluster-sample", // name is hard coded from the devfile-with-service-binding-files.yaml
				obj:                odoTestingUtil.GetTestDevfileObjFromFile("devfile-with-service-binding-files.yaml"),
			},
			want: func() parser.DevfileObj {
				obj := odoTestingUtil.GetTestDevfileObjFromFile("devfile-with-service-binding-files.yaml")
				_ = obj.Data.DeleteComponent("my-nodejs-app-cluster-sample")
				return obj
			}(),
			wantErr: false,
		},
		{
			name: "removed the binding successfully when not bound as files",
			args: args{
				servicebindingName: "my-nodejs-app-cluster-sample",
				obj:                odoTestingUtil.GetTestDevfileObjFromFile("devfile-with-service-binding-envvars.yaml"),
			},
			want: func() parser.DevfileObj {
				obj := odoTestingUtil.GetTestDevfileObjFromFile("devfile-with-service-binding-envvars.yaml")
				_ = obj.Data.DeleteComponent("my-nodejs-app-cluster-sample")
				return obj
			}(),
			wantErr: false,
		},
		{
			name: "failed to remove non-existent binding",
			args: args{
				servicebindingName: "something",
				obj:                odoTestingUtil.GetTestDevfileObjFromFile("devfile-with-service-binding-files.yaml"),
			},
			want:    odoTestingUtil.GetTestDevfileObjFromFile("devfile-with-service-binding-files.yaml"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &BindingClient{}
			got, err := o.RemoveBinding(tt.args.servicebindingName, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveBinding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(devfileCtx.DevfileCtx{})); diff != "" {
				t.Errorf("BindingClient.RemoveBinding() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
