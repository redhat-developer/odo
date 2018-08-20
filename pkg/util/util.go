package util

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
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
func ExtractComponentType(namespacedVersionedComponentType string) string {
	/*
		Possible types of parameters:
		1. "myproject/python:3.5" -- Return python
		2. "python:3.5" -- Return python
		3. nodejs -- Return nodejs
	*/
	s := strings.Split(namespacedVersionedComponentType, "/")
	versionedString := s[0]
	if len(s) == 2 {
		versionedString = s[1]
	}
	s = strings.Split(versionedString, ":")
	return s[0]
}
