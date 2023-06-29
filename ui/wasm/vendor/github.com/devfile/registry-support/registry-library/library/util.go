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
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SplitVersionFromStack takes a stack/version tag and splits the stack name from the version
func SplitVersionFromStack(stackWithVersion string) (string, string, error) {
	var requestVersion string
	var stack string

	if valid, err := ValidateStackVersionTag(stackWithVersion); !valid {
		if err != nil {
			return "", "", err
		}
		return "", "", fmt.Errorf("stack/version tag '%s' is malformed, use form '<stack>:<version>' or '<stack>'",
			stackWithVersion)
	} else if strings.Contains(stackWithVersion, ":") {
		pair := strings.Split(stackWithVersion, ":")
		if len(pair) != 2 {
			return "", "", fmt.Errorf("problem splitting stack/version pair from tag '%s', instead of a pair got a length of %d",
				stackWithVersion, len(pair))
		}

		stack = pair[0]
		requestVersion = pair[1]
	} else {
		stack = stackWithVersion
		requestVersion = ""
	}

	return stack, requestVersion, nil
}

// ValidateStackVersionTag returns true if stack/version tag is well formed
// and false if it is malformed
func ValidateStackVersionTag(stackWithVersion string) (bool, error) {
	const exp = `^[a-z][^:\s]*(:([a-z]|[0-9])[^:\s]*)?$`
	r, err := regexp.Compile(exp)
	if err != nil {
		return false, err
	}

	return r.MatchString(stackWithVersion), nil
}

// decompress extracts the archive file
func decompress(targetDir string, tarFile string, excludeFiles []string) error {
	var returnedErr error

	reader, err := os.Open(filepath.Clean(tarFile))
	if err != nil {
		return err
	}

	defer func() {
		if err = reader.Close(); err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}
	}()

	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		returnedErr = multierror.Append(returnedErr, err)
		return returnedErr
	}

	defer func() {
		if err = gzReader.Close(); err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}
	}()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
			return returnedErr
		}
		if isExcluded(header.Name, excludeFiles) {
			continue
		}

		target := path.Join(targetDir, filepath.Clean(header.Name))
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, os.FileMode(header.Mode))
			if err != nil {
				returnedErr = multierror.Append(returnedErr, err)
				return returnedErr
			}
		case tar.TypeReg:
			/* #nosec G304 -- target is produced using path.Join which cleans the dir path */
			w, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				returnedErr = multierror.Append(returnedErr, err)
				return returnedErr
			}
			/* #nosec G110 -- starter projects are vetted before they are added to a registry.  Their contents can be seen before they are downloaded */
			_, err = io.Copy(w, tarReader)
			if err != nil {
				returnedErr = multierror.Append(returnedErr, err)
				return returnedErr
			}
			err = w.Close()
			if err != nil {
				returnedErr = multierror.Append(returnedErr, err)
				return returnedErr
			}
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

//getHTTPClient returns a new http client object
func getHTTPClient(options RegistryOptions) *http.Client {

	overriddenTimeout := httpRequestResponseTimeout
	timeout := options.HTTPTimeout
	//if value is invalid or unspecified, the default will be used
	if timeout != nil && *timeout > 0 {
		//convert timeout to seconds
		overriddenTimeout = time.Duration(*timeout) * time.Second
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: overriddenTimeout,
			/*#nosec G402 -- documented user option for dev/test, not for prod use */
			TLSClientConfig: &tls.Config{InsecureSkipVerify: options.SkipTLSVerify},
		},
		Timeout: overriddenTimeout,
	}
}
