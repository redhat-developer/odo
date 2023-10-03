package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
)

func TestDevfileState_SetMetadata(t *testing.T) {
	type args struct {
		name              string
		version           string
		displayName       string
		description       string
		tags              string
		architectures     string
		icon              string
		globalMemoryLimit string
		projectType       string
		language          string
		website           string
		provider          string
		supportUrl        string
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Set metadata",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:              "a-name",
				version:           "v1.1.1",
				displayName:       "a display name",
				description:       "a description",
				tags:              "tag1,tag2",
				architectures:     "arch1,arch2",
				icon:              "an.ico",
				globalMemoryLimit: "1Gi",
				projectType:       "a project type",
				language:          "a language",
				website:           "http://example.com",
				provider:          "a provider",
				supportUrl:        "http://support.example.com",
			},
			want: DevfileContent{
				Content: `metadata:
  architectures:
  - arch1
  - arch2
  description: a description
  displayName: a display name
  globalMemoryLimit: 1Gi
  icon: an.ico
  language: a language
  name: a-name
  projectType: a project type
  provider: a provider
  supportUrl: http://support.example.com
  tags:
  - tag1
  - tag2
  version: v1.1.1
  website: http://example.com
schemaVersion: 2.2.0
`,
				Version:    "2.2.0",
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Metadata: Metadata{
					Name:              "a-name",
					Version:           "v1.1.1",
					DisplayName:       "a display name",
					Description:       "a description",
					Tags:              "tag1,tag2",
					Architectures:     "arch1,arch2",
					Icon:              "an.ico",
					GlobalMemoryLimit: "1Gi",
					ProjectType:       "a project type",
					Language:          "a language",
					Website:           "http://example.com",
					Provider:          "a provider",
					SupportUrl:        "http://support.example.com",
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.SetMetadata(tt.args.name, tt.args.version, tt.args.displayName, tt.args.description, tt.args.tags, tt.args.architectures, tt.args.icon, tt.args.globalMemoryLimit, tt.args.projectType, tt.args.language, tt.args.website, tt.args.provider, tt.args.supportUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.SetMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.SetMetadata() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
