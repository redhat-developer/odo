/*   Copyright 2020-2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package library

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"

	orasctx "oras.land/oras-go/pkg/context"

	"github.com/containerd/containerd/remotes/docker"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	versionpkg "github.com/hashicorp/go-version"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

const (
	// Supported Devfile media types
	DevfileMediaType        = "application/vnd.devfileio.devfile.layer.v1"
	DevfileVSXMediaType     = "application/vnd.devfileio.vsx.layer.v1.tar"
	DevfileSVGLogoMediaType = "image/svg+xml"
	DevfilePNGLogoMediaType = "image/png"
	DevfileArchiveMediaType = "application/x-tar"

	OwnersFile = "OWNERS"

	httpRequestResponseTimeout = 30 * time.Second // httpRequestTimeout configures timeout of all HTTP requests
)

var (
	DevfileMediaTypeList     = []string{DevfileMediaType}
	DevfileAllMediaTypesList = []string{DevfileMediaType, DevfilePNGLogoMediaType, DevfileSVGLogoMediaType, DevfileVSXMediaType, DevfileArchiveMediaType}
	ExcludedFiles            = []string{OwnersFile}
)

type Registry struct {
	registryURL      string
	registryContents []indexSchema.Schema
	err              error
}

//TelemetryData structure to pass in client telemetry information
// The User and Locale fields should be passed in by clients if telemetry opt-in is enabled
// the generic Client name will be passed in regardless of opt-in/out choice.  The value
// will be assigned to the UserId field for opt-outs
type TelemetryData struct {
	// User is a generated UUID or generic client name
	User string
	// Locale is the OS or browser locale
	Locale string
	//Client is a generic name that describes the client
	Client string
}

type RegistryOptions struct {
	// SkipTLSVerify is false by default which is the recommended setting for a devfile registry deployed in production.  SkipTLSVerify should only be set to true
	// if you are testing a devfile registry or proxy server that is set up with self-signed certificates in a pre-production environment.
	SkipTLSVerify bool
	// Telemetry allows clients to send telemetry data to the community Devfile Registry
	Telemetry TelemetryData
	// Filter allows clients to specify which architectures they want to filter their devfiles on
	Filter RegistryFilter
	// NewIndexSchema is false by default, which calls GET /index and returns index of default version of each stack using the old index schema struct.
	// If specified to true, calls GET /v2index and returns the new Index schema with multi-version support
	NewIndexSchema bool
	// HTTPTimeout overrides the request and response timeout values for the custom HTTP clients set by the registry library.  If unset or a negative value is specified, the default timeout of 30s will be used.
	HTTPTimeout *int
}

type RegistryFilter struct {
	Architectures []string
	// MinSchemaVersion is set to filter devfile index equal and above a particular devfile schema version (inclusive)
	// only major version and minor version are required. e.g. 2.1, 2.2 ect. service version should not be provided.
	// will only be applied if `NewIndexSchema=true`
	MinSchemaVersion string
	// MaxSchemaVersion is set to filter devfile index equal and below a particular devfile schema version (inclusive)
	// only major version and minor version are required. e.g. 2.1, 2.2 ect. service version should not be provided.
	// will only be applied if `NewIndexSchema=true`
	MaxSchemaVersion string
}

// GetRegistryIndex returns the list of index schema structured stacks and/or samples from a specified devfile registry.
func GetRegistryIndex(registryURL string, options RegistryOptions, devfileTypes ...indexSchema.DevfileType) ([]indexSchema.Schema, error) {
	var registryIndex []indexSchema.Schema

	// Call index server REST API to get the index
	urlObj, err := url.Parse(registryURL)
	if err != nil {
		return nil, err
	}
	getStack := false
	getSample := false
	for _, devfileType := range devfileTypes {
		if devfileType == indexSchema.StackDevfileType {
			getStack = true
		} else if devfileType == indexSchema.SampleDevfileType {
			getSample = true
		}
	}

	var endpoint string
	indexEndpoint := "index"
	if options.NewIndexSchema {
		indexEndpoint = "v2index"
	}
	if getStack && getSample {
		endpoint = path.Join(indexEndpoint, "all")
	} else if getStack && !getSample {
		endpoint = indexEndpoint
	} else if getSample && !getStack {
		endpoint = path.Join(indexEndpoint, "sample")
	} else {
		return registryIndex, nil
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	urlObj = urlObj.ResolveReference(endpointURL)

	if !reflect.DeepEqual(options.Filter, RegistryFilter{}) {
		q := urlObj.Query()
		if len(options.Filter.Architectures) > 0 {
			for _, arch := range options.Filter.Architectures {
				q.Add("arch", arch)
			}
		}

		if options.NewIndexSchema && (options.Filter.MaxSchemaVersion != "" || options.Filter.MinSchemaVersion != "") {
			if options.Filter.MinSchemaVersion != "" {
				q.Add("minSchemaVersion", options.Filter.MinSchemaVersion)
			}
			if options.Filter.MaxSchemaVersion != "" {
				q.Add("maxSchemaVersion", options.Filter.MaxSchemaVersion)
			}
		}
		urlObj.RawQuery = q.Encode()
	}

	url := urlObj.String()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	setHeaders(&req.Header, options)

	httpClient := getHTTPClient(options)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &registryIndex)
	if err != nil {
		return nil, err
	}
	return registryIndex, nil
}

// GetMultipleRegistryIndices returns the list of stacks and/or samples from multiple registries
func GetMultipleRegistryIndices(registryURLs []string, options RegistryOptions, devfileTypes ...indexSchema.DevfileType) []Registry {
	registryList := make([]Registry, len(registryURLs))
	registryContentsChannel := make(chan []indexSchema.Schema)
	errChannel := make(chan error)

	for index, registryURL := range registryURLs {
		go func(chan []indexSchema.Schema, chan error) {
			registryContents, err := GetRegistryIndex(registryURL, options, devfileTypes...)
			registryContentsChannel <- registryContents
			errChannel <- err
		}(registryContentsChannel, errChannel)
		registryList[index].registryURL = registryURL
		registryList[index].registryContents = <-registryContentsChannel
		registryList[index].err = <-errChannel
	}
	return registryList
}

// PrintRegistry prints the registry with devfile type
func PrintRegistry(registryURLs string, devfileType string, options RegistryOptions) error {
	// Get the registry index
	registryURLArray := strings.Split(registryURLs, ",")
	var registryList []Registry

	if devfileType == string(indexSchema.StackDevfileType) {
		registryList = GetMultipleRegistryIndices(registryURLArray, options, indexSchema.StackDevfileType)
	} else if devfileType == string(indexSchema.SampleDevfileType) {
		registryList = GetMultipleRegistryIndices(registryURLArray, options, indexSchema.SampleDevfileType)
	} else if devfileType == "all" {
		registryList = GetMultipleRegistryIndices(registryURLArray, options, indexSchema.StackDevfileType, indexSchema.SampleDevfileType)
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Name", "\t", "Description", "\t", "Registry", "\t", "Error", "\t")
	for _, devfileRegistry := range registryList {
		if devfileRegistry.err != nil {
			fmt.Fprintln(w, "NONE", "\t", "NONE", "\t", devfileRegistry.registryURL, devfileRegistry.err.Error(), "\t")
		} else {
			for _, devfileEntry := range devfileRegistry.registryContents {
				fmt.Fprintln(w, devfileEntry.Name, "\t", devfileEntry.Description, "\t", devfileRegistry.registryURL, "\t", "NONE", "\t")
			}
		}
	}

	_ = w.Flush()
	return nil
}

// PullStackByMediaTypesFromRegistry pulls a specified stack with allowed media types from a given registry URL to the destination directory.
// OWNERS files present in the registry will be excluded
func PullStackByMediaTypesFromRegistry(registry string, stack string, allowedMediaTypes []string, destDir string, options RegistryOptions) error {
	// Get stack link
	stackLink, err := GetStackLink(registry, stack, options)
	if err != nil {
		return err
	}

	// Pull stack initialization
	ctx := orasctx.Background()
	urlObj, err := url.Parse(registry)
	if err != nil {
		return err
	}
	plainHTTP := true
	if urlObj.Scheme == "https" {
		plainHTTP = false
	}
	httpClient := getHTTPClient(options)

	headers := make(http.Header)
	setHeaders(&headers, options)

	resolver := docker.NewResolver(docker.ResolverOptions{Headers: headers, PlainHTTP: plainHTTP, Client: httpClient})
	ref := path.Join(urlObj.Host, stackLink)
	fileStore := content.NewFile(destDir)
	defer fileStore.Close()

	// Pull stack from registry and save it to disk
	_, err = oras.Copy(ctx, resolver, ref, fileStore, ref, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return fmt.Errorf("failed to pull stack %s from %s with allowed media types %v: %v", stack, ref, allowedMediaTypes, err)
	}

	// Decompress archive.tar
	archivePath := filepath.Join(destDir, "archive.tar")
	if _, err := os.Stat(archivePath); err == nil {
		err := decompress(destDir, archivePath, ExcludedFiles)
		if err != nil {
			return err
		}

		err = os.RemoveAll(archivePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// PullStackFromRegistry pulls a specified stack with all devfile supported media types from a registry URL to the destination directory
func PullStackFromRegistry(registry string, stack string, destDir string, options RegistryOptions) error {
	return PullStackByMediaTypesFromRegistry(registry, stack, DevfileAllMediaTypesList, destDir, options)
}

// DownloadStarterProjectAsDir downloads a specified starter project archive and extracts it to given path
func DownloadStarterProjectAsDir(path string, registryURL string, stack string, starterProject string, options RegistryOptions) error {
	var err error

	// Create temp path to download archive
	archivePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.zip", starterProject))

	// Download starter project archive to temp path
	if err = DownloadStarterProject(archivePath, registryURL, stack, starterProject, options); err != nil {
		return err
	}

	// Open archive reader
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("error opening downloaded starter project archive: %v", err)
	}
	defer archive.Close()

	// Extract files from starter project archive to specified directory path
	cleanPath := filepath.Clean(path)
	for _, file := range archive.File {
		filePath := filepath.Join(cleanPath, filepath.Clean(file.Name))

		// validate extracted filepath
		if filePath != file.Name && !strings.HasPrefix(filePath, cleanPath+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path %s", filePath)
		}

		// if file is a directory, create it in destination and continue to next file
		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
				return fmt.Errorf("error creating directory %s: %v", filepath.Dir(filePath), err)
			}
			continue
		}

		// ensure parent directory of current file is created in destination
		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating parent directory %s: %v", filepath.Dir(filePath), err)
		}

		// open destination file
		/* #nosec G304 -- filePath is produced using path.Join which cleans the dir path */
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("error opening destination file at %s: %v", filePath, err)
		}

		// open source file in archive
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening source file %s in archive %s: %v", file.Name, archivePath, err)
		}

		// extract source file to destination file
		/* #nosec G110 -- starter projects are vetted before they are added to a registry.  Their contents can be seen before they are downloaded */
		if _, err = io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("error extracting file %s from archive %s to destination at %s: %v", file.Name, archivePath, filePath, err)
		}

		err = dstFile.Close()
		if err != nil {
			return err
		}
		err = srcFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadStarterProject downloads a specified starter project archive to a given path
