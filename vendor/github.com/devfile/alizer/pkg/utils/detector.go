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
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/schema"
	"github.com/devfile/alizer/pkg/utils/langfiles"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	ignore "github.com/sabhiram/go-gitignore"
)

const FROM_PORT = 0
const TO_PORT = 65535
const FRAMEWORK_WEIGHT = 10
const TOOL_WEIGHT = 5

// GetFilesByRegex returns a slice of file paths from filePaths if the file name matches the regex.
func GetFilesByRegex(filePaths *[]string, regexFile string) []string {
	var matchedPaths []string
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

// GetFile returns the first match where the wantedFile is in a filePaths path.
func GetFile(filePaths *[]string, wantedFile string) string {
	for _, path := range *filePaths {
		if IsPathOfWantedFile(path, wantedFile) {
			return path
		}
	}
	return ""
}

// IsPathOfWantedFile checks if the file is in the path.
func IsPathOfWantedFile(path string, wantedFile string) bool {
	_, file := filepath.Split(path)
	return strings.EqualFold(file, wantedFile)
}

// IsTagInFile checks if the file contains the tag.
func IsTagInFile(file string, tag string) (bool, error) {
	contentInByte, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return false, err
	}
	content := string(contentInByte)
	return strings.Contains(content, tag), nil
}

// IsTagInPomXMLFileArtifactId checks if a pom file contains the artifactId.
func IsTagInPomXMLFileArtifactId(pomFilePath, groupId, artifactId string) (bool, error) {
	pom, err := GetPomFileContent(pomFilePath)
	if err != nil {
		return false, err
	}
	for _, dependency := range pom.Dependencies.Dependency {
		if strings.Contains(dependency.ArtifactId, artifactId) && strings.Contains(dependency.GroupId, groupId) {
			return true, nil
		}
	}
	for _, plugin := range pom.Build.Plugins.Plugin {
		if strings.Contains(plugin.ArtifactId, artifactId) && strings.Contains(plugin.GroupId, groupId) {
			return true, nil
		}
	}
	for _, profile := range pom.Profiles.Profile {
		for _, plugin := range profile.Build.Plugins.Plugin {
			if strings.Contains(plugin.ArtifactId, artifactId) && strings.Contains(plugin.GroupId, groupId) {
				return true, nil
			}
		}
	}
	return false, nil
}

// IsTagInPomXMLFile checks if a pom file contains the tag.
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

// GetPomFileContent returns the pom found in the path.
func GetPomFileContent(pomFilePath string) (schema.Pom, error) {
	cleanPomFilePath := filepath.Clean(pomFilePath)
	xmlFile, err := os.Open(cleanPomFilePath)
	if err != nil {
		return schema.Pom{}, err
	}
	byteValue, _ := io.ReadAll(xmlFile)

	var pom schema.Pom
	err = xml.Unmarshal(byteValue, &pom)
	if err != nil {
		return schema.Pom{}, err
	}
	defer func() error {
		if err := xmlFile.Close(); err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
		return nil
	}()
	return pom, nil
}

// IsTagInPackageJsonFile checks if the file is a package.json and contains the tag.
func IsTagInPackageJsonFile(file string, tag string) bool {
	packageJson, err := GetPackageJsonSchemaFromFile(file)
	if err != nil {
		return false
	}

	hasDependency := isTagInDependencies(packageJson.Dependencies, tag)
	if !hasDependency {
		hasDependency = isTagInDependencies(packageJson.DevDependencies, tag)
	}
	if !hasDependency {
		hasDependency = isTagInDependencies(packageJson.PeerDependencies, tag)
	}
	return hasDependency
}

func isTagInDependencies(deps map[string]string, tag string) bool {
	for dependency := range deps {
		if strings.Contains(dependency, tag) {
			return true
		}
	}
	return false
}

// GetPackageJsonSchemaFromFile returns the package.json found in the path.
func GetPackageJsonSchemaFromFile(path string) (schema.PackageJson, error) {
	cleanPath := filepath.Clean(path)
	bytes, err := os.ReadFile(cleanPath)
	if err != nil {
		return schema.PackageJson{}, err
	}

	var packageJson schema.PackageJson
	err = json.Unmarshal(bytes, &packageJson)
	if err != nil {
		return schema.PackageJson{}, err
	}
	return packageJson, nil
}

// IsTagInComposerJsonFile checks if the file is a composer.json and contains the tag.
func IsTagInComposerJsonFile(file string, tag string) bool {
	composerJson, err := GetComposerJsonSchemaFromFile(file)
	if err != nil {
		return false
	}

	hasDependency := isTagInDependencies(composerJson.Require, tag)
	if !hasDependency {
		hasDependency = isTagInDependencies(composerJson.RequireDev, tag)
	}
	return hasDependency
}

