//go:build tools
// +build tools

package tools

import (
	_ "github.com/frapposelli/wwhrd"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "github.com/securego/gosec/v2/cmd/gosec"
)

// https://github.com/onsi/ginkgo#go-module-tools-package

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
