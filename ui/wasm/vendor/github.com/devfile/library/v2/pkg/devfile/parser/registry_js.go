//go:build js

package parser

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func parseFromRegistry(importReference v1.ImportReference, resolveCtx *resolutionContextTree, tool resolverTools) (d DevfileObj, err error) {
	return DevfileObj{}, nil
}
