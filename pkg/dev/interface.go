package dev

import (
	"io"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"

	"github.com/redhat-developer/odo/pkg/watch"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
)

type Client interface {
	Start(d parser.DevfileObj, platformContext kubernetes.KubernetesContext, ignorePaths []string, path string, w io.Writer, e io.Writer, h Handler) error
	Cleanup() error
}

type Handler interface {
	RegenerateAdapterAndPush(common.PushParameters, watch.WatchParameters) error
}
