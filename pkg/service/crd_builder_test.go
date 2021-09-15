package service

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBuildCRDFromParams(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "params ok",
			params: map[string]string{
				"u":     "1",
				"a.b.c": "2",
				"a.b.d": "3",
				"a.B":   "4",
			},
			want: map[string]interface{}{
				"u": int64(1),
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": int64(2),
						"d": int64(3),
					},
					"B": int64(4),
				},
			},
			wantErr: false,
		},
		{
			name: "typed params",
			params: map[string]string{
				"a.bool":   "true",
				"a.string": "foobar",
				"a.float":  "1.234",
			},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"bool":   true,
					"string": "foobar",
					"float":  1.234,
				},
			},
			wantErr: false,
		},
		{
			name: "params error map defined before value",
			params: map[string]string{
				"u":     "1",
				"a.b.c": "2",
				"a.b":   "3",
				"a.B":   "4",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "params error value defined before map",
			params: map[string]string{
				"u":     "1",
				"a.b":   "2",
				"a.b.c": "3",
				"a.B":   "4",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := BuildCRDFromParams(tt.params, "a group", "a version", "a kind")
			if gotErr != nil != tt.wantErr {
				t.Errorf("got err: %v, expected err: %v\n", gotErr != nil, tt.wantErr)
			}
			if gotErr == nil {
				if !reflect.DeepEqual(got["spec"], tt.want) {
					jsonGot, _ := json.Marshal(got["spec"])
					jsonWant, _ := json.Marshal(tt.want)
					t.Errorf("\ngot:  %+v\n\nwant: %v\n", string(jsonGot), string(jsonWant))
				}
			}
		})
	}
}
