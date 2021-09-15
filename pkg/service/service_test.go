package service

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
)

type inlinedComponent struct {
	name    string
	inlined string
}

type uriComponent struct {
	name string
	uri  string
}

func getDevfileData(t *testing.T, inlined []inlinedComponent, uriComp []uriComponent) data.DevfileData {
	devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
	if err != nil {
		t.Error(err)
	}
	for _, component := range inlined {
		err = devfileData.AddComponents([]v1alpha2.Component{{
			Name: component.name,
			ComponentUnion: devfile.ComponentUnion{
				Kubernetes: &devfile.KubernetesComponent{
					K8sLikeComponent: devfile.K8sLikeComponent{
						BaseComponent: devfile.BaseComponent{},
						K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
							Inlined: component.inlined,
						},
					},
				},
			},
		},
		})
		if err != nil {
			t.Error(err)
		}
	}
	for _, component := range uriComp {
		err = devfileData.AddComponents([]v1alpha2.Component{{
			Name: component.name,
			ComponentUnion: devfile.ComponentUnion{
				Kubernetes: &devfile.KubernetesComponent{
					K8sLikeComponent: devfile.K8sLikeComponent{
						BaseComponent: devfile.BaseComponent{},
						K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
							Uri: component.uri,
						},
					},
				},
			},
		},
		})
		if err != nil {
			t.Error(err)
		}
	}
	return devfileData
}

func TestAddKubernetesComponentToDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	type args struct {
		crd        string
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Add service CRD to devfile.yaml",
			args: args{
				crd:  "test CRD",
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: getDevfileData(t, nil, nil),
					Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want: []v1alpha2.Component{{
				Name: "testName",
				ComponentUnion: devfile.ComponentUnion{
					Kubernetes: &devfile.KubernetesComponent{
						K8sLikeComponent: devfile.K8sLikeComponent{
							BaseComponent: devfile.BaseComponent{},
							K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
								Inlined: "test CRD",
							},
						},
					},
				},
			},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddKubernetesComponentToDevfile(tt.args.crd, tt.args.name, tt.args.devfileObj); (err != nil) != tt.wantErr {
				t.Errorf("AddKubernetesComponentToDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteKubernetesComponentFromDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := setup(testFolderName, fs)
	if err != nil {
		t.Errorf("unexpected error : %v", err)
		return
	}

	type args struct {
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Remove a CRD from devfile.yaml",
			args: args{
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: getDevfileData(t, []inlinedComponent{
						{
							name:    "testName",
							inlined: "test CRD",
						},
					}, nil),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
		{
			name: "Case 2: Remove a uri based component from devfile.yaml",
			args: args{
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]v1alpha2.Component{{
							Name: "testName",
							ComponentUnion: devfile.ComponentUnion{
								Kubernetes: &devfile.KubernetesComponent{
									K8sLikeComponent: devfile.K8sLikeComponent{
										BaseComponent: devfile.BaseComponent{},
										K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
											Uri: filepath.Join(uriFolder, filepath.Base(testFileName.Name())),
										},
									},
								},
							},
						},
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteKubernetesComponentFromDevfile(tt.args.name, tt.args.devfileObj, testFolderName, fs); (err != nil) != tt.wantErr {
				t.Errorf("DeleteKubernetesComponentFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListDevfileServices(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := setup(testFolderName, fs)
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
		wantKeys   []string
		wantErr    error
	}{
		{
			name: "No service in devfile",
			devfileObj: parser.DevfileObj{
				Data: getDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			wantKeys: []string{},
			wantErr:  nil,
		},
		{
			name: "Services including service bindings in devfile",
			devfileObj: parser.DevfileObj{
				Data: getDevfileData(t, []inlinedComponent{
					{
						name: "link1",
						inlined: `
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
				}, []uriComponent{
					{
						name: "service1",
						uri:  filepath.Join(uriFolder, filepath.Base(testFileName.Name())),
					},
				}),
			},
			wantKeys: []string{"Redis/service1", "ServiceBinding/link1"},
			wantErr:  nil,
		},
	}

	getKeys := func(m map[string]unstructured.Unstructured) []string {
		keys := make([]string, len(m))
		i := 0
		for key := range m {
			keys[i] = key
			i += 1
		}
		sort.Strings(keys)
		return keys
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := listDevfileServices(tt.devfileObj, testFolderName, fs)
			gotKeys := getKeys(got)
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotKeys, tt.wantKeys)
			}
			if gotErr != tt.wantErr {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotErr, tt.wantErr)
			}
		})
	}
}

func TestListDevfileLinks(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := setup(testFolderName, fs)
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
				Data: getDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "Services including service bindings in devfile",
			devfileObj: parser.DevfileObj{
				Data: getDevfileData(t, []inlinedComponent{
					{
						name: "link1",
						inlined: `
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
						name: "link2",
						inlined: `
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
				}, []uriComponent{
					{
						name: "service1",
						uri:  filepath.Join(uriFolder, filepath.Base(testFileName.Name())),
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

func TestFindDevfileServiceBinding(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := setup(testFolderName, fs)
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

	devfileObj := parser.DevfileObj{
		Data: getDevfileData(t, []inlinedComponent{
			{
				name: "link1",
				inlined: `
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
				name: "link2",
				inlined: `
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
		}, []uriComponent{
			{
				name: "service1",
				uri:  filepath.Join(uriFolder, filepath.Base(testFileName.Name())),
			},
		}),
		Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
	}

	type args struct {
		kind string
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantOK  bool
		wantErr error
	}{
		{
			name: "found",
			args: args{
				kind: "Redis",
				name: "redis",
			},
			want:    "nodejs-prj1-api-vtzg-redis-redis",
			wantOK:  true,
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				kind: "NotFound",
				name: "notfound",
			},
			want:    "",
			wantOK:  false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK, gotErr := findDevfileServiceBinding(devfileObj, tt.args.kind, tt.args.name, testFolderName, fs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got %v, expect %v", t.Name(), got, tt.want)
			}
			if !reflect.DeepEqual(gotOK, tt.wantOK) {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotOK, tt.wantOK)
			}
			if gotErr != tt.wantErr {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotErr, tt.wantErr)
			}
		})
	}
}

func setup(testFolderName string, fs devfileFileSystem.Filesystem) (devfileFileSystem.File, error) {
	err := fs.MkdirAll(testFolderName, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = fs.MkdirAll(filepath.Join(testFolderName, uriFolder), os.ModePerm)
	if err != nil {
		return nil, err
	}
	testFileName, err := fs.Create(filepath.Join(testFolderName, uriFolder, "example.yaml"))
	if err != nil {
		return nil, err
	}
	return testFileName, nil
}

func Test_addKubernetesComponent(t *testing.T) {

	type args struct {
		crd               string
		name              string
		componentContext  string
		devfileObj        parser.DevfileObj
		fs                devfileFileSystem.Filesystem
		uriFolderExists   bool
		fileAlreadyExists bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: the uri folder doesn't exist",
			args: args{
				crd:              "example",
				name:             "redis-service",
				componentContext: "/",
			},
		},
		{
			name: "case 2: the uri folder exist",
			args: args{
				crd:             "example",
				name:            "redis-service",
				uriFolderExists: true,
			},
		},
		{
			name: "case 3: the file already exists",
			args: args{
				crd:               "example",
				name:              "redis-service",
				uriFolderExists:   true,
				fileAlreadyExists: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := devfileFileSystem.NewFakeFs()
			tt.args.devfileObj = testingutil.GetTestDevfileObj(fs)
			tt.args.fs = fs

			if tt.args.uriFolderExists || tt.args.fileAlreadyExists {
				err := fs.MkdirAll(uriFolder, os.ModePerm)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				defer os.RemoveAll(uriFolder)
			}

			if tt.args.fileAlreadyExists {
				testFileName, err := fs.Create(filepath.Join(uriFolder, tt.args.name+".yaml"))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				defer os.RemoveAll(testFileName.Name())
			}

			if err := addKubernetesComponent(tt.args.crd, tt.args.name, tt.args.componentContext, tt.args.devfileObj, tt.args.fs); (err != nil) != tt.wantErr {
				t.Errorf("addKubernetesComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
