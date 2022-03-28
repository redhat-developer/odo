package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"gopkg.in/yaml.v2"
)

// CreateIfNotExists creates the directory and the file if it doesn't exist
func CreateIfNotExists(configFile string) error {
	_, err := os.Stat(filepath.Dir(configFile))
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(configFile), 0750)
		if err != nil {
			return fmt.Errorf("unable to create directory: %w", err)
		}
	}
	// Check whether config file is present or not
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		file, err := os.Create(configFile)
		if err != nil {
			return fmt.Errorf("unable to create config file: %w", err)
		}
		defer file.Close() // #nosec G307
	}

	return nil
}

// GetFromFile unmarshals a struct from a odo config file
func GetFromFile(c interface{}, filename string) error {
	configData, err := filesystem.Get().ReadFile(filename)
	if err != nil {
		return fmt.Errorf("unable to read file %v: %w", filename, err)
	}

	err = yaml.Unmarshal(configData, c)
	if err != nil {
		return fmt.Errorf("unable to unmarshal odo config file: %w", err)
	}

	return nil
}

// WriteToYAMLFile marshals a struct to a file
func WriteToYAMLFile(c interface{}, filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to marshal odo config data: %w", err)
	}

	if err = CreateIfNotExists(filename); err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		return fmt.Errorf("unable to write config to file %v: %w", c, err)
	}

	return nil
}

// IsSet uses reflection to check if a parameter is set in a struct
// using the name in a case insensitive manner
// only supports flat structs
// TODO: support deeper struct using recursion
func IsSet(info interface{}, parameter string) bool {
	imm := reflect.ValueOf(info)
	if imm.Kind() == reflect.Ptr {
		imm = imm.Elem()
	}
	val := imm.FieldByNameFunc(CaseInsensitive(parameter))

	// If the value kind is string, only checks if the value is valid
	// since by definition argument must be a chan, func, interface, map, pointer, or slice
	// for IsNil() function of reflect package
	if val.Kind() == reflect.String && val.IsValid() {
		return true
	}

	if !val.IsValid() || val.IsNil() {
		return false
	}
	if val.IsNil() {
		return false
	}
	// if the value is a Ptr then we need to de-ref it
	if val.Kind() == reflect.Ptr {
		return true
	}

	return true
}

// CaseInsensitive returns a function which compares two words
// caseinsensitively
func CaseInsensitive(parameter string) func(word string) bool {
	return func(word string) bool {
		return strings.EqualFold(word, parameter)
	}
}

// DeleteConfiguration sets a parameter to null in a struct using reflection
func DeleteConfiguration(info interface{}, parameter string) error {

	imm := reflect.ValueOf(info)
	if imm.Kind() == reflect.Ptr {
		imm = imm.Elem()
	}
	val := imm.FieldByNameFunc(CaseInsensitive(parameter))
	if !val.IsValid() {
		return fmt.Errorf("unknown parameter :'%s' is not a parameter in odo config", parameter)
	}

	if val.CanSet() {
		val.Set(reflect.Zero(val.Type()))
		return nil
	}
	return fmt.Errorf("cannot set %s to nil", parameter)

}

// GetLowerCaseParameters creates a set-like map of supported parameters from the supported parameter names
func GetLowerCaseParameters(parameters []string) map[string]bool {
	result := make(map[string]bool, len(parameters))
	for _, v := range parameters {
		result[strings.ToLower(v)] = true
	}
	return result
}
