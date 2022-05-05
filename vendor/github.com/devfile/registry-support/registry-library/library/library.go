//
// Copyright (c) 2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package library

import (
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

	httpRequestTimeout    = 30 * time.Second // httpRequestTimeout configures timeout of all HTTP requests
	responseHeaderTimeout = 30 * time.Second // responseHeaderTimeout is the timeout to retrieve the server's response headers
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
	// if you are testing a devfile registry that is set up with self-signed certificates in a pre-production environment.
	SkipTLSVerify bool
	// Telemetry allows clients to send telemetry data to the community Devfile Registry
	Telemetry TelemetryData
	// Filter allows clients to specify which architectures they want to filter their devfiles on
	Filter RegistryFilter
	// NewIndexSchema is false by default, which calls GET /index and returns index of default version of each stack using the old index schema struct.
	// If specified to true, calls GET /v2index and returns the new Index schema with multi-version support
	NewIndexSchema bool
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

	httpClient := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: responseHeaderTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: options.SkipTLSVerify},
		},
		Timeout: httpRequestTimeout,
	}
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
	w.Flush()
	return nil
}

// PullStackByMediaTypesFromRegistry pulls a specified stack with allowed media types from a given registry URL to the destination directory.
// OWNERS files present in the registry will be excluded
func PullStackByMediaTypesFromRegistry(registry string, stack string, allowedMediaTypes []string, destDir string, options RegistryOptions) error {
	var requestVersion string
	if strings.Contains(stack, ":") {
		stackWithVersion := strings.Split(stack, ":")
		stack = stackWithVersion[0]
		requestVersion = stackWithVersion[1]
	}
	// Get the registry index
	registryIndex, err := GetRegistryIndex(registry, options, indexSchema.StackDevfileType)
	if err != nil {
		return err
	}

	// Parse the index to get the specified stack's metadata in the index
	var stackIndex indexSchema.Schema
	exist := false
	for _, item := range registryIndex {
		if item.Name == stack {
			stackIndex = item
			exist = true
			break
		}
	}
	if !exist {
		return fmt.Errorf("stack %s does not exist in the registry %s", stack, registry)
	}
	var stackLink string

	if options.NewIndexSchema {
		latestVersionIndex := 0
		latest, err := versionpkg.NewVersion(stackIndex.Versions[latestVersionIndex].Version)
		if err != nil {
			return fmt.Errorf("failed to parse the stack version %s for stack %s", stackIndex.Versions[latestVersionIndex].Version, stack)
		}
		for index, version := range stackIndex.Versions {
			if (requestVersion == "" && version.Default) || (version.Version == requestVersion) {
				stackLink = version.Links["self"]
				break
			} else if requestVersion == "latest" {
				current, err := versionpkg.NewVersion(version.Version)
				if err != nil {
					return fmt.Errorf("failed to parse the stack version %s for stack %s", version.Version, stack)
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
		if stackLink == "" {
			return fmt.Errorf("the requested verion %s for stack %s does not exist in the registry %s", requestVersion, stack, registry)
		}
	} else {
		stackLink = stackIndex.Links["self"]
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
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: options.SkipTLSVerify},
		},
	}
	headers := make(http.Header)
	setHeaders(&headers, options)

	resolver := docker.NewResolver(docker.ResolverOptions{Headers: headers, PlainHTTP: plainHTTP, Client: httpClient})
	ref := path.Join(urlObj.Host, stackLink)
	fileStore := content.NewFileStore(destDir)
	defer fileStore.Close()

	// Pull stack from registry and save it to disk
	_, _, err = oras.Pull(ctx, resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
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

// decompress extracts the archive file
func decompress(targetDir string, tarFile string, excludeFiles []string) error {
	reader, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if isExcluded(header.Name, excludeFiles) {
			continue
		}

		target := path.Join(targetDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			w, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(w, tarReader)
			if err != nil {
				return err
			}
			w.Close()
		default:
			log.Printf("Unsupported type: %v", header.Typeflag)
		}
	}

	return nil
}

func isExcluded(name string, excludeFiles []string) bool {
	basename := filepath.Base(name)
	for _, excludeFile := range excludeFiles {
		if basename == excludeFile {
			return true
		}
	}
	return false
}

//setHeaders sets the request headers
func setHeaders(headers *http.Header, options RegistryOptions) {
	t := options.Telemetry
	if t.User != "" {
		headers.Add("User", t.User)
	}
	if t.Client != "" {
		headers.Add("Client", t.Client)
	}
	if t.Locale != "" {
		headers.Add("Locale", t.Locale)
	}
}
