package adapters

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/google/go-cmp/cmp"
)

func TestGetSyncFilesFromAttributes(t *testing.T) {
	type args struct {
		command v1alpha2.Command
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "no attributes",
			args: args{
				command: v1alpha2.Command{},
			},
			want: make(map[string]string),
		},
		{
			name: "no matching attributes",
			args: args{
				command: v1alpha2.Command{
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"some-custom-attribute-key":    "some-value",
						"another-custom-attribute-key": "some-value",
					}),
				},
			},
			want: make(map[string]string),
		},
		{
			name: "dev.odo.push.path attribute key",
			args: args{
				command: v1alpha2.Command{
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"dev.odo.push.path": "some-value",
					}),
				},
			},
			want: make(map[string]string),
		},
		{
			name: "attribute with only matching prefix as key",
			args: args{
				command: v1alpha2.Command{
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						_devPushPathAttributePrefix: "server/",
					}),
				},
			},
			want: map[string]string{
				".": "server",
			},
		},
		{
			name: "multiple matching attributes",
			args: args{
				command: v1alpha2.Command{
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"some-custom-attribute-key":                      "some-value",
						_devPushPathAttributePrefix + "server.js":        "server/server.js",
						"some-other-custom-attribute-key":                "some-value",
						_devPushPathAttributePrefix + "some-path/README": "another/nested/path/README.md",
						_devPushPathAttributePrefix + "random-file.txt":  "/tmp/rand-file.txt",
					}),
				},
			},
			want: map[string]string{
				"server.js":        "server/server.js",
				"some-path/README": "another/nested/path/README.md",
				"random-file.txt":  "/tmp/rand-file.txt",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSyncFilesFromAttributes(tt.args.command)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetSyncFilesFromAttributes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
