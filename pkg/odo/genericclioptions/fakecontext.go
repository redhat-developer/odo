package genericclioptions

import (
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
)

func NewFakeContext(project, application, component string, client *occlient.Client, kclient *kclient.Client) *Context {
	client.SetKubeClient(kclient)
	return &Context{
		internalCxt{
			client:      client,
			project:     &project,
			Application: application,
			cmp:         component,
		},
	}
}