// GetComposerJsonSchemaFromFile returns the composer.json found in the path.
func GetComposerJsonSchemaFromFile(path string) (schema.ComposerJson, error) {
	cleanPath := filepath.Clean(path)
	bytes, err := os.ReadFile(cleanPath)
	if err != nil {
		return schema.ComposerJson{}, err
	}

	var composerJson schema.ComposerJson
	err = json.Unmarshal(bytes, &composerJson)
	if err != nil {
		return schema.ComposerJson{}, err
	}
	return composerJson, nil
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

// GetFilePathsFromRoot walks the file tree starting from root and returns a slice of all file paths found.
// Ignores files from .gitignore if it exists.
func GetFilePathsFromRoot(root string) ([]string, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	var files []string
	ignoreFile, errorIgnoreFile := getIgnoreFile(root)
	excludedFolders := langfiles.Get().GetExcludedFolders()
	errWalk := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			relativePath := strings.Replace(path, root, "", 1)
			// skip directories from excluded folders
			for _, excludedFolder := range excludedFolders {
				if strings.Contains(relativePath, excludedFolder) {
					return filepath.SkipDir
				}
			}
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

// GetFilePathsInRoot returns a slice of all files in the root.
func GetFilePathsInRoot(root string) ([]string, error) {
	fileInfos, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, fileInfo := range fileInfos {
		files = append(files, filepath.Join(root, fileInfo.Name()))
	}
	return files, nil
}

// ConvertPropertiesFileAsPathToMap fetches a file from a given path and transforms it into a map
func ConvertPropertiesFileAsPathToMap(path string) (map[string]string, error) {
	bytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return ConvertPropertiesFileToMap(bytes)
}

// ConvertPropertiesFileAsPathToMap transforms a slice of bytes it into a map
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

// GetValidPortsFromEnvs returns a slice of valid ports.
func GetValidPortsFromEnvs(envs []string) []int {
	var validPorts []int
	for _, env := range envs {
		envValue := os.Getenv(env)
		if port, err := GetValidPort(envValue); err == nil {
			validPorts = append(validPorts, port)
		}
	}
	return validPorts
}

// GetEnvVarsFromDockerFile returns a slice of env vars from Dockerfiles in the given directory.
func GetEnvVarsFromDockerFile(root string) ([]model.EnvVar, error) {
	locations := GetLocations(root)
	for _, location := range locations {
		filePath := filepath.Join(root, location)
		cleanFilePath := filepath.Clean(filePath)
		file, err := os.Open(cleanFilePath)
		if err == nil {
			defer func() error {
				if err := file.Close(); err != nil {
					return fmt.Errorf("error closing file: %s", err)
				}
				return nil
			}()
			return readEnvVarsFromDockerfile(file)
		}
	}
	return nil, fmt.Errorf("no dockefile found inside dir: %s", root)
}

// GetValidPortsFromEnvs returns a slice of valid ports from a dockerfile.
func GetValidPortsFromEnvDockerfile(envs []string, envVars []model.EnvVar) []int {
	var validPorts []int
	for _, env := range envs {
		for _, envVar := range envVars {
			if envVar.Name == env {
				if port, err := GetValidPort(envVar.Value); err == nil {
					validPorts = append(validPorts, port)
				}
				break
			}
		}
	}
	return validPorts
}

// GetLocations returns a list of file paths representing common locations where Dockerfiles might be found
// within the specified 'root' directory and one level down.
//
// It starts with a predefined list of common file names for a Dockerfile
// ('Dockerfile', 'Containerfile', 'dockerfile', 'containerfile'), and appends such file names to the 'root' subdirectories.
//
// Note that hidden files and directories (starting with a dot, e.g., '.git') are ignored while traversing the 'root' directory.
func GetLocations(root string) []string {
	filenames := []string{"Dockerfile", "Containerfile", "dockerfile", "containerfile"}
	locations := make([]string, len(filenames))

	copy(locations, filenames)

	entries, err := os.ReadDir(root)
	if err != nil {
		return locations
	}
	dirItems := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return locations
		}
		dirItems = append(dirItems, info)
	}
	for _, item := range dirItems {
		if strings.HasPrefix(item.Name(), ".") {
			continue
		}
		tmpPath := filepath.Join(root, item.Name())
		fileInfo, err := os.Stat(tmpPath)
		if err != nil {
			continue
		}
		if fileInfo.IsDir() {
			for _, filename := range filenames {
				locations = append(locations, filepath.Join(item.Name(), filename))
			}
		}
	}
	return locations
}

