package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// RandString returns a random string of given length
func RandString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// WaitForCmdOut runs a command until it gets
// the expected output.
// It accepts 5 arguments, program (program to be run)
// args (arguments to the program)
// timeout (the time to wait for the output)
// errOnFail (flag to set if test should fail if command fails)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func WaitForCmdOut(program string, args []string, timeout int, errOnFail bool, check func(output string) bool) bool {
	pingTimeout := time.After(time.Duration(timeout) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout out after %v minutes", timeout))

		case <-tick:
			stdOut := CmdShouldPass(program, args...)
			if check(strings.TrimSpace(string(stdOut))) {
				return true
			}
		}
	}
}

// MatchAllInOutput ensures all strings are in output
func MatchAllInOutput(output string, tomatch []string) {
	for _, i := range tomatch {
		Expect(output).To(ContainSubstring(i))
	}
}

// DontMatchAllInOutput ensures all strings are not in output
func DontMatchAllInOutput(output string, tonotmatch []string) {
	for _, i := range tonotmatch {
		Expect(output).ToNot(ContainSubstring(i))
	}
}

// Unindented returns the unindented version of the jsonStr passed to it
func Unindented(jsonStr string) (string, error) {
	var tmpMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &tmpMap)
	if err != nil {
		return "", err
	}

	obj, err := json.Marshal(tmpMap)
	if err != nil {
		return "", err
	}
	return string(obj), err
}

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
	cmpSetting := VerifyLocalConfig(context + "/.odo/config.yaml")
	Expect(cmpSetting.ComponentSettings.Type).To(ContainSubstring(cmpType))
	Expect(cmpSetting.ComponentSettings.Name).To(ContainSubstring(cmpName))
	Expect(cmpSetting.ComponentSettings.Application).To(ContainSubstring(appName))
}
