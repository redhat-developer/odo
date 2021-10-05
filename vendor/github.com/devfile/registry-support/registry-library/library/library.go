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
	"strings"
	"text/tabwriter"
	"time"

	orasctx "github.com/deislabs/oras/pkg/context"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

const (
	// Devfile media types
	DevfileConfigMediaType  = "application/vnd.devfileio.devfile.config.v2+json"
	DevfileMediaType        = "application/vnd.devfileio.devfile.layer.v1"
	DevfileVSXMediaType     = "application/vnd.devfileio.vsx.layer.v1.tar"
	DevfileSVGLogoMediaType = "image/svg+xml"
	DevfilePNGLogoMediaType = "image/png"
	DevfileArchiveMediaType = "application/x-tar"

	httpRequestTimeout    = 30 * time.Second // httpRequestTimeout configures timeout of all HTTP requests
	responseHeaderTimeout = 30 * time.Second // responseHeaderTimeout is the timeout to retrieve the server's response headers
)

var (
	DevfileMediaTypeList     = []string{DevfileMediaType}
	DevfileAllMediaTypesList = []string{DevfileMediaType, DevfilePNGLogoMediaType, DevfileSVGLogoMediaType, DevfileVSXMediaType, DevfileArchiveMediaType}
)

type Registry struct {
	registryURL      string
	registryContents []indexSchema.Schema
	err              error
}

// GetRegistryIndex returns the list of stacks and/or samples, more specifically
// it gets the stacks and/or samples content of the index of the specified registry
// for listing the stacks and/or samples
func GetRegistryIndex(registryURL string, skipTLSVerify bool, user string, devfileTypes ...indexSchema.DevfileType) ([]indexSchema.Schema, error) {
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
	if getStack && getSample {
		urlObj.Path = path.Join(urlObj.Path, "index", "all")
	} else if getStack && !getSample {
		urlObj.Path = path.Join(urlObj.Path, "index")
	} else if getSample && !getStack {
		urlObj.Path = path.Join(urlObj.Path, "index", "sample")
	} else {
		return registryIndex, nil
	}
	url := urlObj.String()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if user != "" {
		req.Header.Add("User", user)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: responseHeaderTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipTLSVerify},
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

// GetMultipleRegistryIndices returns returns the list of stacks and/or samples of multiple registries
func GetMultipleRegistryIndices(registryURLs []string, skipTLSVerify bool, user string, devfileTypes ...indexSchema.DevfileType) []Registry {
	registryList := make([]Registry, len(registryURLs))
	registryContentsChannel := make(chan []indexSchema.Schema)
	errChannel := make(chan error)

	for index, registryURL := range registryURLs {
		go func(chan []indexSchema.Schema, chan error) {
			registryContents, err := GetRegistryIndex(registryURL, skipTLSVerify, user, devfileTypes...)
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
func PrintRegistry(registryURLs string, devfileType string, skipTLSVerify bool, user string) error {
	// Get the registry index
	registryURLArray := strings.Split(registryURLs, ",")
	var registryList []Registry

	if devfileType == string(indexSchema.StackDevfileType) {
		registryList = GetMultipleRegistryIndices(registryURLArray, skipTLSVerify, user, indexSchema.StackDevfileType)
	} else if devfileType == string(indexSchema.SampleDevfileType) {
		registryList = GetMultipleRegistryIndices(registryURLArray, skipTLSVerify, user, indexSchema.SampleDevfileType)
	} else if devfileType == "all" {
		registryList = GetMultipleRegistryIndices(registryURLArray, skipTLSVerify, user, indexSchema.StackDevfileType, indexSchema.SampleDevfileType)
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

// PullStackByMediaTypesFromRegistry pulls stack from registry with allowed media types to the destination directory
func PullStackByMediaTypesFromRegistry(registry string, stack string, allowedMediaTypes []string, destDir string, skipTLSVerify bool, user string) error {
	// Get the registry index
	registryIndex, err := GetRegistryIndex(registry, skipTLSVerify, user, indexSchema.StackDevfileType)
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
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		},
	}
	headers := make(http.Header)
	if user != "" {
		headers.Add("User", user)
	}
	resolver := docker.NewResolver(docker.ResolverOptions{Headers: headers, PlainHTTP: plainHTTP, Client: httpClient})
	ref := path.Join(urlObj.Host, stackIndex.Links["self"])
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
		err := decompress(destDir, archivePath)
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

// PullStackFromRegistry pulls stack from registry with all stack resources (all media types) to the destination directory
func PullStackFromRegistry(registry string, stack string, destDir string, skipTLSVerify bool, user string) error {
	return PullStackByMediaTypesFromRegistry(registry, stack, DevfileAllMediaTypesList, destDir, skipTLSVerify, user)
}

// decompress extracts the archive file
func decompress(targetDir string, tarFile string) error {
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
