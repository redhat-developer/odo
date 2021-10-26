package genericclioptions

import (
	"github.com/openshift/odo/v2/pkg/kclient"
	"github.com/openshift/odo/v2/pkg/occlient"
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
