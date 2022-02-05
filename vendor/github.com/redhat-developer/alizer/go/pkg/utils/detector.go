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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/schema"
)

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
	return file == wantedFile
}

func IsTagInFile(file string, tag string) bool {
	contentInByte, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	content := string(contentInByte)
	return strings.Contains(content, tag)
}

func IsTagInPomXMLFile(file string, tag string) bool {
	xmlFile, err := os.Open(file)
	if err != nil {
		return false
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var pom schema.Pom
	xml.Unmarshal(byteValue, &pom)

	defer xmlFile.Close()
	for _, dependency := range pom.Dependencies.Dependency {
		if strings.Contains(dependency.GroupId, tag) {
			return true
		}
	}
	return false
}

func IsTagInPackageJsonFile(file string, tag string) bool {
	jsonFile, err := os.Open(file)
	if err != nil {
		return false
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var packageJson schema.PackageJson
	json.Unmarshal(byteValue, &packageJson)

	defer jsonFile.Close()
	if packageJson.Dependencies != nil {
		for dependency := range packageJson.Dependencies {
			if strings.Contains(dependency, tag) {
				return true
			}
		}
	}
	return false
}

func AddToArrayIfValueExist(arr *[]string, val string) {
	if val != "" {
		*arr = append(*arr, val)
	}
}
