package testingutil

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCustomHomeDir(t *testing.T) {
	tests := []struct {
		name    string
		wanterr bool
	}{
		{
			name:    "Test if specifying customDir results in appropriate resolution of config dir",
			wanterr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldCustomHomeDir := customHomeDir
			var err error
			customHomeDir, err = ioutil.TempDir("", "testconf")
			if err != nil {
				t.Error(err.Error())
			}
			odoConfigFile, kubeConfigFile, err := SetUp(
				ConfigDetails{
					FileName:      "odo-test-config",
					Config:        FakeOdoConfig("odo-test-config", false, ""),
					ConfigPathEnv: "GLOBALODOCONFIG",
				}, ConfigDetails{
					FileName:      "kube-test-config",
					Config:        FakeKubeClientConfig(),
					ConfigPathEnv: "KUBECONFIG",
				},
			)
			defer CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
			if err != nil {
				t.Errorf("failed to create mock odo and kube config files. Error %v", err)
			}
			customHomeDir = oldCustomHomeDir
		})
	}
}
