package api

import "github.com/devfile/library/pkg/devfile/parser/data"

// DevfileData describes a devfile content
type DevfileData struct {
	Devfile              data.DevfileData
	SupportedOdoFeatures SupportedOdoFeatures
}

// SupportedOdoFeatures indicates the support of high-level (odo) features by a devfile component
type SupportedOdoFeatures struct {
	Dev    bool
	Deploy bool
	Debug  bool
}