// ReadPortsFromDockerfile returns a slice of port numbers.
func ReadPortsFromDockerfile(file io.Reader) []int {
	var ports []int
	res, err := parser.Parse(file)
	if err != nil {
		return ports
	}

	for _, child := range res.AST.Children {
		// check for the potential port number in a Dockerfile/Containerfile
		if strings.ToLower(child.Value) == "expose" {
			for n := child.Next; n != nil; n = n.Next {
				if port, err := strconv.Atoi(n.Value); err == nil {
					ports = append(ports, port)
				}

			}
		}
	}
	return ports
}

func upsertEnvVar(envVars []model.EnvVar, envVar model.EnvVar) []model.EnvVar {
	isPresent := false
	for i := range envVars {
		if envVars[i].Name == envVar.Name {
			isPresent = true
			envVars[i].Value = envVar.Value
		}
	}
	if !isPresent {
		envVars = append(envVars, envVar)
	}
	return envVars
}

// readEnvVarsFromDockerfile returns a slice of envVars.
func readEnvVarsFromDockerfile(file io.Reader) ([]model.EnvVar, error) {
	var envVars []model.EnvVar
	res, err := parser.Parse(file)
	if err != nil {
		return envVars, err
	}

	for _, child := range res.AST.Children {
		// check for the potential env var in a Dockerfile/Containerfile
		if strings.ToLower(child.Value) != "env" {
			continue
		}
		firstNode := child.Next
		var secondNode *parser.Node
		if firstNode == nil {
			continue
		}
		secondNode = firstNode.Next
		if secondNode == nil {
			continue
		}
		envVar := model.EnvVar{
			Name:  firstNode.Value,
			Value: secondNode.Value,
		}
		envVars = upsertEnvVar(envVars, envVar)
	}

	return envVars, nil
}

// GetEnvVarPortValueFromDockerfile gets port value defined as env vars.
func GetEnvVarPortValueFromDockerfile(path string, portPlaceholders []string) ([]int, error) {
	envVars, err := GetEnvVarsFromDockerFile(path)
	ports := []int{}
	if err != nil {
		return ports, err
	}
	for _, envVar := range envVars {
		for _, portPlaceholder := range portPlaceholders {
			if envVar.Name != portPlaceholder {
				continue
			}
			if port, err := GetValidPort(envVar.Value); err == nil {
				ports = append(ports, port)
			}
		}
	}
	return ports, nil
}

// GetValidPorts returns a slice of valid ports.
func GetValidPorts(ports []string) []int {
	var validPorts []int
	for _, portValue := range ports {
		if port, err := GetValidPort(portValue); err == nil {
			validPorts = append(validPorts, port)
		}
	}
	return validPorts
}

// GetValidPort checks if a string is a valid port and returns the port.
// Returns -1 if not a valid port.
func GetValidPort(port string) (int, error) {
	if port, err := strconv.Atoi(port); err == nil && IsValidPort(port) {
		return port, nil
	}
	return -1, errors.New("no valid port found")
}

// IsValidPort checks if an integer is a valid port.
func IsValidPort(port int) bool {
	return port > FROM_PORT && port < TO_PORT
}

