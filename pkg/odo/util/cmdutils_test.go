package util

import (
	"reflect"
	"testing"
)

func TestGetFullName(t *testing.T) {
	parent := "odo foo"
	child := "bar"
	expected := parent + " " + child
	actual := GetFullName(parent, child)
	if expected != actual {
		t.Errorf("test failed, expected %s, got %s", expected, actual)
	}
}

func TestMapFromParameters1(t *testing.T) {
	type args struct {
		params []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "Case 1: All valid parameters with =",
			args:    args{params: []string{"key1=value1", "key2=value2"}},
			want:    map[string]string{"key1": "value1", "key2": "value2"},
			wantErr: false,
		},
		{
			name:    "Case 2: One invalid parameter without =",
			args:    args{params: []string{"key1=value1", "key2 value2"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapFromParameters(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapFromParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapFromParameters() got = %v, want %v", got, tt.want)
			}
		})
	}
}
