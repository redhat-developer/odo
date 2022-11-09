package registry

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestIsSecure(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	tempConfigFileName := tempConfigFile.Name()

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
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{
				Globalodoconfig: &tempConfigFileName,
			})
			cfg, err := preference.NewClient(ctx)
			if err != nil {
				t.Errorf("Unable to get preference file with error: %v", err)
			}
			err = cfg.RegistryHandler(tt.registryOperation, tt.registryName, tt.registryURL, tt.forceFlag, tt.isSecure)
			if err != nil {
				t.Errorf("Unable to add registry to preference file with error: %v", err)
			}

			got := IsSecure(cfg, tt.registryName)
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
			registryURL: "https://registry.devfile.io",
			want:        false,
		},
		{
			name:        "Case 3: Returns false if URL git based on raw.githubusercontent",
			registryURL: "https://raw.githubusercontent.com/odo-devfiles/registry",
			want:        true,
		},
		{
			name:        "Case 4: Returns false if URL contains github.com in a non-host position",
			registryURL: "https://my.registry.example.com/github.com",
			want:        false,
		},
		{
			name:        "Case 5: Returns false if URL contains raw.githubusercontent.com in a non-host position",
			registryURL: "https://my.registry.example.com/raw.githubusercontent.com",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := IsGithubBasedRegistry(tt.registryURL)
			if err != nil {
				t.Errorf("unexpected error %s occoured", err.Error())
			}
			if actual != tt.want {
				t.Errorf("failed checking if registry is git based, got %t want %t", actual, tt.want)
			}
		})
	}
}
