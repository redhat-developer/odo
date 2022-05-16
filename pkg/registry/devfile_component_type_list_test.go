package registry

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/api"
)

func TestDevfileStackList_GetLanguages(t *testing.T) {
	type fields struct {
		Items []api.DevfileStack
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "no devfiles",
			want: []string{},
		},
		{
			name: "some devfiles",
			fields: fields{
				Items: []api.DevfileStack{
					{
						Name:        "devfile4",
						DisplayName: "first devfile for lang3",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang3",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: api.Registry{
							Name: "Registry2",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile3",
						DisplayName: "another devfile for lang2",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang2",
					},
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
				},
			},
			want: []string{"lang1", "lang2", "lang3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DevfileStackList{
				Items: tt.fields.Items,
			}
			if got := o.GetLanguages(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevfileStackList.GetLanguages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevfileStackList_GetProjectTypes(t *testing.T) {
	type fields struct {
		Items []api.DevfileStack
	}
	type args struct {
		language string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   TypesWithDetails
	}{
		{
			name: "No devfiles => no project types",
			want: TypesWithDetails{},
		},
		{
			name: "project types for lang1",
			fields: fields{
				Items: []api.DevfileStack{
					{
						Name:        "devfile4",
						DisplayName: "first devfile for lang3",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang3",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: api.Registry{
							Name: "Registry2",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile3",
						DisplayName: "another devfile for lang2",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang2",
					},
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Registry: api.Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
				},
			},
			args: args{
				language: "lang1",
			},
			want: TypesWithDetails{
				"first devfile for lang1": []api.DevfileStack{
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Language:    "lang1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Language:    "lang1",
						Registry: api.Registry{
							Name: "Registry2",
						},
					},
				},
				"second devfile for lang1": []api.DevfileStack{
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Language:    "lang1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DevfileStackList{
				Items: tt.fields.Items,
			}
			if got := o.GetProjectTypes(tt.args.language); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevfileStackList.GetProjectTypes() = \n%+v, want \n%+v", got, tt.want)
			}
		})
	}
}
