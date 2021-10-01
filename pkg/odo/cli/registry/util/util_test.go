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

			got, err := IsSecure(tt.registryName)
			if err != nil {
				t.Errorf("Unable to check if the registry is secure or not")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %t, want %t", got, tt.want)
			}
		})
	}
}

func TestIsGitBasedRegistry(t *testing.T) {
	tests := []struct {
		name        string
		registryURL string
		want        bool
	}{
		{
			name:        "Case 1: Returns true if URL contains github",
			registryURL: "https://github.com/odo-devfiles/registry",
			want:        true,
		},
		{
			name:        "Case 2: Returns false if URL does not contain github",
			registryURL: " https://registry.devfile.io",
			want:        false,
		},
		{
			name:        "Case 3: Returns false if URL git based on raw.githubusercontent",
			registryURL: "https://raw.githubusercontent.com/odo-devfiles/registry",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actual := IsGitBasedRegistry(tt.registryURL); actual != tt.want {
				t.Errorf("failed checking if registry is git based, got %t want %t", actual, tt.want)
			}
		})
	}
}
