package genericclioptions

import (
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
)

func NewFakeContext(project, application, component string, client *occlient.Client, kclient kclient.ClientInterface) *Context {
	return &Context{
		internalCxt{
			Client:      client,
			KClient:     kclient,
			Project:     project,
			Application: application,
			cmp:         component,
		},
	}
}
