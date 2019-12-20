package helper

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/gomega"
)

type Config struct {
	ComponentSettings struct {
		Type           string   `yaml:"Type,omitempty"`
		SourceLocation string   `yaml:"SourceLocation,omitempty"`
		Ref            string   `yaml:"Ref,omitempty"`
		SourceType     string   `yaml:"SourceType,omitempty"`
		Ports          []string `yaml:"Ports,omitempty"`
		Application    string   `yaml:"Application,omitempty"`
		Project        string   `yaml:"Project,omitempty"`
		Name           string   `yaml:"Name,omitempty"`
		MinMemory      string   `yaml:"MinMemory,omitempty"`
		MaxMemory      string   `yaml:"MaxMemory,omitempty"`
		DebugPort      []int    `yaml:"DebugPort,omitempty"`
		Storage        []struct {
			Name string `yaml:"Name,omitempty"`
			Size string `yaml:"Size,omitempty"`
			Path string `yaml:"Path,omitempty"`
		} `yaml:"Storage,omitempty"`
		Ignore bool   `yaml:"Ignore,omitempty"`
		MinCPU string `yaml:"MinCPU,omitempty"`
		MaxCPU string `yaml:"MaxCPU,omitempty"`
		URL    []struct {
			// Name of the URL
			Name string `yaml:"Name,omitempty"`
			// Port number for the url of the component, required in case of components which expose more than one service port
			Port int `yaml:"Port,omitempty"`
		} `yaml:"Url,omitempty"`
	} `yaml:"ComponentSettings,omitempty"`
}

// VerifyLocalConfig verifies the content of the config.yaml file
func VerifyLocalConfig(context string) Config {
	var conf Config

	yamlFile, err := ioutil.ReadFile(context)
	if err != nil {
		fmt.Println(err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		fmt.Println(err)
	}
	return conf
}

// ValidateLocalCmpExist verifies the local config parameter
func ValidateLocalCmpExist(context, cmpType, cmpName, appName string) {
	cmpSetting := VerifyLocalConfig(filepath.Join(context, ".odo", "config.yaml"))
	Expect(cmpSetting.ComponentSettings.Type).To(ContainSubstring(cmpType))
	Expect(cmpSetting.ComponentSettings.Name).To(ContainSubstring(cmpName))
	Expect(cmpSetting.ComponentSettings.Application).To(ContainSubstring(appName))
}
