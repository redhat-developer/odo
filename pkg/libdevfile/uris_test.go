package libdevfile

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/api/v2/pkg/validation"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	context "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/libdevfile/generator"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
)

func TestGetReferencedLocalFiles(t *testing.T) {

	imageComponentNoDockerfile := generator.GetImageComponent(generator.ImageComponentParams{
		Name: "image-component",
		Image: v1alpha2.Image{
			ImageName: "an-image-name",
		},
	})
	imageComponentWithHTTPDockerfile := generator.GetImageComponent(generator.ImageComponentParams{
		Name: "image-component",
		Image: v1alpha2.Image{
			ImageName: "an-image-name",
			ImageUnion: v1alpha2.ImageUnion{
				Dockerfile: &v1alpha2.DockerfileImage{
					DockerfileSrc: v1alpha2.DockerfileSrc{
						Uri: "http://example.com",
					},
				},
			},
		},
	})
	imageComponentWithLocalDockerfile := generator.GetImageComponent(generator.ImageComponentParams{
		Name: "image-component",
		Image: v1alpha2.Image{
			ImageName: "an-image-name",
			ImageUnion: v1alpha2.ImageUnion{
				Dockerfile: &v1alpha2.DockerfileImage{
					DockerfileSrc: v1alpha2.DockerfileSrc{
						Uri: "path/to/Dockerfile",
					},
				},
			},
		},
	})

	kubeComponentInlined := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
		Name: "kube-component",
		Kubernetes: &v1alpha2.KubernetesComponent{
			K8sLikeComponent: v1alpha2.K8sLikeComponent{
				K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
					Inlined: "",
				},
			},
		},
	})
	kubeComponentHTTPUri := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
		Name: "kube-component",
		Kubernetes: &v1alpha2.KubernetesComponent{
			K8sLikeComponent: v1alpha2.K8sLikeComponent{
				K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
					Uri: "http://example.com",
				},
			},
		},
	})
	kubeComponentLocalUri := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
		Name: "kube-component",
		Kubernetes: &v1alpha2.KubernetesComponent{
			K8sLikeComponent: v1alpha2.K8sLikeComponent{
				K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
					Uri: "path/to/manifest",
				},
			},
		},
	})

	openshiftComponentInlined := generator.GetOpenshiftComponent(generator.OpenshiftComponentParams{
		Name: "openshift-component",
		Openshift: &v1alpha2.OpenshiftComponent{
			K8sLikeComponent: v1alpha2.K8sLikeComponent{
				K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
					Inlined: "",
				},
			},
		},
	})
	openshiftComponentHTTPUri := generator.GetOpenshiftComponent(generator.OpenshiftComponentParams{
		Name: "openshift-component",
		Openshift: &v1alpha2.OpenshiftComponent{
			K8sLikeComponent: v1alpha2.K8sLikeComponent{
				K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
					Uri: "http://example.com",
				},
			},
		},
	})
	openshiftComponentLocalUri := generator.GetOpenshiftComponent(generator.OpenshiftComponentParams{
		Name: "openshift-component",
		Openshift: &v1alpha2.OpenshiftComponent{
			K8sLikeComponent: v1alpha2.K8sLikeComponent{
				K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
					Uri: "path/to/manifest",
				},
			},
		},
	})

	type args struct {
		devfileObj func(fs filesystem.Filesystem) parser.DevfileObj
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "image without Dockerfile",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{imageComponentNoDockerfile})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name: "image with HTTP Dockerfile",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{imageComponentWithHTTPDockerfile})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name: "image with local Dockerfile",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{imageComponentWithLocalDockerfile})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{"path/to/Dockerfile"},
			wantErr:    false,
		},

		{
			name: "inlined Kubernetes component",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{kubeComponentInlined})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name: "Kubernetes component with HTTP uri",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{kubeComponentHTTPUri})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name: "Kubernetes component with local uri",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{kubeComponentLocalUri})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{"path/to/manifest"},
			wantErr:    false,
		},

		{
			name: "inlined Openshift component",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{openshiftComponentInlined})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name: "Openshift component with HTTP uri",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{openshiftComponentHTTPUri})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name: "Openshift component with local uri",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddComponents([]v1alpha2.Component{openshiftComponentLocalUri})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: []string{"path/to/manifest"},
			wantErr:    false,
		},

		{
			name: "With parent Devfile, non flattened",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					parentData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					parentObj := parser.DevfileObj{
						Data: parentData,
						Ctx:  context.FakeContext(fs, "/path/to/parent/devfile.yaml"),
					}
					_ = parentObj.WriteYamlDevfile()

					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					dData.SetParent(&v1alpha2.Parent{
						ImportReference: v1alpha2.ImportReference{
							ImportReferenceUnion: v1alpha2.ImportReferenceUnion{
								Uri: "/path/to/parent/devfile.yaml",
							},
						},
					})
					return parser.DevfileObj{
						Data: dData,
					}
				},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "With parent Devfile, flattened",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					obj, _, _ := devfile.ParseDevfileAndValidate(parser.ParserArgs{
						Path:             "testdata/child-devfile.yaml",
						FlattenedDevfile: pointer.Bool(true),
					})
					return obj
				},
			},
			wantResult: []string{
				"manifest.yaml", // TODO should be parent/manifest.yaml (see https://github.com/devfile/api/issues/904)
				"parent/parent-devfile.yaml",
			},
			wantErr: false,
		},
		{
			name: "With parent Devfile containing only commands, flattened",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					obj, _, _ := devfile.ParseDevfileAndValidate(parser.ParserArgs{
						Path:             "testdata/child-devfile-components-only.yaml",
						FlattenedDevfile: pointer.Bool(true),
					})
					return obj
				},
			},
			wantResult: []string{
				"manifest.yaml", // TODO should be parent/manifest.yaml (see https://github.com/devfile/api/issues/904)
				"parent/parent-devfile-commands-only.yaml",
			},
			wantErr: false,
		},
		{
			name: "With parent Devfile containing only components, flattened",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					obj, _, _ := devfile.ParseDevfileAndValidate(parser.ParserArgs{
						Path:             "testdata/child-devfile-commands-only.yaml",
						FlattenedDevfile: pointer.Bool(true),
					})
					return obj
				},
			},
			wantResult: []string{
				"manifest.yaml", // TODO should be parent/manifest.yaml (see https://github.com/devfile/api/issues/904)
				"parent/parent-devfile-components-only.yaml",
			},
			wantErr: false,
		},
		{
			name: "With empty parent Devfile, flattened. TODO find a way to detect the parent",
			args: args{
				devfileObj: func(fs filesystem.Filesystem) parser.DevfileObj {
					obj, _, _ := devfile.ParseDevfileAndValidate(parser.ParserArgs{
						Path:             "testdata/child-devfile-complete.yaml",
						FlattenedDevfile: pointer.Bool(true),
					})
					return obj
				},
			},
			wantResult: []string{
				"manifest.yaml", // TODO should be parent/manifest.yaml (see https://github.com/devfile/api/issues/904)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := filesystem.NewFakeFs()
			gotResult, err := GetReferencedLocalFiles(tt.args.devfileObj(fs))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReferencedLocalFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantResult, gotResult); diff != "" {
				t.Errorf("GetReferencedLocalFiles() wantResult mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_appendUriIfFile(t *testing.T) {
	type args struct {
		result map[string]struct{}
		uri    string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "empty uri",
			args: args{
				result: map[string]struct{}{},
				uri:    "",
			},
			want:    map[string]struct{}{},
			wantErr: false,
		},
		{
			name: "http uri",
			args: args{
				result: map[string]struct{}{},
				uri:    "http://example.com",
			},
			want:    map[string]struct{}{},
			wantErr: false,
		},
		{
			name: "file uri",
			args: args{
				result: map[string]struct{}{},
				uri:    "path/to/file",
			},
			want: map[string]struct{}{
				"path/to/file": {},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appendUriIfFile(tt.args.result, tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("appendUriIfFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("appendUriIfFile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

/*
func Test_getFromAttributes(t *testing.T) {
	type args struct {
		result     map[string]struct{}
		attributes attributes.Attributes
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "todo",
			args: args{
				result: map[string]struct{}{},
				attributes: attributes.Attributes{
					validation.ImportSourceAttribute: apiext.JSON{
						Raw: []byte("uri: path/to/devfile.yaml"),
					},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getFromAttributes(tt.args.result, tt.args.attributes)
		})
	}
}
*/

func Test_getFromAttributes(t *testing.T) {
	type args struct {
		result     map[string]struct{}
		attributes attributes.Attributes
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "uri is local file",
			args: args{
				result: map[string]struct{}{},
				attributes: attributes.Attributes{
					validation.ImportSourceAttribute: apiext.JSON{
						Raw: []byte("uri: path/to/devfile.yaml"),
					},
				},
			},
			want: map[string]struct{}{
				"path/to/devfile.yaml": {},
			},
		},
		{
			name: "uri is http file",
			args: args{
				result: map[string]struct{}{},
				attributes: attributes.Attributes{
					validation.ImportSourceAttribute: apiext.JSON{
						Raw: []byte("uri: http://example.com/devfile.yaml"),
					},
				},
			},
			want: map[string]struct{}{},
		},
		{
			name: "id and registryURL",
			args: args{
				result: map[string]struct{}{},
				attributes: attributes.Attributes{
					validation.ImportSourceAttribute: apiext.JSON{
						Raw: []byte("id: my-id, registryURL: http://myregistry.com"),
					},
				},
			},
			want: map[string]struct{}{},
		},
		{
			name: "name and namespace",
			args: args{
				result: map[string]struct{}{},
				attributes: attributes.Attributes{
					validation.ImportSourceAttribute: apiext.JSON{
						Raw: []byte("name: aname, namespace: anamespace"),
					},
				},
			},
			want: map[string]struct{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFromAttributes(tt.args.result, tt.args.attributes)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFromAttributes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getFromAttributes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
