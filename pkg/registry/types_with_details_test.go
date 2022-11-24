package registry

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/api"
)

func TestTypesWithDetails_GetOrderedLabels(t *testing.T) {
	tests := []struct {
		name  string
		types TypesWithDetails
		want  []string
	}{
		{
			name: "some entries",
			types: TypesWithDetails{
				"second devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile2",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry2",
						},
					},
				},
			},
			want: []string{
				"first devfile for lang1 (devfile1, registry: Registry1)",
				"first devfile for lang1 (devfile1, registry: Registry2)",
				"second devfile for lang1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.types.GetOrderedLabels()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetOrderedLabels() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTypesWithDetails_GetAtOrderedPosition(t *testing.T) {
	type args struct {
		pos int
	}
	tests := []struct {
		name    string
		types   TypesWithDetails
		args    args
		want    api.DevfileStack
		wantErr bool
	}{
		{
			name: "get a pos 0",
			types: TypesWithDetails{
				"second devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile2",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 0,
			},
			want: api.DevfileStack{
				Name: "devfile1",
				Registry: api.Registry{
					Name: "Registry1",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 1",
			types: TypesWithDetails{
				"second devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile2",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 1,
			},
			want: api.DevfileStack{
				Name: "devfile1",
				Registry: api.Registry{
					Name: "Registry2",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 2",
			types: TypesWithDetails{
				"second devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile2",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 2,
			},
			want: api.DevfileStack{
				Name: "devfile2",
				Registry: api.Registry{
					Name: "Registry1",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 4: not found",
			types: TypesWithDetails{
				"second devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile2",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []api.DevfileStack{
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: api.Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 4,
			},
			want:    api.DevfileStack{},
			wantErr: true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.types.GetAtOrderedPosition(tt.args.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypesWithDetails.GetAtOrderedPosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetAtOrderedPosition() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
