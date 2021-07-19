package data

import (
	"reflect"

	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	v200 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.0.0"
	v210 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.1.0"
	v220 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.2.0"
)

// SupportedApiVersions stores the supported devfile API versions
type supportedApiVersion string

// Supported devfile API versions
const (
	APISchemaVersion200 supportedApiVersion = "2.0.0"
	APISchemaVersion210 supportedApiVersion = "2.1.0"
	APISchemaVersion220 supportedApiVersion = "2.2.0"
	APIVersionAlpha2    supportedApiVersion = "v1alpha2"
)

// ------------- Init functions ------------- //

// apiVersionToDevfileStruct maps supported devfile API versions to their corresponding devfile structs
var apiVersionToDevfileStruct map[supportedApiVersion]reflect.Type

// Initializes a map of supported devfile api versions and devfile structs
func init() {
	apiVersionToDevfileStruct = make(map[supportedApiVersion]reflect.Type)
	apiVersionToDevfileStruct[APISchemaVersion200] = reflect.TypeOf(v2.DevfileV2{})
	apiVersionToDevfileStruct[APISchemaVersion210] = reflect.TypeOf(v2.DevfileV2{})
	apiVersionToDevfileStruct[APISchemaVersion220] = reflect.TypeOf(v2.DevfileV2{})
	apiVersionToDevfileStruct[APIVersionAlpha2] = reflect.TypeOf(v2.DevfileV2{})
}

// Map to store mappings between supported devfile API versions and respective devfile JSON schemas
var devfileApiVersionToJSONSchema map[supportedApiVersion]string

// init initializes a map of supported devfile apiVersions with it's respective devfile JSON schema
func init() {
	devfileApiVersionToJSONSchema = make(map[supportedApiVersion]string)
	devfileApiVersionToJSONSchema[APISchemaVersion200] = v200.JsonSchema200
	devfileApiVersionToJSONSchema[APISchemaVersion210] = v210.JsonSchema210
	devfileApiVersionToJSONSchema[APISchemaVersion220] = v220.JsonSchema220
	// should use hightest v2 schema version since it is expected to be backward compatible with the same api version
	devfileApiVersionToJSONSchema[APIVersionAlpha2] = v210.JsonSchema210
}
