package util

import (
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

// ConvertLabelsToSelector converts the given labels to selector
func ConvertLabelsToSelector(labels map[string]string) string {
	var selector string
	isFirst := true
	for k, v := range labels {
		if isFirst {
			isFirst = false
			if v == "" {
				selector = selector + fmt.Sprintf("%v", k)
			} else {
				selector = fmt.Sprintf("%v=%v", k, v)
			}
		} else {
			if v == "" {
				selector = selector + fmt.Sprintf(",%v", k)
			} else {
				selector = selector + fmt.Sprintf(",%v=%v", k, v)
			}
		}
	}
	return selector
}

// GenerateRandomString generates a random string of lower case characters of
// the given size
func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// TruncateString truncates passed string to given length
// Note: if -1 is passed, the original string is returned
func TruncateString(str string, maxLen int) string {
	if maxLen == -1 {
		return str
	}
	if len(str) > maxLen {
		return str[:maxLen]
	}
	return str
}

// StringInSlice returns if a passed string exists in the passed slice of strings
func StringInSlice(checkStr string, strSlice []string) bool {
	// Iterate the slice
	for _, ele := range strSlice {
		// if current element is same as element to search(checkStr), return true indicating checkStr found
		if ele == checkStr {
			return true
		}
	}
	// Exited loop without finding required element(checkStr) in slice, so return false indicating checkStr not found in slice
	return false
}

// GetRandomName returns a randomly generated name which can be used for naming odo and/or openshift entities
// prefix: Desired prefix part of the name
// prefixMaxLen: Desired maximum length of prefix part of random name; if -1 is passed, no limit on length will be enforced
// existList: List to verify that the returned name does not already exist
// retries: number of retries to try generating a unique name
// Returns:
//		1. randomname: is prefix-suffix, where:
//				prefix: string passed as prefix or fetched current directory of length same as the passed prefixMaxLen
//				suffix: 4 char random string
//      2. error: if requested number of retries also failed to generate unique name
func GetRandomName(prefix string, prefixMaxLen int, existList []string, retries int) (string, error) {
	prefix = TruncateString(GetDNS1123Name(strings.ToLower(prefix)), prefixMaxLen)
	name := fmt.Sprintf("%s-%s", prefix, GenerateRandomString(4))

	//Create a map of existing names for efficient iteration to find if the newly generated name is same as any of the already existing ones
	existingNames := make(map[string]bool)
	for _, existingName := range existList {
		existingNames[existingName] = true
	}

	// check if generated name is already used in the existList
	if _, ok := existingNames[name]; ok {
		prevName := name
		trial := 0
		// keep generating names until generated name is not unique. So, loop terminates when name is unique and hence for condition is false
		for ok {
			trial = trial + 1
			prevName = name
			// Attempt unique name generation from prefix-suffix by concatenating prefix-suffix withrandom string of length 4
			prevName = fmt.Sprintf("%s-%s", prevName, GenerateRandomString(4))
			_, ok = existingNames[prevName]
			if trial >= retries {
				// Avoid infinite loops and fail after passed number of retries
				return "", fmt.Errorf("failed to generate a unique name even after %d retrials", retries)
			}
		}
		// If found to be unique, set name as generated name
		name = prevName
	}
	// return name
	return name, nil
}

// ComponentCreateType is an enum to indicate the type of source of component -- local source/binary or git for the generation of app/component names
type ComponentCreateType string

const (
	// GIT as source of component
	GIT ComponentCreateType = "git"
	// SOURCE Local source path as source of component
	SOURCE ComponentCreateType = "source"
	// BINARY Local Binary as source of component
	BINARY ComponentCreateType = "binary"
	// NONE indicates there's no information about the type of source of the component
	NONE ComponentCreateType = ""
)

// GetComponentDir returns source repo name
// Parameters:
//		path: git url or source path or binary path
//		paramType: One of ComponentCreateType as in GIT/SOURCE/BINARY
// Returns: directory name
func GetComponentDir(path string, paramType ComponentCreateType) (string, error) {
	retVal := ""
	switch paramType {
	case GIT:
		retVal = strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".git")
	case SOURCE:
		retVal = filepath.Base(path)
	case BINARY:
		filename := filepath.Base(path)
		var extension = filepath.Ext(filename)
		retVal = filename[0 : len(filename)-len(extension)]
	default:
		currDir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrapf(err, "unable to generate a random name as getting current directory failed")
		}
		retVal = filepath.Base(currDir)
	}
	retVal = strings.TrimSpace(GetDNS1123Name(strings.ToLower(retVal)))
	return retVal, nil
}

