package occlient

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestGetOcBinary(t *testing.T) {

	// test setup
	// test shouldn't have external dependency, so we are faking oc binary with empty tmpfile
	tmpfile, err := ioutil.TempFile("", "fake-oc")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile1, err := ioutil.TempFile("", "fake-oc")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer os.Remove(tmpfile1.Name())

	type args struct {
		oc string
	}
	tests := []struct {
		name    string
		envs    map[string]string
		want    string
		wantErr bool
	}{
		{
			name: "set via KUBECTL_PLUGINS_CALLER exists",
			envs: map[string]string{
				"KUBECTL_PLUGINS_CALLER": tmpfile.Name(),
			},
			want:    tmpfile.Name(),
			wantErr: false,
		},
		{
			name: "set via KUBECTL_PLUGINS_CALLER (invalid file)",
			envs: map[string]string{
				"KUBECTL_PLUGINS_CALLER": "invalid",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "set via OC_BIN exists",
			envs: map[string]string{
				"OC_BIN": tmpfile.Name(),
			},
			want:    tmpfile.Name(),
			wantErr: false,
		},
		{
			name: "set via OC_BIN (invalid file)",
			envs: map[string]string{
				"OC_BIN": "invalid",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "bot OC_BIN and KUBECTL_PLUGINS_CALLER set",
			envs: map[string]string{
				"OC_BIN":                 tmpfile.Name(),
				"KUBECTL_PLUGINS_CALLER": tmpfile1.Name(),
			},
			want:    tmpfile1.Name(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// cleanup variables before running test
			os.Unsetenv("OC_BIN")
			os.Unsetenv("KUBECTL_PLUGINS_CALLER")

			for k, v := range tt.envs {
				if err := os.Setenv(k, v); err != nil {
					t.Error(err)
				}
			}
			got, err := getOcBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("getOcBinary() unexpected error \n%v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOcBinary() \ngot: %v \nwant: %v", got, tt.want)
			}
		})
	}
}

func TestAddLabelsToArgs(t *testing.T) {
	tests := []struct {
		name     string
		argsIn   []string
		labels   map[string]string
		argsOut1 []string
		argsOut2 []string
	}{
		{
			name:   "one label in empty args",
			argsIn: []string{},
			labels: map[string]string{
				"label1": "value1",
			},
			argsOut1: []string{
				"--labels", "label1=value1",
			},
		},
		{
			name: "one label with existing args",
			argsIn: []string{
				"--foo", "bar",
			},
			labels: map[string]string{
				"label1": "value1",
			},
			argsOut1: []string{
				"--foo", "bar",
				"--labels", "label1=value1",
			},
		},
		{
			name: "multiple label with existing args",
			argsIn: []string{
				"--foo", "bar",
			},
			labels: map[string]string{
				"label1": "value1",
				"label2": "value2",
			},
			argsOut1: []string{
				"--foo", "bar",
				"--labels", "label1=value1,label2=value2",
			},
			argsOut2: []string{
				"--foo", "bar",
				"--labels", "label2=value2,label1=value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsGot := addLabelsToArgs(tt.labels, tt.argsIn)

			if !reflect.DeepEqual(argsGot, tt.argsOut1) && !reflect.DeepEqual(argsGot, tt.argsOut2) {
				t.Errorf("addLabelsToArgs() \ngot:  %#v \nwant: %#v or %#v", argsGot, tt.argsOut1, tt.argsOut2)
			}
		})
	}
}
