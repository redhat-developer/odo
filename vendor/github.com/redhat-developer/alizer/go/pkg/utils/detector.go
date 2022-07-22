/*******************************************************************************
 * Copyright (c) 2021 Red Hat, Inc.
 * Distributed under license by Red Hat, Inc. All rights reserved.
 * This program is made available under the terms of the
 * Eclipse Public License v2.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 * Red Hat, Inc.
 ******************************************************************************/
package utils

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/schema"
)

func GetFilesByRegex(filePaths *[]string, regexFile string) []string {
	matchedPaths := []string{}
	for _, path := range *filePaths {
		if isPathOfWantedRegex(path, regexFile) {
			matchedPaths = append(matchedPaths, path)
		}
	}
	return matchedPaths
}

func isPathOfWantedRegex(path string, regexFile string) bool {
	_, file := filepath.Split(path)
	matched, _ := regexp.MatchString(regexFile, file)
	return matched
}

func GetFile(filePaths *[]string, wantedFile string) string {
	for _, path := range *filePaths {
		if IsPathOfWantedFile(path, wantedFile) {
			return path
		}
	}
	return ""
}

func HasFile(files *[]string, wantedFile string) bool {
	for _, path := range *files {
		if IsPathOfWantedFile(path, wantedFile) {
			return true
		}
	}
	return false
}

func IsPathOfWantedFile(path string, wantedFile string) bool {
	_, file := filepath.Split(path)
	return strings.EqualFold(file, wantedFile)
}

func IsTagInFile(file string, tag string) (bool, error) {
	contentInByte, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err
	}
	content := string(contentInByte)
	return strings.Contains(content, tag), nil
}

func IsTagInPomXMLFile(pomFilePath string, tag string) (bool, error) {
	pom, err := GetPomFileContent(pomFilePath)
	if err != nil {
		return false, err
	}
	for _, dependency := range pom.Dependencies.Dependency {
		if strings.Contains(dependency.GroupId, tag) {
			return true, nil
		}
	}
	return false, nil
}

func GetPomFileContent(pomFilePath string) (schema.Pom, error) {
	xmlFile, err := os.Open(pomFilePath)
	if err != nil {
		return schema.Pom{}, err
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var pom schema.Pom
	xml.Unmarshal(byteValue, &pom)

	defer xmlFile.Close()
	return pom, nil
}

func IsTagInPackageJsonFile(file string, tag string) bool {
	packageJson, err := GetPackageJsonFile(file)
	if err == nil && packageJson.Dependencies != nil {
		for dependency := range packageJson.Dependencies {
			if strings.Contains(dependency, tag) {
				return true
			}
		}
	}
	return false
}

func GetPackageJsonFile(file string) (schema.PackageJson, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return schema.PackageJson{}, errors.New("error opening file")
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var packageJson schema.PackageJson
	err = json.Unmarshal(byteValue, &packageJson)
	defer jsonFile.Close()
	return packageJson, err
}

func AddToArrayIfValueExist(arr *[]string, val string) {
	if val != "" {
		*arr = append(*arr, val)
	}
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
