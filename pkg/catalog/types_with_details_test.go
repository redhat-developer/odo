package catalog

import (
	"reflect"
	"testing"
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
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
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
			if got := tt.types.GetOrderedLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypesWithDetails.GetOrderedLabels() = %v, want %v", got, tt.want)
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
		want    DevfileComponentType
		wantErr bool
	}{
		{
			name: "get a pos 0",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 0,
			},
			want: DevfileComponentType{
				Name: "devfile1",
				Registry: Registry{
					Name: "Registry1",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 1",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 1,
			},
			want: DevfileComponentType{
				Name: "devfile1",
				Registry: Registry{
					Name: "Registry2",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 2",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 2,
			},
			want: DevfileComponentType{
				Name: "devfile2",
				Registry: Registry{
					Name: "Registry1",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 4: not found",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 4,
			},
			want:    DevfileComponentType{},
			wantErr: true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.types.GetAtOrderedPosition(tt.args.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypesWithDetails.GetAtOrderedPosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypesWithDetails.GetAtOrderedPosition() got = %v, want %v", got, tt.want)
			}
		})
	}
}
