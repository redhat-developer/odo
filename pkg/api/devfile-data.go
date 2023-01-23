package api

import "github.com/devfile/library/v2/pkg/devfile/parser/data"

// DevfileData describes a devfile content
type DevfileData struct {
	Devfile              data.DevfileData      `json:"devfile"`
	SupportedOdoFeatures *SupportedOdoFeatures `json:"supportedOdoFeatures,omitempty"`
}

// SupportedOdoFeatures indicates the support of high-level (odo) features by a devfile component
type SupportedOdoFeatures struct {
	Dev    bool `json:"dev"`
	Deploy bool `json:"deploy"`
	Debug  bool `json:"debug"`
}
