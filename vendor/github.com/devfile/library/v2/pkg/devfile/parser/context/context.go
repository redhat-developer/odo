//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"net/url"

	parserUtil "github.com/devfile/library/v2/pkg/devfile/parser/util"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/devfile/library/v2/pkg/util"
	"k8s.io/klog"
)

// DevfileCtx stores context info regarding devfile
type DevfileCtx struct {

	// devfile ApiVersion
	apiVersion string

	// absolute path of devfile
	absPath string

	// relative path of devfile.
	// It can also be a relative or absolute path to a folder containing one or more devfiles,
	// in which case the library will try to pick an existing one, based on the following priority order:
	// devfile.yaml > .devfile.yaml > devfile.yml > .devfile.yml
	relPath string

	// raw content of the devfile
	rawContent []byte

	// devfile json schema
	jsonSchema string

	// url path of the devfile
	url string

	// token is a personal access token used with a private git repo URL
	token string

	// filesystem for devfile
	fs filesystem.Filesystem

	// devfile kubernetes components has been converted from uri to inlined in memory
	convertUriToInlined bool
}

// NewDevfileCtx returns a new DevfileCtx type object
func NewDevfileCtx(path string) DevfileCtx {
	return DevfileCtx{
		relPath: path,
		fs:      filesystem.DefaultFs{},
	}
}

// NewURLDevfileCtx returns a new DevfileCtx type object
func NewURLDevfileCtx(url string) DevfileCtx {
	return DevfileCtx{
		url: url,
	}
}

// NewByteContentDevfileCtx set devfile content from byte data and returns a new DevfileCtx type object and error
func NewByteContentDevfileCtx(data []byte) (d DevfileCtx, err error) {
	err = d.SetDevfileContentFromBytes(data)
	if err != nil {
		return DevfileCtx{}, err
	}
	return d, nil
}

// populateDevfile checks the API version is supported and returns the JSON schema for the given devfile API Version
func (d *DevfileCtx) populateDevfile() (err error) {

	// Get devfile APIVersion
	if err := d.SetDevfileAPIVersion(); err != nil {
		return err
	}

	// Read and save devfile JSON schema for provided apiVersion
	return d.SetDevfileJSONSchema()
}

// Populate fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) Populate(devfileUtilsClient parserUtil.DevfileUtils) (err error) {
	d.relPath, err = lookupDevfileFromPath(d.fs, d.relPath)
	if err != nil {
		return err
	}
	if err = d.SetAbsPath(); err != nil {
		return err
	}
	klog.V(4).Infof("absolute devfile path: '%s'", d.absPath)
	// Read and save devfile content
	if err = d.SetDevfileContent(devfileUtilsClient); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromURL fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) PopulateFromURL(devfileUtilsClient parserUtil.DevfileUtils) (err error) {
	_, err = url.ParseRequestURI(d.url)
	if err != nil {
		return err
	}
	// Read and save devfile content
	if err := d.SetDevfileContent(devfileUtilsClient); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromRaw fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) PopulateFromRaw() (err error) {
	return d.populateDevfile()
}

// Validate func validates devfile JSON schema for the given apiVersion
func (d *DevfileCtx) Validate() error {

	// Validate devfile
	return d.ValidateDevfileSchema()
}

// GetAbsPath func returns current devfile absolute path
func (d *DevfileCtx) GetAbsPath() string {
	return d.absPath
}

// GetURL func returns current devfile absolute URL address
func (d *DevfileCtx) GetURL() string {
	return d.url
}

// GetToken func returns current devfile token
func (d *DevfileCtx) GetToken() string {
	return d.token
}

// SetToken sets the token for the devfile
func (d *DevfileCtx) SetToken(token string) {
	d.token = token
}

// SetAbsPath sets absolute file path for devfile
func (d *DevfileCtx) SetAbsPath() (err error) {
	// Set devfile absolute path
	if d.absPath, err = util.GetAbsPath(d.relPath); err != nil {
		return err
	}
	klog.V(2).Infof("absolute devfile path: '%s'", d.absPath)

	return nil

}

// GetConvertUriToInlined func returns if the devfile kubernetes comp has been converted from uri to inlined
func (d *DevfileCtx) GetConvertUriToInlined() bool {
	return d.convertUriToInlined
}

// SetConvertUriToInlined sets if the devfile kubernetes comp has been converted from uri to inlined
func (d *DevfileCtx) SetConvertUriToInlined(value bool) {
	d.convertUriToInlined = value
}
