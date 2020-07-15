package util

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/preference"
)

const (
	// GlobalConfigEnvName is the environment variable GLOBALODOCONFIG
	GlobalConfigEnvName = "GLOBALODOCONFIG"
)

func TestIsSecure(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name              string
		registryOperation string
		registryName      string
		registryURL       string
		forceFlag         bool
		isSecure          bool
		want              bool
	}{
		{
			name:              "Case 1: Test registry is secure",
			registryOperation: "add",
			registryName:      "secureRegistry",
			registryURL:       "https://github.com/test/secure-registry",
			forceFlag:         true,
			isSecure:          true,
			want:              true,
		},
		{
			name:              "Case 2: Test non-registry is secure",
			registryOperation: "add",
			registryName:      "nonSecureRegistry",
			registryURL:       "https://github.com/test/non-secure-registryy",
			forceFlag:         true,
			isSecure:          false,
			want:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := preference.New()
			if err != nil {
				t.Errorf("Unable to get preference file with error: %v", err)
			}
			err = cfg.RegistryHandler(tt.registryOperation, tt.registryName, tt.registryURL, tt.forceFlag, tt.isSecure)
			if err != nil {
				t.Errorf("Unable to add registry to preference file with error: %v", err)
			}

			got := IsSecure(tt.registryName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %t, want %t", got, tt.want)
			}
		})
	}
}
