package service

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/devfile/consts"
	devfiletesting "github.com/redhat-developer/odo/pkg/devfile/testing"

	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
)

func TestListDevfileLinks(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := devfiletesting.SetupTestFolder(testFolderName, fs)
	if err != nil {
		t.Errorf("unexpected error : %v", err)
		return
	}

	uriData := `
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
 name: redis
spec:
 kubernetesConfig:
   image: quay.io/opstree/redis:v6.2`

	err = fs.WriteFile(testFileName.Name(), []byte(uriData), os.ModePerm)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []struct {
		name       string
		devfileObj parser.DevfileObj
		want       []string
		wantErr    error
	}{
		{
			name: "No service in devfile",
			devfileObj: parser.DevfileObj{
				Data: devfiletesting.GetDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "Services including service bindings in devfile",
			devfileObj: parser.DevfileObj{
				Data: devfiletesting.GetDevfileData(t, []devfiletesting.InlinedComponent{
					{
						Name: "link1",
						Inlined: `
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
 name: nodejs-prj1-api-vtzg-redis-redis
spec:
 application:
   group: apps
   name: nodejs-prj1-api-vtzg-app
   resource: deployments
   version: v1
 bindAsFiles: false
 detectBindingResources: true
 services:
 - group: redis.redis.opstreelabs.in
   kind: Redis
   name: redis
   version: v1beta1`,
					},
					{
						Name: "link2",
						Inlined: `
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
 name: nodejs-prj1-api-vtzg-redis-redis
spec:
 application:
   group: apps
   name: nodejs-prj1-api-vtzg-app
   resource: deployments
   version: v1
 bindAsFiles: false
 detectBindingResources: true
 services:
 - group: redis.redis.opstreelabs.in
   kind: Service
   name: other
   version: v1beta1`,
					},
				}, []devfiletesting.URIComponent{
					{
						Name: "service1",
						URI:  filepath.Join(consts.UriFolder, filepath.Base(testFileName.Name())),
					},
				}),
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			want:    []string{"Redis/redis", "other"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := listDevfileLinks(tt.devfileObj, testFolderName, fs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got %v, expect %v", t.Name(), got, tt.want)
			}
			if gotErr != tt.wantErr {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotErr, tt.wantErr)
			}
		})
	}
}
