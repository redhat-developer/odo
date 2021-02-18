package parser

import (
	"fmt"
	"net/url"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/devfile/library/pkg/util"
	"k8s.io/klog"
)

var URIMap = make(map[string]bool)

// DevfileCtx stores context info regarding devfile
type DevfileCtx struct {

	// devfile ApiVersion
	apiVersion string

	// absolute path of devfile
	absPath string

	// relative path of devfile
	relPath string

	// raw content of the devfile
	rawContent []byte

	// devfile json schema
	jsonSchema string

	//url path of the devfile
	url string

	// filesystem for devfile
	fs filesystem.Filesystem
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
func (d *DevfileCtx) Populate() (err error) {

	if err := d.SetAbsPath(); err != nil {
		return err
	}
	klog.V(4).Infof("absolute devfile path: '%s'", d.absPath)
	if URIMap[d.absPath] {
		return fmt.Errorf("URI %v is recursively referenced", d.absPath)
	}
	URIMap[d.absPath] = true
	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromURL fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) PopulateFromURL() (err error) {

	_, err = url.ParseRequestURI(d.url)
	if err != nil {
		return err
	}
	if URIMap[d.url] {
		return fmt.Errorf("URI %v is recursively referenced", d.url)
	}
	URIMap[d.url] = true
	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
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

// SetAbsPath sets absolute file path for devfile
func (d *DevfileCtx) SetAbsPath() (err error) {
	// Set devfile absolute path
	if d.absPath, err = util.GetAbsPath(d.relPath); err != nil {
		return err
	}
	klog.V(2).Infof("absolute devfile path: '%s'", d.absPath)

	return nil

}
