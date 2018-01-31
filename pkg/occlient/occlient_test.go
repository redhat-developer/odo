package occlient

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestNewOcClient(t *testing.T) {

	// test setup
	// test shouldn't have external dependency, so we are faking oc binary with empty tmpfile
	tmpfile, err := ioutil.TempFile("", "fake-oc")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	type args struct {
		oc string
	}
	tests := []struct {
		name    string
		args    args
		want    *OcClient
		wantErr bool
	}{
		{
			name: "oc path exists",
			args: args{
				oc: tmpfile.Name(),
			},
			want: &OcClient{
				oc: tmpfile.Name(),
			},
			wantErr: false,
		},
		{
			name: "oc path doesn't exists",
			args: args{
				oc: "/non/existing",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOcClient(tt.args.oc)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOcClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOcClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
