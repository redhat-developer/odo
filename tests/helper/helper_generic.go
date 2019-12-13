package helper

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

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
	Kind	string	`yaml:"Kind"`
	apiversion	string	`yaml:"apiversion"`
	ComponentSettings struct {
		Type	string	`yaml:"Type,omitempty"`
		SourceLocation	string `yaml:"SourceLocation"`
		SourceType	string `yaml:"SourceType"`
		Ports	*[]string `yaml:"Ports"`
		Application	string `yaml:"Application"`
		Project	string `yaml:"Project"`
		Name	string `yaml:"Name"`
		URL	struct {
			// Name of the URL
			Name string `yaml:"Name"`
			// Port number for the url of the component, required in case of components which expose more than one service port
			Port int `yaml:"Port"`
		} `yaml:"Url"`
	} `yaml:"ComponentSettings"`
}

// VerifyLocalConfig verifies the content of the config.yaml file
func VerifyLocalConfig(context string) Config {

	var conf Config

	yamlFile, err := ioutil.ReadFile(context)
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		fmt.Println(err)
	}
	return conf
}
