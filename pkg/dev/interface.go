package dev

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"io"
)

type Client interface {
	Start(d parser.DevfileObj, platformContext kubernetes.KubernetesContext, ignorePaths []string, path string, w io.Writer) error
	Cleanup() error
}
