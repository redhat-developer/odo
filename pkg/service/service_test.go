package service

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/redhat-developer/odo/pkg/devfile/consts"
	devfiletesting "github.com/redhat-developer/odo/pkg/devfile/testing"
	"github.com/redhat-developer/odo/pkg/kclient"

	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestListDevfileServices(t *testing.T) {
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
				Data: devfiletesting.GetDevfileData(t, nil, nil),
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
				Data: devfiletesting.GetDevfileData(t, nil, []devfiletesting.URIComponent{
					{
						Name: "service1",
						URI:  filepath.Join(consts.UriFolder, filepath.Base(testFileName.Name())),
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

func TestListSucceededClusterServiceVersions(t *testing.T) {

	tests := []struct {
		name              string
		isCSVsupported    bool
		isCSVsupportedErr error
		list              *olm.ClusterServiceVersionList
		listErr           error
		expectedList      *olm.ClusterServiceVersionList
		expectedErr       bool
	}{
		{
			name:              "error getting supported csv",
			isCSVsupported:    false,
			isCSVsupportedErr: errors.New("an error"),
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       true,
		},
		{
			name:              "non supported csv",
			isCSVsupported:    false,
			isCSVsupportedErr: nil,
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       false,
		},
		{
			name:              "error getting list",
			isCSVsupported:    true,
			isCSVsupportedErr: nil,
			list:              nil,
			listErr:           errors.New("an error"),
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       true,
		},
		{
			name:              "supported csv, empty list",
			isCSVsupported:    true,
			isCSVsupportedErr: nil,
			list:              &olm.ClusterServiceVersionList{},
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       false,
		},
		{
			name:              "supported csv, return succeeded only",
			isCSVsupported:    true,
			isCSVsupportedErr: nil,
			list: &olm.ClusterServiceVersionList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "a kind",
					APIVersion: "a version",
				},
				Items: []olm.ClusterServiceVersion{
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "Succeeded",
						},
					},
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "",
						},
					},
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "other phase",
						},
					},
				},
			},
			expectedList: &olm.ClusterServiceVersionList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "a kind",
					APIVersion: "a version",
				},
				Items: []olm.ClusterServiceVersion{
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "Succeeded",
						},
					},
				},
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kc := kclient.NewMockClientInterface(ctrl)
			kc.EXPECT().IsCSVSupported().Return(tt.isCSVsupported, tt.isCSVsupportedErr).AnyTimes()
			kc.EXPECT().ListClusterServiceVersions().Return(tt.list, tt.listErr).AnyTimes()
			got, gotErr := ListSucceededClusterServiceVersions(kc)
			if gotErr != nil != tt.expectedErr {
				t.Errorf("Got error %v, expected error %v\n", gotErr, tt.expectedErr)
			}
			if !reflect.DeepEqual(got, tt.expectedList) {
				t.Errorf("Got %v, expected %v\n", got, tt.expectedList)
			}
		})
	}
}
