package data

import (
	"reflect"

	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	v200 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.0.0"
	v210 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.1.0"
)

// SupportedApiVersions stores the supported devfile API versions
type supportedApiVersion string

// Supported devfile API versions
const (
	APIVersion200 supportedApiVersion = "2.0.0"
	APIVersion210 supportedApiVersion = "2.1.0"
)

// ------------- Init functions ------------- //

// apiVersionToDevfileStruct maps supported devfile API versions to their corresponding devfile structs
var apiVersionToDevfileStruct map[supportedApiVersion]reflect.Type

// Initializes a map of supported devfile api versions and devfile structs
func init() {
	apiVersionToDevfileStruct = make(map[supportedApiVersion]reflect.Type)
	apiVersionToDevfileStruct[APIVersion200] = reflect.TypeOf(v2.DevfileV2{})
	apiVersionToDevfileStruct[APIVersion210] = reflect.TypeOf(v2.DevfileV2{})
}

// Map to store mappings between supported devfile API versions and respective devfile JSON schemas
var devfileApiVersionToJSONSchema map[supportedApiVersion]string

// init initializes a map of supported devfile apiVersions with it's respective devfile JSON schema
func init() {
	devfileApiVersionToJSONSchema = make(map[supportedApiVersion]string)
	devfileApiVersionToJSONSchema[APIVersion200] = v200.JsonSchema200
	devfileApiVersionToJSONSchema[APIVersion210] = v210.JsonSchema210
}
