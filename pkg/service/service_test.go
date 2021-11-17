package service

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/kclient"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/golang/mock/gomock"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
)

func TestListDevfileServices(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := devfile.SetupTestFolder(testFolderName, fs)
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
		name             string
		devfileObj       parser.DevfileObj
		wantKeys         []string
		wantErr          error
		csvSupport       bool
		csvSupportErr    error
		gvrList          []meta.RESTMapping
		gvrListErr       error
		restMapping      *meta.RESTMapping
		restMappingErr   error
		u                unstructured.Unstructured
		inlinedComponent string
	}{
		{
			name: "No service in devfile",
			devfileObj: parser.DevfileObj{
				Data: devfile.GetDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			wantKeys:         []string{},
			wantErr:          nil,
			csvSupport:       true,
			csvSupportErr:    nil,
			gvrList:          []meta.RESTMapping{},
			gvrListErr:       nil,
			restMapping:      &meta.RESTMapping{},
			restMappingErr:   nil,
			u:                unstructured.Unstructured{},
			inlinedComponent: "",
		},
		{
			name: "Services including service bindings in devfile",
			devfileObj: parser.DevfileObj{
				Data: devfile.GetDevfileData(t, []devfile.InlinedComponent{
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
				}, nil),
			},
			wantKeys:       []string{"ServiceBinding/link1"},
			wantErr:        nil,
			csvSupport:     true,
			csvSupportErr:  nil,
			gvrList:        []meta.RESTMapping{},
			gvrListErr:     nil,
			restMapping:    &meta.RESTMapping{},
			restMappingErr: errors.New("some error"), // because SBO is not installed
			u:              unstructured.Unstructured{},
			inlinedComponent: `
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
			name: "URI reference in devfile",
			devfileObj: parser.DevfileObj{
				Data: devfile.GetDevfileData(t, nil, []devfile.URIComponent{
					{
						Name: "service1",
						URI:  filepath.Join(devfile.UriFolder, filepath.Base(testFileName.Name())),
					},
				}),
			},
			wantKeys:         []string{"Redis/service1"},
			wantErr:          nil,
			csvSupport:       false,
			csvSupportErr:    nil,
			gvrList:          nil,
			gvrListErr:       nil,
			restMapping:      nil,
			restMappingErr:   errors.New("some error"), // because Redis Operator is not installed
			u:                unstructured.Unstructured{},
			inlinedComponent: uriData,
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
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			fkClient := kclient.NewMockClientInterface(mockCtrl)
			fkClient.EXPECT().IsCSVSupported().Return(tt.csvSupport, tt.csvSupportErr).AnyTimes()
			fkClient.EXPECT().GetOperatorGVRList().Return(tt.gvrList, tt.gvrListErr).AnyTimes()
			_ = yaml.Unmarshal([]byte(tt.inlinedComponent), &tt.u)

			fkClient.EXPECT().GetRestMappingFromUnstructured(tt.u).Return(tt.restMapping, tt.restMappingErr).AnyTimes()
			//fkClient.EXPECT().GetRestMappingFromUnstructured(tt.u).Return(tt.restMapping, tt.restMappingErr).Times(2)

			got, gotErr := listDevfileServices(fkClient, tt.devfileObj, testFolderName, fs)
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
	testFileName, err := devfile.SetupTestFolder(testFolderName, fs)
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
				Data: devfile.GetDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "Services including service bindings in devfile",
			devfileObj: parser.DevfileObj{
				Data: devfile.GetDevfileData(t, []devfile.InlinedComponent{
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
				}, []devfile.URIComponent{
					{
						Name: "service1",
						URI:  filepath.Join(devfile.UriFolder, filepath.Base(testFileName.Name())),
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
	testFileName, err := devfile.SetupTestFolder(testFolderName, fs)
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
		Data: devfile.GetDevfileData(t, []devfile.InlinedComponent{
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
		}, []devfile.URIComponent{
			{
				Name: "service1",
				URI:  filepath.Join(devfile.UriFolder, filepath.Base(testFileName.Name())),
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

func TestGetAlmExample(t *testing.T) {
	tests := []struct {
		name     string
		examples []map[string]interface{}
		crd      string
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name: "rhoas-operator.0.9.0",
			crd:  "CloudServiceAccountRequest",
			examples: []map[string]interface{}{
				{
					"apiVersion": "rhoas.redhat.com/v1alpha1",
					"kind":       "ServiceRegistryConnection",
					"metadata": map[string]interface{}{
						"name":      "example",
						"namespace": "example-namespace",
						"labels": map[string]interface{}{
							"app.kubernetes.io/component":  "external-service",
							"app.kubernetes.io/managed-by": "rhoas",
						},
					},
					"spec": map[string]interface{}{
						"accessTokenSecretName": "rh-managed-services-api-accesstoken",
						"serviceRegistryId":     "exampleId",
						"credentials": map[string]interface{}{
							"serviceAccountSecretName": "service-account-secret",
						},
					},
				},
				{
					"apiVersion": "rhoas.redhat.com/v1alpha1",
					"kind":       "CloudServiceAccountRequest",
					"metadata": map[string]interface{}{
						"name":      "example",
						"namespace": "example-namespace",
					},
					"spec": map[string]interface{}{
						"serviceAccountName":        "rhoas-sa",
						"serviceAccountDescription": "Operator created service account",
						"serviceAccountSecretName":  "service-account-credentials",
						"accessTokenSecretName":     "rh-managed-services-api-accesstoken",
					},
				},
				{
					"apiVersion": "rhoas.redhat.com/v1alpha1",
					"kind":       "CloudServicesRequest",
					"metadata": map[string]interface{}{
						"name":      "example",
						"namespace": "example-namespace",
						"labels": map[string]interface{}{
							"app.kubernetes.io/component":  "external-service",
							"app.kubernetes.io/managed-by": "rhoas",
						},
					},
					"spec": map[string]interface{}{
						"accessTokenSecretName": "rh-cloud-services-api-accesstoken",
					},
				},
			},
			want: map[string]interface{}{
				"apiVersion": "rhoas.redhat.com/v1alpha1",
				"kind":       "CloudServiceAccountRequest",
				"metadata": map[string]interface{}{
					"name": "example",
				},
				"spec": map[string]interface{}{
					"serviceAccountName":        "rhoas-sa",
					"serviceAccountDescription": "Operator created service account",
					"serviceAccountSecretName":  "service-account-credentials",
					"accessTokenSecretName":     "rh-managed-services-api-accesstoken",
				},
			},
		},
		{
			name: "postgresoperator.v5.0.3",
			crd:  "PostgresCluster",
			examples: []map[string]interface{}{
				{
					"apiVersion": "postgres-operator.crunchydata.com/v1beta1",
					"kind":       "PostgresCluster",
					"metadata": map[string]interface{}{
						"name": "example",
					},
					"spec": map[string]interface{}{
						"instances": []map[string]interface{}{
							{
								"dataVolumeClaimSpec": map[string]interface{}{
									"accessModes": []string{
										"ReadWriteOnce",
									},
									"resources": map[string]interface{}{
										"requests": map[string]interface{}{
											"storage": "1Gi",
										},
									},
								},
								"replicas": 1,
							},
						},
						"postgresVersion": 13,
					},
				},
			},
			want: map[string]interface{}{
				"apiVersion": "postgres-operator.crunchydata.com/v1beta1",
				"kind":       "PostgresCluster",
				"metadata": map[string]interface{}{
					"name": "example",
				},
				"spec": map[string]interface{}{
					"instances": []map[string]interface{}{
						{
							"dataVolumeClaimSpec": map[string]interface{}{
								"accessModes": []string{
									"ReadWriteOnce",
								},
								"resources": map[string]interface{}{
									"requests": map[string]interface{}{
										"storage": "1Gi",
									},
								},
							},
							"replicas": 1,
						},
					},
					"postgresVersion": 13,
				},
			},
		},
		{
			name:     "crd not found",
			crd:      "unknown",
			examples: []map[string]interface{}{},
			want:     nil,
			wantErr:  true,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getAlmExample(tt.examples, tt.crd, "an operator")
			if err != nil != tt.wantErr {
				t.Errorf("Expected error %v but got %q\n", tt.wantErr, err)
			}
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("\nExpected: %+v\n     Got: %+v\n", tt.want, result)
			}
		})
	}
}
