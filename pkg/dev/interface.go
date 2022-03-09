package dev

import (
	"io"

	"github.com/redhat-developer/odo/pkg/envinfo"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"

	"github.com/redhat-developer/odo/pkg/watch"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
)

type Client interface {
	Start(parser.DevfileObj, kubernetes.KubernetesContext, []string, string, io.Writer, io.Writer, Handler) error
	SetupPortForwarding(parser.DevfileObj, *envinfo.EnvSpecificInfo, io.Writer, io.Writer) error
	Cleanup() error
}

type Handler interface {
	RegenerateAdapterAndPush(common.PushParameters, watch.WatchParameters) error
}
