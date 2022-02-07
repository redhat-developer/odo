package dev

import (
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/watch"
	"io"
)

var _ Client = (*Dev)(nil)

type Dev struct {
	client kclient.ClientInterface
	// devfileObj is stored for Cleanup; ideally populated by Start method
	//devfileObj parser.DevfileObj
}

func NewDev(client kclient.ClientInterface) *Dev {
	return &Dev{client: client}
}

// GetComponents returns a slice of components to be started for inner loop
func (o *Dev) GetComponents() (devfile.Component, error) {
	var components devfile.Component
	var err error
	return components, err
}

// Start starts the resources on the Kubernetes cluster
func (o *Dev) Start(devfileObj parser.DevfileObj, out io.Writer, path string) error {
	var err error
	// store the devfileObj so that we can reuse it in Cleanup
	//o.devfileObj = devfileObj
	watchParamaters := watch.WatchParameters{
		Path:            path,
		ComponentName:   devfileObj.GetMetadataName(),
		ApplicationName: "app",
		ExtChan:         make(chan bool),
	}
	watch.WatchAndPush(o.client, out, watchParamaters)
	return err
}

// Cleanup cleans the resources created by Push
func (o *Dev) Cleanup() error {
	var err error
	return err
}
