package parser

import (
	"fmt"

	"github.com/cli-playground/devfile-parser/pkg/testingutil/filesystem"
	"github.com/cli-playground/devfile-parser/pkg/util"
	"k8s.io/klog"
)

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

	// filesystem for devfile
	Fs filesystem.Filesystem
}

// NewDevfileCtx returns a new DevfileCtx type object
func NewDevfileCtx(path string) DevfileCtx {
	return DevfileCtx{
		relPath: path,
		Fs:      filesystem.DefaultFs{},
	}
}

// populateDevfile checks the API version is supported and returns the JSON schema for the given devfile API Version
func (d *DevfileCtx) populateDevfile() (err error) {

	// Get devfile APIVersion
	if err := d.SetDevfileAPIVersion(); err != nil {
		return err
	}

	// Check if the apiVersion is supported
	if !d.IsApiVersionSupported() {
		return fmt.Errorf("devfile apiVersion '%s' not supported in odo", d.apiVersion)
	}
	klog.V(4).Infof("devfile apiVersion '%s' is supported in odo", d.apiVersion)

	// Read and save devfile JSON schema for provided apiVersion
	return d.SetDevfileJSONSchema()
}

// Populate fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) Populate() (err error) {

	// Get devfile absolute path
	if d.absPath, err = util.GetAbsPath(d.relPath); err != nil {
		return err
	}
	klog.V(4).Infof("absolute devfile path: '%s'", d.absPath)

	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromBytes fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) PopulateFromBytes(bytes []byte) (err error) {

	// Read and save devfile content
	if err := d.SetDevfileContentFromBytes(bytes); err != nil {
		return err
	}
	return d.populateDevfile()
}

// Validate func validates devfile JSON schema for the given apiVersion
func (d *DevfileCtx) Validate() error {

	// Validate devfile
	return d.ValidateDevfileSchema()
}