// Hyphenate applicationName and componentName
func NamespaceOpenShiftObject(componentName string, applicationName string) (string, error) {

	// Error if it's blank
	if componentName == "" {
		return "", errors.New("namespacing: component name cannot be blank")
	}

	// Error if it's blank
	if applicationName == "" {
		return "", errors.New("namespacing: application name cannot be blank")
	}

	// Return the hyphenated namespaced name
	return fmt.Sprintf("%s-%s", strings.Replace(componentName, "/", "-", -1), applicationName), nil
}

// ExtractComponentType returns only component type part from passed component type(default unqualified, fully qualified, versioned, etc...and their combinations) for use as component name
// Possible types of parameters:
// 1. "myproject/python:3.5" -- Return python
// 2. "python:3.5" -- Return python
// 3. nodejs -- Return nodejs
func ExtractComponentType(namespacedVersionedComponentType string) string {
	s := strings.Split(namespacedVersionedComponentType, "/")
	versionedString := s[0]
	if len(s) == 2 {
		versionedString = s[1]
	}
	s = strings.Split(versionedString, ":")
	return s[0]
}

// parseCreateCmdArgs returns
// 1. image name
// 2. component type i.e, builder image name
// 3. component name default value is component type else the user requested component name
// 4. component version which is by default latest else version passed with builder image name
func ParseCreateCmdArgs(args []string) (string, string, string, string) {
	// We don't have to check it anymore, Args check made sure that args has at least one item
	// and no more than two

	// "Default" values
	componentImageName := args[0]
	componentType := args[0]
	componentName := ExtractComponentType(componentType)
	componentVersion := "latest"

	// Check if componentType includes ":", if so, then we need to spit it into using versions
	if strings.ContainsAny(componentImageName, ":") {
		versionSplit := strings.Split(args[0], ":")
		componentType = versionSplit[0]
		componentName = ExtractComponentType(componentType)
		componentVersion = versionSplit[1]
	}
	return componentImageName, componentType, componentName, componentVersion
}

const WIN = "windows"

// Reads file path form URL file:///C:/path/to/file to C:\path\to\file
func ReadFilePath(u *url.URL, os string) string {
	location := u.Path
	if os == WIN {
		location = strings.Replace(u.Path, "/", "\\", -1)
		location = location[1:]
	}
	return location
}

// Converts file path on windows to /C:/path/to/file to work in URL
func GenFileUrl(location string, os string) string {
	urlPath := location
	if os == WIN {
		urlPath = "/" + strings.Replace(location, "\\", "/", -1)
	}
	return "file://" + urlPath
}

<<<<<<< HEAD
// ConvertKeyValueStringToMap converts String Slice of Parameters to a Map[String]string
// Each value of the slice is expected to be in the key=value format
// Values that do not conform to this "spec", will be ignored
func ConvertKeyValueStringToMap(params []string) map[string]string {
	result := make(map[string]string, len(params))
	for _, param := range params {
		str := strings.Split(param, "=")
		if len(str) != 2 {
			glog.Fatalf("Parameter %s is not in the expected key=value format", param)
		} else {
			result[str[0]] = str[1]
		}
	}
	return result
=======
// GetDNS1123Name Converts passed string into DNS-1123 string
func GetDNS1123Name(str string) string {
	replacer := strings.NewReplacer(
		" ", "-",
		".", "-",
		",", "-",
		"(", "-",
		")", "-",
		"/", "-",
		":", "-",
		"--", "-",
	)
	return strings.TrimSpace(replacer.Replace(strings.ToLower(str)))
>>>>>>> Fix travis failures
}
