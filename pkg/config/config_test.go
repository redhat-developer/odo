package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestGetOcDevConfigFile(t *testing.T) {
	// TODO: implement this
}

func TestNew(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "ocdevconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(configEnvName, tempConfigFile.Name())

	tests := []struct {
		name    string
		output  *ConfigInfo
		success bool
	}{
		{
			name: "Test filename is being set",
			output: &ConfigInfo{
				Filename: tempConfigFile.Name(),
			},
			success: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfi, err := New()
			switch test.success {
			case true:
				if err != nil {
					t.Errorf("Expected test to pass, but it failed with error: %v", err)
				}
			case false:
				if err == nil {
					t.Errorf("Expected test to fail, but it passed!")
				}
			}
			if !reflect.DeepEqual(test.output, cfi) {
				t.Errorf("Expected output: %#v", test.output)
				t.Errorf("Actual output: %#v", cfi)
			}
		})
	}
}

//
//func TestGet(t *testing.T) {
//
//}
//
//func TestSet(t *testing.T) {
//
//}
//
//func TestApplicationExists(t *testing.T) {
//
//}
//
//func TestAddApplication(t *testing.T) {
//
//}