func DownloadStarterProject(path string, registryURL string, stack string, starterProject string, options RegistryOptions) error {
	var fileStream *os.File
	var returnedErr error

	cleanPath := filepath.Clean(path)
	// Download Starter Project archive bytes
	bytes, err := DownloadStarterProjectAsBytes(registryURL, stack, starterProject, options)
	if err != nil {
		return err
	}

	// Error if parent directory does not exist
	if _, err = os.Stat(filepath.Dir(cleanPath)); os.IsNotExist(err) {
		return fmt.Errorf("parent directory '%s' does not exist: %v", filepath.Dir(path), err)
	}

	// If file does not exist, create a new one
	// Else open existing for overwriting
	if _, err = os.Stat(path); os.IsNotExist(err) {
		fileStream, err = os.Create(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to create file '%s': %v", path, err)
		}
	} else {
		fileStream, err = os.OpenFile(cleanPath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to open file '%s': %v", path, err)
		}
	}

	defer func() {
		if err = fileStream.Close(); err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}
	}()

	// Write downloaded bytes to file
	_, err = fileStream.Write(bytes)
	if err != nil {
		returnedErr = multierror.Append(returnedErr, fmt.Errorf("failed writing to '%s': %v", path, err))
		return returnedErr
	}

	return nil
}