// GetAnyApplicationFilePath returns the location of a file if it exists in the directory and the given file name is a substring.
func GetAnyApplicationFilePath(root string, propsFiles []model.ApplicationFileInfo, ctx *context.Context) string {
	files, err := GetCachedFilePathsFromRoot(root, ctx)
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

// GetAnyApplicationFilePathExactMatch returns the location of a file if it exists in the directory and matches the given file name.
func GetAnyApplicationFilePathExactMatch(root string, propsFiles []model.ApplicationFileInfo) string {
	for _, propsFile := range propsFiles {
		fileToBeFound := filepath.Join(root, propsFile.Dir, propsFile.File)
		if _, err := os.Stat(fileToBeFound); !os.IsNotExist(err) {
			return fileToBeFound
		}
	}

	return ""
}

// GenerateApplicationFileFromFilters generates a slice of model.ApplicationFileInfo
// from a given list of files and the root path of a component. If suffix exists
// it generates items only for files ending with this suffix.
func GenerateApplicationFileFromFilters(files []string, path string, suffix string, ctx *context.Context) []model.ApplicationFileInfo {
	applicationFileInfos := []model.ApplicationFileInfo{}
	for _, file := range files {
		if strings.HasSuffix(file, suffix) {
			cleanPath := filepath.Clean(file)
			filename := filepath.Base(cleanPath)
			tmpDir := strings.ReplaceAll(file, path, "")
			dir := strings.ReplaceAll(tmpDir, filename, "")
			appFileInfo := model.ApplicationFileInfo{
				Context: ctx,
				Root:    path,
				Dir:     dir,
				File:    filename,
			}
			applicationFileInfos = append(applicationFileInfos, appFileInfo)
		}
	}
	return applicationFileInfos
}

// GetApplicationFileContents returns a slice of strings for all file contents found for a given
// slice of ApplicationFileInfo.
func GetApplicationFileContents(appFileInfos []model.ApplicationFileInfo) ([]string, error) {
	fileContents := []string{}
	for _, appFileInfo := range appFileInfos {
		fileContent, err := GetApplicationFileBytes(appFileInfo)
		if err == nil {
			fileContents = append(fileContents, string(fileContent))
		}
	}
	if len(fileContents) == 0 {
		return fileContents, fmt.Errorf("error: no application file found matching given criteria")
	}
	return fileContents, nil
}

// GetApplicationFileBytes returns a slice of bytes of a file if it exists in the directory and the given file name is a substring.
func GetApplicationFileBytes(propsFile model.ApplicationFileInfo) ([]byte, error) {
	bytes, err := readAnyApplicationFile(propsFile.Root, []model.ApplicationFileInfo{propsFile}, false, propsFile.Context)
	if err != nil {
		return bytes, fmt.Errorf("error: %s", err)
	}
	return bytes, nil
}

// GetApplicationFileInfo returns an item from a slice of applicationFileInfos if it matches the given filename
func GetApplicationFileInfo(propsFiles []model.ApplicationFileInfo, filename string) (model.ApplicationFileInfo, error) {
	for _, propsFile := range propsFiles {
		if propsFile.File == filename {
			return propsFile, nil
		}
	}
	return model.ApplicationFileInfo{}, fmt.Errorf("no ApplicationFileInfo found")
}

// ReadAnyApplicationFileExactMatch returns a byte slice if the exact given file exists in the directory.
func ReadAnyApplicationFileExactMatch(root string, propsFiles []model.ApplicationFileInfo) ([]byte, error) {
	return readAnyApplicationFile(root, propsFiles, true, nil)
}

// readAnyApplicationFile returns a byte of a file if it exists.
func readAnyApplicationFile(root string, propsFiles []model.ApplicationFileInfo, exactMatch bool, ctx *context.Context) ([]byte, error) {
	var path string
	if exactMatch {
		path = GetAnyApplicationFilePathExactMatch(root, propsFiles)
	} else {
		path = GetAnyApplicationFilePath(root, propsFiles, ctx)
	}
	if path != "" {
		return os.ReadFile(filepath.Clean(path))
	}
	return nil, errors.New("no file found")
}

// FindPortSubMatch returns a port number in case it finds one for a given regex group
func FindPortSubmatch(re *regexp.Regexp, text string, group int) int {
	potentialPortGroup := FindPotentialPortGroup(re, text, group)
	if potentialPortGroup != "" {
		if port, err := GetValidPort(potentialPortGroup); err == nil {
			return port
		}
	}
	return -1
}

// FindPotentialPortGroup returns a placeholder for port if is found
func FindPotentialPortGroup(re *regexp.Regexp, text string, group int) string {
	if text != "" {
		matches := re.FindStringSubmatch(text)
		if len(matches) > group {
			return matches[group]
		}
	}
	return ""
}

// FindAllPortsSubmatch returns a slice of port int values, matching a regex inside a given text
func FindAllPortsSubmatch(re *regexp.Regexp, text string, group int) []int {
	var ports []int
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

// GetPortValueFromEnvFile returns the first port value of a slice of port values
// declared from env var files.
func GetPortValueFromEnvFile(root string, regex string) int {
	ports := GetPortValuesFromEnvFile(root, []string{regex})
	if len(ports) > 0 {
		return ports[0]
	}
	return -1
}

// GetPortValuesFromEnvFile returns all port values found inside an env var file
func GetPortValuesFromEnvFile(root string, regexes []string) []int {
	var ports []int
	text, err := getEnvFileContent(root)
	if err != nil {
		return ports
	}
	for _, regex := range regexes {
		re, err := regexp.Compile(regex)
		if err != nil {
			continue
		}
		port := FindPortSubmatch(re, text, 1)
		if port != -1 {
			ports = append(ports, port)
		}
	}
	return ports
}

// GetStringValueFromEnvFile returns port values as string from env file
func GetStringValueFromEnvFile(root string, regex string) string {
	text, err := getEnvFileContent(root)
	if err != nil {
		return ""
	}
	re, err := regexp.Compile(regex)
	if err != nil {
		return ""
	}
	if text != "" {
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// getEnvFileContent is exposed as a global variable for the purpose of running mock tests
var getEnvFileContent = func(root string) (string, error) {
	envPath := filepath.Join(root, ".env")
	cleanEnvPath := filepath.Clean(envPath)
	bytes, err := os.ReadFile(cleanEnvPath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// NormalizeSplit splits a filepath into dir and filename
func NormalizeSplit(file string) (string, string) {
	dir, fileName := filepath.Split(file)
	if dir == "" {
		dir = "./"
	}
	return dir, fileName
}
