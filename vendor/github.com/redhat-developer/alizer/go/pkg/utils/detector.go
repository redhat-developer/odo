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
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/schema"
	ignore "github.com/sabhiram/go-gitignore"
)

const FROM_PORT = 0
const TO_PORT = 65535

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
	for _, plugin := range pom.Build.Plugins.Plugin {
		if strings.Contains(plugin.GroupId, tag) {
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
	packageJson, err := GetPackageJsonSchemaFromFile(file)
	if err != nil {
		return false
	}
	if packageJson.Dependencies != nil {
		for dependency := range packageJson.Dependencies {
			if strings.Contains(dependency, tag) {
				return true
			}
		}
	}
	if packageJson.PeerDependencies != nil {
		for dependency := range packageJson.PeerDependencies {
			if strings.Contains(dependency, tag) {
				return true
			}
		}
	}
	return false
}

func GetPackageJsonSchemaFromFile(path string) (schema.PackageJson, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return schema.PackageJson{}, err
	}

	var packageJson schema.PackageJson
	json.Unmarshal(bytes, &packageJson)
	return packageJson, nil
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

func GetFilePathsFromRoot(root string) ([]string, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	var files []string
	ignoreFile, errorIgnoreFile := getIgnoreFile(root)
	errWalk := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			relativePath := strings.Replace(path, root, "", 1)
			if errorIgnoreFile == nil && ignoreFile.MatchesPath(relativePath) {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
			if !info.IsDir() && isFileInRoot(root, path) {
				files = append([]string{path}, files...)
			} else {
				files = append(files, path)
			}
			return nil
		})
	return files, errWalk
}

func getIgnoreFile(root string) (*ignore.GitIgnore, error) {
	gitIgnorePath := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(gitIgnorePath); err == nil {
		return ignore.CompileIgnoreFile(gitIgnorePath)
	}
	return nil, errors.New("no git ignore file found")
}

func isFileInRoot(root string, file string) bool {
	dir, _ := filepath.Split(file)
	return strings.EqualFold(filepath.Clean(dir), filepath.Clean(root))
}

func GetFilePathsInRoot(root string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, fileInfo := range fileInfos {
		files = append(files, filepath.Join(root, fileInfo.Name()))
	}
	return files, nil
}

func ConvertPropertiesFileAsPathToMap(path string) (map[string]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ConvertPropertiesFileToMap(bytes)
}

func ConvertPropertiesFileToMap(fileInBytes []byte) (map[string]string, error) {
	config := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(fileInBytes))
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func GetValidPortsFromEnvs(envs []string) []int {
	validPorts := []int{}
	for _, env := range envs {
		envValue := os.Getenv(env)
		if port, err := GetValidPort(envValue); err == nil {
			validPorts = append(validPorts, port)
		}
	}
	return validPorts
}

func GetValidPorts(ports []string) []int {
	validPorts := []int{}
	for _, portValue := range ports {
		if port, err := GetValidPort(portValue); err == nil {
			validPorts = append(validPorts, port)
		}
	}
	return validPorts
}

func GetValidPort(port string) (int, error) {
	if port, err := strconv.Atoi(port); err == nil && IsValidPort(port) {
		return port, nil
	}
	return -1, errors.New("no valid port found")
}

func IsValidPort(port int) bool {
	return port > FROM_PORT && port < TO_PORT
}

func GetAnyApplicationFilePath(root string, propsFiles []model.ApplicationFileInfo) string {
	files, err := GetFilePathsFromRoot(root)
	if err != nil {
		return ""
	}
	for _, path := range files {
		dir, file := filepath.Split(path)
		for _, propsFile := range propsFiles {
			if match, _ := regexp.MatchString(propsFile.File, file); match && strings.Contains(dir, propsFile.Dir) {
				return path
			}

		}
	}
	return ""
}

func ReadAnyApplicationFile(root string, propsFiles []model.ApplicationFileInfo) ([]byte, error) {
	path := GetAnyApplicationFilePath(root, propsFiles)
	if path != "" {
		return ioutil.ReadFile(path)
	}
	return nil, errors.New("no file found")
}

func FindPortSubmatch(re *regexp.Regexp, text string, group int) int {
	if text != "" {
		matches := re.FindStringSubmatch(text)
		if len(matches) > group {
			if port, err := GetValidPort(matches[group]); err == nil {
				return port
			}
		}
	}
	return -1
}

func FindAllPortsSubmatch(re *regexp.Regexp, text string, group int) []int {
	ports := []int{}
	if text != "" {
		matchIndexesSlice := re.FindAllStringSubmatch(text, -1)
		for _, matches := range matchIndexesSlice {
			if len(matches) > group {
				portValue := matches[group]
				if port, err := GetValidPort(portValue); err == nil {
					ports = append(ports, port)
				}
			}
		}
	}
	return ports
}

func GetPortValueFromEnvFile(root string, regex string) int {
	ports := GetPortValuesFromEnvFile(root, []string{regex})
	if len(ports) > 0 {
		return ports[0]
	}
	return -1
}

func GetPortValuesFromEnvFile(root string, regexes []string) []int {
	ports := []int{}
	text, err := getEnvFileContent(root)
	if err != nil {
		return ports
	}
	for _, regex := range regexes {
		re := regexp.MustCompile(regex)
		port := FindPortSubmatch(re, text, 1)
		if port != -1 {
			ports = append(ports, port)
		}
	}
	return ports
}

func GetStringValueFromEnvFile(root string, regex string) string {
	text, err := getEnvFileContent(root)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(regex)
	if text != "" {
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func getEnvFileContent(root string) (string, error) {
	envPath := filepath.Join(root, ".env")
	bytes, err := os.ReadFile(envPath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