// DownloadStarterProjectAsBytes downloads the file bytes of a specified starter project archive and return these bytes
func DownloadStarterProjectAsBytes(registryURL string, stack string, starterProject string, options RegistryOptions) ([]byte, error) {
	stackName, _, err := SplitVersionFromStack(stack)
	if err != nil {
		return nil, fmt.Errorf("problem in stack/version tag: %v", err)
	}

	exists, err := IsStarterProjectExists(registryURL, stack, starterProject, options)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("the starter project '%s' does not exist under the stack '%s'", starterProject, stackName)
	}

	urlObj, err := url.Parse(registryURL)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s://%s", urlObj.Scheme, path.Join(urlObj.Host, "devfiles", stackName, "starter-projects", starterProject))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	setHeaders(&req.Header, options)

	httpClient := getHTTPClient(options)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Return downloaded starter project as bytes or error if unsuccessful.
	return ioutil.ReadAll(resp.Body)
}

// IsStarterProjectExists checks if starter project exists for a given stack
func IsStarterProjectExists(registryURL string, stack string, starterProject string, options RegistryOptions) (bool, error) {
	// Get stack index
	stackIndex, err := GetStackIndex(registryURL, stack, options)
	if err != nil {
		return false, err
	}

	// Check if starter project exists in the stack index
	exists := false
	for _, sp := range stackIndex.StarterProjects {
		if sp == starterProject {
			exists = true
			break
		}
	}

	if !exists && options.NewIndexSchema {
		var starterProjects []string

		for _, version := range stackIndex.Versions {
			starterProjects = append(starterProjects, version.StarterProjects...)
		}

		exists = false
		for _, sp := range starterProjects {
			if sp == starterProject {
				exists = true
				break
			}
		}

		return exists, nil
	} else {
		return exists, nil
	}
}

