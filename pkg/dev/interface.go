package dev

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"io"
)

type Client interface {
	Start(d parser.DevfileObj, w io.Writer, path string, platformContext kubernetes.KubernetesContext) error
	Cleanup() error
}
