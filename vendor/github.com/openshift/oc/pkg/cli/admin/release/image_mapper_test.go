package release

import (
	"reflect"
	"strings"
	"testing"

	imageapi "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/diff"
)

func TestNewImageMapper(t *testing.T) {
	type args struct {
		images map[string]ImageReference
	}
	tests := []struct {
		name    string
		args    args
		input   string
		output  string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "empty input"},
		{
			name: "empty source repository",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						TargetPullSpec: "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate source repositories",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
					"etcd-2": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:5678",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "replace repository with tag",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "image: quay.io/coreos/etcd:latest",
			output: "image: quay.io/openshift/origin-etcd@sha256:1234",
		},
		{
			name: "replace tag with digest",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd:latest",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "image: quay.io/coreos/etcd:latest",
			output: "image: quay.io/openshift/origin-etcd@sha256:1234",
		},
		{
			name: "replace repository with tag with trailing whitespace",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "image: quay.io/coreos/etcd:latest\n",
			output: "image: quay.io/openshift/origin-etcd@sha256:1234\n",
		},
		{
			name: "replace repository with digest",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "image: quay.io/coreos/etcd@sha256:5678",
			output: "image: quay.io/openshift/origin-etcd@sha256:1234",
		},
		{
			name: "replace with digest on a multi-line file with quotes and newlines",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/openshift/origin-prometheus:latest",
						TargetPullSpec:   "quay.io/openshift/origin-prometheus@sha256:1234",
					},
				},
			},
			input: `
	- "-images=prometheus=quay.io/openshift/origin-prometheus:latest"
	- "-images=alertmanager=quay.io/openshift/origin-prometheus-alertmanager:latest"
`,
			output: `
	- "-images=prometheus=quay.io/openshift/origin-prometheus@sha256:1234"
	- "-images=alertmanager=quay.io/openshift/origin-prometheus-alertmanager:latest"
`,
		},
		{
			name: "replace with digest on a multi-line file with quotes and newlines",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/openshift/origin-prometheus:latest",
						TargetPullSpec:   "quay.io/openshift/origin-prometheus@sha256:1234",
					},
				},
			},
			input: `
	- "quay.io/openshift/origin-prometheus:latest"
`,
			output: `
	- "quay.io/openshift/origin-prometheus@sha256:1234"
`,
		},
		{
			name: "replace bare repository when told to do so",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "image: quay.io/coreos/etcd",
			output: "image: quay.io/openshift/origin-etcd@sha256:1234",
		},
		{
			name: "replace bare repository with trailing whitespace when told to do so",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "image: quay.io/coreos/etcd ",
			output: "image: quay.io/openshift/origin-etcd@sha256:1234 ",
		},
		{
			name: "Ignore things that only look like images",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "example_url: https://quay.io/coreos/etcd:8443/test",
			output: "example_url: https://quay.io/coreos/etcd:8443/test",
		},
		{
			name: "replace entire file - just to verify the regex",
			args: args{
				images: map[string]ImageReference{
					"etcd": {
						SourceRepository: "quay.io/coreos/etcd",
						TargetPullSpec:   "quay.io/openshift/origin-etcd@sha256:1234",
					},
				},
			},
			input:  "quay.io/coreos/etcd:latest",
			output: "quay.io/openshift/origin-etcd@sha256:1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewImageMapper(tt.args.images)
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if err != nil {
				return
			}
			out, err := m([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if err != nil {
				return
			}
			if string(out) != tt.output {
				t.Errorf("unexpected output, wanted\n%s\ngot\n%s", tt.output, string(out))
			}
		})
	}
}

