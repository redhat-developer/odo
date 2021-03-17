package helper

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/envinfo"
	"gopkg.in/yaml.v2"
)

const configFileDirectory = ".odo"
const configFileName = "config.yaml"

type config struct {
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
		Envs   []struct {
			Name  string `yaml:"Name,omitempty"`
			Value string `yaml:"Value,omitempty"`
		} `yaml:"Envs,omitempty"`
		URL []struct {
			// Name of the URL
			Name string `yaml:"Name,omitempty"`
			// Port number for the url of the component, required in case of components which expose more than one service port
			Port int `yaml:"Port,omitempty"`
		} `yaml:"Url,omitempty"`
	} `yaml:"ComponentSettings,omitempty"`
}

// VerifyLocalConfig verifies the content of the config.yaml file
func verifyLocalConfig(context string) (config, error) {
	var conf config

	yamlFile, err := ioutil.ReadFile(context)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// Search for the item in cmpfield string array
func Search(cmpField []string, val string) bool {
	for _, item := range cmpField {
		if item == val {
			return true
		}
	}
	return false
}

// newInterfaceValue takes interface and keyValue of args
// It returns new initialized value
func newInterfaceValue(cmpSetting *config, keyValue ...string) reflect.Value {
	indexNum, _ := strconv.Atoi(keyValue[1])
	if keyValue[0] == "URL" {
		return reflect.ValueOf(cmpSetting.ComponentSettings.URL[indexNum])
	}
	if keyValue[0] == "Storage" {
		return reflect.ValueOf(cmpSetting.ComponentSettings.Storage[indexNum])
	}
	return reflect.ValueOf(cmpSetting.ComponentSettings.Envs[indexNum])
}

// ValidateLocalCmpExist verifies the local config parameter
// It takes context and fieldType,value string as args
// URL and Storage parameter takes key,indexnumber,fieldType,value as args
func ValidateLocalCmpExist(context string, args ...string) {
	var interfaceVal reflect.Value
	cmpField := []string{"URL", "Storage", "Envs"}
	cmpSetting, err := verifyLocalConfig(filepath.Join(context, configFileDirectory, configFileName))
	if err != nil {
		Expect(err).To(Equal(nil))
	}

	for i := 0; i < len(args); i++ {
		keyValue := strings.Split(args[i], ",")

		// if any of the cmp type like 'URL' is interface and matches cmpField
		// New value is initialised for that particular interface in newInterfaceValue
		// else New value is initialised for ComponentSettings interface
		if Search(cmpField, keyValue[0]) {
			interfaceVal = newInterfaceValue(&cmpSetting, keyValue[0], keyValue[1])
			keyValue[0] = keyValue[2]
			keyValue[1] = keyValue[3]
		} else {
			interfaceVal = reflect.ValueOf(cmpSetting.ComponentSettings)
		}

		for i := 0; i < interfaceVal.NumField(); i++ {

			// Get the field, returns https://golang.org/pkg/reflect/#StructField
			field := interfaceVal.Field(i)
			typeField := interfaceVal.Type().Field(i)

			f := field.Interface()
			// Get the value of the field
			fieldVal := reflect.ValueOf(f)
			if typeField.Name == keyValue[0] {

				// validate the corresponding parameters of the type field
				// convert the field value into the string
				switch fieldVal.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					Expect(strconv.FormatInt(fieldVal.Int(), 10)).To(Equal(keyValue[1]))
				case reflect.String:
					Expect(fieldVal.String()).To(Equal(keyValue[1]))
				case reflect.Slice:
					sliceVal := fmt.Sprint(fieldVal)
					Expect(sliceVal).To(Equal(keyValue[1]))
				default:
					fmt.Println("Invalid Kind of the field value")

				}
			}

		}

	}

}

func LocalEnvInfo(context string) *envinfo.EnvSpecificInfo {
	info, err := envinfo.NewEnvSpecificInfo(filepath.Join(context, configFileDirectory, envInfoFile))
	if err != nil {
		Expect(err).To(Equal(nil))
	}
	return info
}
