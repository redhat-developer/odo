package params

import (
	"reflect"
	"testing"
)

func TestFlagsBuilder_ParamsBuild(t *testing.T) {
	type fields struct {
		flags map[string]string
	}
	tests := []struct {
		name         string
		fields       fields
		wantAdequate bool
		want         InitParams
		wantErr      bool
	}{
		{
			name: "no field defined",
			fields: fields{
				flags: map[string]string{},
			},
			wantAdequate: false,
			wantErr:      false,
			want:         InitParams{},
		},
		{
			name: "all fields defined",
			fields: fields{
				flags: map[string]string{
					FLAG_NAME:             "aname",
					FLAG_DEVFILE:          "adevfile",
					FLAG_DEVFILE_PATH:     "apath",
					FLAG_DEVFILE_REGISTRY: "aregistry",
					FLAG_STARTER:          "astarter",
				},
			},
			wantAdequate: true,
			wantErr:      false,
			want: InitParams{
				Name:            "aname",
				Devfile:         "adevfile",
				DevfilePath:     "apath",
				DevfileRegistry: "aregistry",
				Starter:         "astarter",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBuilder{}
			adequate := o.IsAdequate(tt.fields.flags)
			if adequate != tt.wantAdequate {
				t.Errorf("IsAdequate should return %v but returns %v", tt.wantAdequate, adequate)
			}
			got, err := o.ParamsBuild()
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBuilder.ParamsBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlagsBuilder.ParamsBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}
