package init

import "testing"

func Test_initParams_validate(t *testing.T) {
	type fields struct {
		name            string
		devfile         string
		devfileRegistry string
		starter         string
		devfilePath     string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no name passed",
			fields: fields{
				name: "",
			},
			wantErr: true,
		},
		{
			name: "no devfile info passed",
			fields: fields{
				name: "aname",
			},
			wantErr: true,
		},
		{
			name: "devfile passed",
			fields: fields{
				name:    "aname",
				devfile: "adevfile",
			},
			wantErr: false,
		},
		{
			name: "devfile and devfile-path passed",
			fields: fields{
				name:        "aname",
				devfile:     "adevfile",
				devfilePath: "apath",
			},
			wantErr: true,
		},
		{
			name: "devfile and devfile-registry passed",
			fields: fields{
				name:            "aname",
				devfile:         "adevfile",
				devfileRegistry: "aregistry",
			},
			wantErr: false,
		},
		{
			name: "devfile-path and devfile-registry passed",
			fields: fields{
				name:            "aname",
				devfilePath:     "apath",
				devfileRegistry: "aregistry",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &initParams{
				name:            tt.fields.name,
				devfile:         tt.fields.devfile,
				devfileRegistry: tt.fields.devfileRegistry,
				starter:         tt.fields.starter,
				devfilePath:     tt.fields.devfilePath,
			}
			if err := o.validate(); (err != nil) != tt.wantErr {
				t.Errorf("initParams.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
