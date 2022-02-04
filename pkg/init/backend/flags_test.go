package backend

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestFlagsBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		flags map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantOk  bool
		want    *DevfileLocation
		wantErr bool
	}{
		{
			name: "no field defined",
			fields: fields{
				flags: map[string]string{},
			},
			wantOk:  false,
			wantErr: false,
			want:    nil,
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
			wantOk:  true,
			wantErr: false,
			want: &DevfileLocation{
				Devfile:         "adevfile",
				DevfilePath:     "apath",
				DevfileRegistry: "aregistry",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{}
			ok, got, err := o.SelectDevfile(tt.fields.flags)
			if ok != tt.wantOk {
				t.Errorf("FlagsBackend.SelectDevfile() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.SelectDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlagsBackend.SelectDevfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlagsBackend_Validate(t *testing.T) {
	type fields struct {
	}
	type args struct {
		flags map[string]string
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		registryNameExists bool
		registryList       []preference.Registry
		wantErr            bool
	}{
		{
			name: "no args",
			args: args{
				flags: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "no name passed",
			args: args{
				flags: map[string]string{
					"name": "",
				},
			},
			wantErr: true,
		},
		{
			name: "no devfile info passed",
			args: args{
				flags: map[string]string{
					"name": "aname",
				},
			},
			wantErr: true,
		},
		{
			name: "devfile passed with a single registry",
			args: args{
				flags: map[string]string{
					"name":    "aname",
					"devfile": "adevfile",
				},
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
			args: args{
				flags: map[string]string{
					"name":         "aname",
					"devfile":      "adevfile",
					"devfile-path": "apath",
				},
			},
			wantErr: true,
		},
		{
			name: "devfile and devfile-registry passed",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile":          "adevfile",
					"devfile-registry": "aregistry",
				},
			},
			registryNameExists: true,
			wantErr:            false,
		},
		{
			name: "devfile and devfile-registry passed with non existing registry",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile":          "adevfile",
					"devfile-registry": "aregistry",
				},
			},
			registryNameExists: false,
			wantErr:            true,
		},
		{
			name: "devfile-path and devfile-registry passed",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile-path":     "apath",
					"devfile-registry": "aregistry",
				},
			},
			registryNameExists: true,
			wantErr:            true,
		},
		{
			name: "numeric name",
			args: args{
				flags: map[string]string{
					"name":    "1234",
					"devfile": "adevfile",
				},
			},
			wantErr: true,
		},
		{
			name: "non DNS name",
			args: args{
				flags: map[string]string{
					"name":    "WrongName",
					"devfile": "adevfile",
				},
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			prefClient.EXPECT().RegistryNameExists(gomock.Any()).Return(tt.registryNameExists).AnyTimes()
			prefClient.EXPECT().RegistryList().Return(&tt.registryList).AnyTimes()

			o := &FlagsBackend{
				preferenceClient: prefClient,
			}
			if err := o.Validate(tt.args.flags); (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
