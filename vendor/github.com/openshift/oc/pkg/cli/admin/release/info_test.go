package release

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/diff"

	imageapi "github.com/openshift/api/image/v1"
)

func Test_contentStream_Read(t *testing.T) {
	tests := []struct {
		name    string
		parts   [][]byte
		want    string
		wantN   int64
		wantErr bool
	}{
		{
			parts: [][]byte{[]byte("test"), []byte("other"), []byte("a")},
			want:  "testothera",
			wantN: 10,
		},
		{
			parts: [][]byte{[]byte("test"), []byte(strings.Repeat("a", 4096))},
			want:  "test" + strings.Repeat("a", 4096),
			wantN: 4100,
		},
		{
			parts: nil,
			want:  "",
			wantN: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			s := &contentStream{
				parts: tt.parts,
			}
			gotN, err := io.Copy(buf, s)
			if (err != nil) != tt.wantErr {
				t.Errorf("contentStream.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("expected %d but got %d", tt.wantN, gotN)
			}
			if !bytes.Equal([]byte(tt.want), buf.Bytes()) {
				t.Errorf("contentStream.Read():\n%s\n%s", hex.Dump(buf.Bytes()), hex.Dump([]byte(tt.want)))
			}
		})
	}
}

func Test_readComponentVersions(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name    string
		is      *imageapi.ImageStream
		want    ComponentVersions
		wantErr []error
	}{
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Annotations: map[string]string{
								annotationBuildVersions:             "",
								annotationBuildVersionsDisplayNames: "",
							},
						},
					},
				},
			},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0",
								annotationBuildVersionsDisplayNames: "",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0"},
			},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0,b1=1.0.1",
								annotationBuildVersionsDisplayNames: "b1=Test Name",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0"},
				"b1": {Version: "1.0.1", DisplayName: "Test Name"},
			},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test1",
							Annotations: map[string]string{
								annotationBuildVersions: "a1=",
							},
						},
					},
				},
			},
			wantErr: []error{fmt.Errorf("the referenced image test1 had an invalid version annotation: the version pair \"a1=\" must have a valid semantic version: Version string empty")},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test1",
							Annotations: map[string]string{
								annotationBuildVersions: "a1=1.0.0",
							},
						},
						{
							Name: "test2",
							Annotations: map[string]string{
								annotationBuildVersions: "a1=1.0.0",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0"},
			},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test1",
							Annotations: map[string]string{
								annotationBuildVersions: "a1=1.0.0",
							},
						},
						{
							Name: "test2",
							Annotations: map[string]string{
								annotationBuildVersions: "a1=1.0.1",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0"},
			},
			wantErr: []error{fmt.Errorf("multiple versions or display names reported for the following component(s): a1")},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test1",
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0",
								annotationBuildVersionsDisplayNames: "a1=Test Name",
							},
						},
						{
							Name: "test2",
							Annotations: map[string]string{
								annotationBuildVersions: "a1=1.0.1",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0", DisplayName: ""},
			},
			wantErr: []error{fmt.Errorf("multiple versions or display names reported for the following component(s): a1")},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test1",
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0",
								annotationBuildVersionsDisplayNames: "a1=Test Name",
							},
						},
						{
							Name: "test2",
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0",
								annotationBuildVersionsDisplayNames: "a1=Test Name",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0", DisplayName: "Test Name"},
			},
		},
		{
			is: &imageapi.ImageStream{
				Spec: imageapi.ImageStreamSpec{
					Tags: []imageapi.TagReference{
						{
							Name: "test1",
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0",
								annotationBuildVersionsDisplayNames: "a1=Test Name",
							},
						},
						{
							Name: "test2",
							Annotations: map[string]string{
								annotationBuildVersions:             "a1=1.0.0",
								annotationBuildVersionsDisplayNames: "a1=Test Name 2",
							},
						},
					},
				},
			},
			want: ComponentVersions{
				"a1": {Version: "1.0.0", DisplayName: "Test Name"},
			},
			wantErr: []error{fmt.Errorf("multiple versions or display names reported for the following component(s): a1")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := readComponentVersions(tt.is)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s", diff.ObjectReflectDiff(got, tt.want))
			}
			if a, b := asStrings(got1), asStrings(tt.wantErr); !reflect.DeepEqual(a, b) {
				t.Errorf("%s", diff.ObjectReflectDiff(a, b))
			}
		})
	}
}

func asStrings(a []error) []string {
	if a == nil {
		return nil
	}
	var out []string
	for _, err := range a {
		out = append(out, err.Error())
	}
	return out
}
