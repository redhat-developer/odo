package dev

import (
	"io"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"

	"github.com/redhat-developer/odo/pkg/watch"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
)

type Client interface {
	Start(parser.DevfileObj, kubernetes.KubernetesContext, string) error
	SetupPortForwarding(parser.DevfileObj, string, io.Writer, io.Writer) (map[string]string, error)
	Watch(parser.DevfileObj, string, []string, io.Writer, Handler) error
	Cleanup() error
}

type Handler interface {
	RegenerateAdapterAndPush(common.PushParameters, watch.WatchParameters) error
}
