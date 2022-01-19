package params

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/preference"
)

func Test_InitParams_Validate(t *testing.T) {
	type fields struct {
		name            string
		devfile         string
		devfileRegistry string
		starter         string
		devfilePath     string
	}
	tests := []struct {
		name               string
		fields             fields
		registryNameExists bool
		wantErr            bool
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
			registryNameExists: true,
			wantErr:            false,
		},
		{
			name: "devfile and devfile-registry passed with non existing registry",
			fields: fields{
				name:            "aname",
				devfile:         "adevfile",
				devfileRegistry: "aregistry",
			},
			registryNameExists: false,
			wantErr:            true,
		},
		{
			name: "devfile-path and devfile-registry passed",
			fields: fields{
				name:            "aname",
				devfilePath:     "apath",
				devfileRegistry: "aregistry",
			},
			registryNameExists: true,
			wantErr:            true,
		},
		{
			name: "numeric name",
			fields: fields{
				name:    "1234",
				devfile: "adevfile",
			},
			wantErr: true,
		},
		{
			name: "non DNS name",
			fields: fields{
				name:    "WrongName",
				devfile: "adevfile",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			prefClient.EXPECT().RegistryNameExists(gomock.Any()).Return(tt.registryNameExists).AnyTimes()
			o := &InitParams{
				Name:            tt.fields.name,
				Devfile:         tt.fields.devfile,
				DevfileRegistry: tt.fields.devfileRegistry,
				Starter:         tt.fields.starter,
				DevfilePath:     tt.fields.devfilePath,
			}
			if err := o.Validate(prefClient); (err != nil) != tt.wantErr {
				t.Errorf("initParams.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