func TestNewExactMapper(t *testing.T) {
	type args struct {
		mappings map[string]string
	}
	tests := []struct {
		name    string
		args    args
		input   string
		output  string
		wantErr bool
	}{
		{
			name:   "replace at end of file",
			args:   args{mappings: map[string]string{"reg/repo@sha256:01234": "reg2/repo2@sha256:01234"}},
			input:  "image: reg/repo@sha256:01234",
			output: "image: reg2/repo2@sha256:01234",
		},
		{
			name:   "replace at beginning of file",
			args:   args{mappings: map[string]string{"reg/repo@sha256:01234": "reg2/repo2@sha256:01234"}},
			input:  "reg/repo@sha256:01234",
			output: "reg2/repo2@sha256:01234",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewExactMapper(tt.args.mappings)
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if err != nil {
				return
			}
			out, err := m([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if err != nil {
				return
			}
			if string(out) != tt.output {
				t.Errorf("unexpected output, wanted\n%s\ngot\n%s", tt.output, string(out))
			}
		})
	}
}

func TestNewComponentVersionsMapper(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name        string
		releaseName string
		versions    ComponentVersions
		imagesByTag map[string][]string
		in          string
		out         string
		wantErr     string
	}{
		{
			in:  `version: 0.0.1-snapshot\n`,
			out: `version: 0.0.1-snapshot\n`,
		},
		{
			in:      `version: 0.0.1-snapshot-\n`,
			wantErr: `empty version references are not allowed`,
		},
		{
			in:      `version: 0.0.1-snapshot-a\n`,
			wantErr: `unknown version reference "a"`,
		},
		{
			releaseName: "2.0.0",
			in:          `version: 0.0.1-snapshot\n`,
			out:         `version: 2.0.0\n`,
		},
		{
			name:        "release name is not semver",
			releaseName: "2.0",
			in:          `version: 0.0.1-snapshot\n`,
			out:         `version: 0.0.1-snapshot\n`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			in:       `version: 0.0.1-snapshot-a\n`,
			out:      `version: 2.0.0\n`,
		},
		{
			versions:    ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			imagesByTag: map[string][]string{"a": {"tag1", "tag2"}},
			in:          `version: 0.0.1-snapshot-a\n`,
			wantErr:     `the version for "a" is inconsistent across the referenced images: tag1, tag2`,
		},
		{
			versions:    ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}, "b": ComponentVersion{Version: "3.0.0"}},
			imagesByTag: map[string][]string{"a": {"tag1", "tag2"}},
			in:          `version: 0.0.1-snapshot-b\n`,
			out:         `version: 3.0.0\n`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			in:       `version: 0.0.1-snapshot-a`,
			out:      `version: 2.0.0`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			in:       `0.0.1-snapshot-a`,
			out:      `2.0.0`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			in:       `:0.0.1-snapshot-a`,
			out:      `:2.0.0`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			in:       `-0.0.1-snapshot-a_`,
			out:      `-2.0.0_`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}},
			in:       `0.0.1-snapshot-a 0.0.1-snapshot-b`,
			wantErr:  `unknown version reference "b"`,
		},
		{
			versions: ComponentVersions{"a": ComponentVersion{Version: "2.0.0"}, "b": ComponentVersion{Version: "1.0.0"}},
			in:       `0.0.1-snapshot-a 0.0.1-snapshot-b`,
			out:      `2.0.0 1.0.0`,
		},
		{
			in:      `0.0.1-snapshot-a0.0.1-snapshot-b`,
			wantErr: `unknown version reference "a0"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewComponentVersionsMapper(tt.releaseName, tt.versions, tt.imagesByTag)
			out, err := m([]byte(tt.in))
			if (err != nil) != (len(tt.wantErr) > 0) {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if tt.out != string(out) {
				t.Errorf("mismatch:\n%s\n%s", tt.out, out)
			}
		})
	}
}

func Test_parseComponentVersionsLabel(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name         string
		label        string
		displayNames string
		want         ComponentVersions
		wantErr      bool
	}{
		{label: ""},
		{displayNames: "a=b"},
		{label: "a=1.0.0", wantErr: true},
		{label: "a!=1.0.0", wantErr: true},
		{label: "ab", wantErr: true},
		{label: "a1=1.0.0", displayNames: "a=b", wantErr: true},
		{label: "a1=1.0.0", want: ComponentVersions{"a1": {Version: "1.0.0"}}},
		{label: "a1=1.0.0", displayNames: "a1=b 1 c d : -", want: ComponentVersions{"a1": {Version: "1.0.0", DisplayName: "b 1 c d : -"}}},
		{label: "a1=1.0.0", displayNames: "a1=!", wantErr: true},
		{label: "a1=1.0.0", displayNames: "a1=!,a1=valid", wantErr: true},
		{label: "a1=1.0.0", displayNames: "a1=other,a1=valid", want: ComponentVersions{"a1": {Version: "1.0.0", DisplayName: "valid"}}},
		{label: "a1=1.0.0,b1=2.0.0", displayNames: "a1=other,a1=valid", want: ComponentVersions{"a1": {Version: "1.0.0", DisplayName: "valid"}, "b1": {Version: "2.0.0"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseComponentVersionsLabel(tt.label, tt.displayNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComponentVersionsLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseComponentVersionsLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadImageStreamTransforms(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		input              *imageapi.ImageStream
		local              *imageapi.ImageStream
		allowMissingImages bool
		src                string

		name     string
		args     args
		want     ComponentVersions
		wantTags map[string][]string
		wantRefs map[string]ImageReference
		wantErr  bool
	}{
		{
			name: "error if no source input",
			input: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{},
				},
			},
			local: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "ignored",
							From: &corev1.ObjectReference{
								Name: "a_name",
								Kind: "SomethingElse",
							},
						},
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other:value",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ignore tag for non matching kind",
			input: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			local: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "ignored",
							From: &corev1.ObjectReference{
								Name: "a_name",
								Kind: "SomethingElse",
							},
						},
					},
				},
			},
			want:     ComponentVersions{},
			wantTags: map[string][]string{},
			wantRefs: map[string]ImageReference{},
		},
		{
			name: "resolve tag",
			input: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			local: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "ignored",
							From: &corev1.ObjectReference{
								Name: "a_name",
								Kind: "SomethingElse",
							},
						},
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other:value",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			want:     ComponentVersions{},
			wantTags: map[string][]string{},
			wantRefs: map[string]ImageReference{
				"test": {
					SourceRepository: "quay.io/test/other:value",
					TargetPullSpec:   "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
				},
			},
		},
		{
			name: "resolve referenced component",
			input: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
								Kind: "DockerImage",
							},
							Annotations: map[string]string{
								annotationBuildVersions: "other=1.0.0",
							},
						},
					},
				},
			},
			local: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "ignored",
							From: &corev1.ObjectReference{
								Name: "a_name",
								Kind: "SomethingElse",
							},
						},
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other:value",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"other": {Version: "1.0.0"},
			},
			wantTags: map[string][]string{
				"other": {"test"},
			},
			wantRefs: map[string]ImageReference{
				"test": {
					SourceRepository: "quay.io/test/other:value",
					TargetPullSpec:   "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
				},
			},
		},

		{
			name: "resolve optional display name for component",
			input: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
								Kind: "DockerImage",
							},
							Annotations: map[string]string{
								annotationBuildVersions: "other=1.0.0,test=1.3.6",
							},
						},
						{
							Name: "test-2",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
								Kind: "DockerImage",
							},
							Annotations: map[string]string{
								annotationBuildVersions:             "other=1.0.0,test=1.3.6",
								annotationBuildVersionsDisplayNames: "other=Some Cool Component",
							},
						},
					},
				},
			},
			local: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "ignored",
							From: &corev1.ObjectReference{
								Name: "a_name",
								Kind: "SomethingElse",
							},
						},
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other:value",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"test":  {Version: "1.3.6"},
				"other": {Version: "1.0.0"},
			},
			wantTags: map[string][]string{
				"other": {"test"},
				"test":  {"test"},
			},
			wantRefs: map[string]ImageReference{
				"test": {
					SourceRepository: "quay.io/test/other:value",
					TargetPullSpec:   "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
				},
			},
		},
		{
			name: "conflicting resolved tags triggers an error",
			input: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000001",
								Kind: "DockerImage",
							},
							Annotations: map[string]string{
								annotationBuildVersions: "other=1.0.0,test=1.3.7",
							},
						},
						{
							Name: "test-2",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other@sha256:0000000000000000000000000000000000000002",
								Kind: "DockerImage",
							},
							Annotations: map[string]string{
								annotationBuildVersions:             "other=1.0.0,test=1.3.6",
								annotationBuildVersionsDisplayNames: "other=Some Cool Component",
							},
						},
					},
				},
			},
			local: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/other:value",
								Kind: "DockerImage",
							},
						},
						{
							Name: "test-2",
							From: &corev1.ObjectReference{
								Name: "quay.io/test/test:value",
								Kind: "DockerImage",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := loadImageStreamTransforms(tt.input, tt.local, tt.allowMissingImages, tt.src)
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadImageStreamTransforms() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s", diff.ObjectReflectDiff(got, tt.want))
			}
			if !reflect.DeepEqual(got1, tt.wantTags) {
				t.Errorf("%s", diff.ObjectReflectDiff(got1, tt.wantTags))
			}
			if !reflect.DeepEqual(got2, tt.wantRefs) {
				t.Errorf("%s", diff.ObjectReflectDiff(got2, tt.wantRefs))
			}
		})
	}
}