// GetStackLink returns the slug needed to pull a specified stack from a registry URL
func GetStackLink(registryURL string, stack string, options RegistryOptions) (string, error) {
	var stackLink string

	// Get stack index
	stackIndex, err := GetStackIndex(registryURL, stack, options)
	if err != nil {
		return "", err
	}

	// Split version from stack label if specified
	stack, requestVersion, err := SplitVersionFromStack(stack)
	if err != nil {
		return "", fmt.Errorf("problem in stack/version tag: %v", err)
	}

	if options.NewIndexSchema {
		latestVersionIndex := 0
		latest, err := versionpkg.NewVersion(stackIndex.Versions[latestVersionIndex].Version)
		if err != nil {
			return "", fmt.Errorf("failed to parse the stack version %s for stack %s", stackIndex.Versions[latestVersionIndex].Version, stack)
		}
		for index, version := range stackIndex.Versions {
			if (requestVersion == "" && version.Default) || (version.Version == requestVersion) {
				stackLink = version.Links["self"]
				break
			} else if requestVersion == "latest" {
				current, err := versionpkg.NewVersion(version.Version)
				if err != nil {
					return "", fmt.Errorf("failed to parse the stack version %s for stack %s", version.Version, stack)
				}
				if current.GreaterThan(latest) {
					latestVersionIndex = index
					latest = current
				}
			}
		}
		if requestVersion == "latest" {
			stackLink = stackIndex.Versions[latestVersionIndex].Links["self"]
		}
		if requestVersion == "" && stackLink == "" {
			return "", fmt.Errorf("no version specified for stack %s which no default version exists in the registry %s", stack, registryURL)
		} else if stackLink == "" {
			return "", fmt.Errorf("the requested version %s for stack %s does not exist in the registry %s", requestVersion, stack, registryURL)
		}
	} else {
		stackLink = stackIndex.Links["self"]
	}

	return stackLink, nil
}

// GetStackIndex returns the schema index of a specified stack
func GetStackIndex(registryURL string, stack string, options RegistryOptions) (indexSchema.Schema, error) {
	// Get the registry index
	registryIndex, err := GetRegistryIndex(registryURL, options, indexSchema.StackDevfileType)
	if err != nil {
		return indexSchema.Schema{}, err
	}

	// Prune version from stack label if specified
	stack, _, err = SplitVersionFromStack(stack)
	if err != nil {
		return indexSchema.Schema{}, fmt.Errorf("problem in stack/version tag: %v", err)
	}

	// Parse the index to get the specified stack's metadata in the index
	for _, item := range registryIndex {
		// Return index of specified stack if found
		if item.Name == stack {
			return item, nil
		}
	}

	// Return error if stack index is not found
	return indexSchema.Schema{}, fmt.Errorf("stack %s does not exist in the registry %s", stack, registryURL)
}
