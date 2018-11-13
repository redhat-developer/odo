package genericclioptions

import "github.com/redhat-developer/odo/pkg/occlient"

func NewFakeContext(project, application, component string, client *occlient.Client) *Context {
	return &Context{
		internalCxt{
			Client:      client,
			Project:     project,
			Application: application,
			cmp:         component,
		},
	}
}
