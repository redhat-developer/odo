package util

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"net/url"
	"strings"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
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
// existList: List to verify that the returned name does not already exist
// preferredSuffix: Optional suffix if passed, will be checked if prefix-suffix does not exist in existList if not, will add an additional timestamp to make name unique
// Returns randomname is prefix-suffix, if suffix is passed else generated. Aditionally, if the prefix-suffix is used, it'll be appended with additional 4 char string to take the form prefix-suffix-{a-z A-z}4+
func GetRandomName(prefix string, existList []string, preferredSuffix string) string {
	if preferredSuffix == "" {
		// Generate suffix if not passed, using random country names generated from Pallinder/go-randomdata as suffix
		preferredSuffix = strings.Replace(
			strings.Replace(strings.Replace(
				strings.Replace(
					strings.Replace(
						strings.ToLower(randomdata.Country(randomdata.FullCountry)),
						" ",
						"-",
						-1,
					),
					".",
					"-",
					-1,
				),
				",",
				"-",
				-1,
			),
				"(",
				"-",
				-1,
			),
			")",
			"-",
			-1,
		)
	}
	// name is prefix-suffix
	name := fmt.Sprintf("%s-%s", prefix, preferredSuffix)
	// check if generated name is already used in the existList
	if StringInSlice(name, existList) {
		prevName := name
		// keep generating names until generated name is not unique. So, loop terminates when name is unique and hence for condition is false
		for StringInSlice(prevName, existList) {
			prevName = name
			// Attempt unique name generation from prefix-suffix by concatenating prefix-suffix withrandom string of length 4
			prevName = fmt.Sprintf("%s-%s", prevName, GenerateRandomString(4))
		}
		// If found to be unique, set name as generated name
		name = prevName
	}
	// return name
	return name
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
}
