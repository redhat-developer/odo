package params

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/preference"
)

func Test_DevfileLocation_Validate(t *testing.T) {
	type fields struct {
		devfile         string
		devfileRegistry string
		devfilePath     string
	}
	tests := []struct {
		name               string
		fields             fields
		registryNameExists bool
		registryList       []preference.Registry
		wantErr            bool
	}{
		{
			name:    "no devfile info passed",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "devfile passed with a single registry",
			fields: fields{
				devfile: "adevfile",
			},
			registryList: []preference.Registry{
				{
					Name: "aregistry",
				},
			},
			registryNameExists: true,
			wantErr:            false,
		},
		{
			name: "devfile and devfile-path passed",
			fields: fields{
				devfile:     "adevfile",
				devfilePath: "apath",
			},
			wantErr: true,
		},
		{
			name: "devfile and devfile-registry passed",
			fields: fields{
				devfile:         "adevfile",
				devfileRegistry: "aregistry",
			},
			registryNameExists: true,
			wantErr:            false,
		},
		{
			name: "devfile and devfile-registry passed with non existing registry",
			fields: fields{
				devfile:         "adevfile",
				devfileRegistry: "aregistry",
			},
			registryNameExists: false,
			wantErr:            true,
		},
		{
			name: "devfile-path and devfile-registry passed",
			fields: fields{
				devfilePath:     "apath",
				devfileRegistry: "aregistry",
			},
			registryNameExists: true,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			prefClient.EXPECT().RegistryNameExists(gomock.Any()).Return(tt.registryNameExists).AnyTimes()
			prefClient.EXPECT().RegistryList().Return(&tt.registryList).AnyTimes()
			o := &DevfileLocation{
				Devfile:         tt.fields.devfile,
				DevfileRegistry: tt.fields.devfileRegistry,
				DevfilePath:     tt.fields.devfilePath,
			}
			if err := o.Validate(prefClient); (err != nil) != tt.wantErr {
				t.Errorf("initParams.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
