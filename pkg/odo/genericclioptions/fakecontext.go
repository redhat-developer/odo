package genericclioptions

import (
	"github.com/redhat-developer/odo/pkg/kclient"
)

func NewFakeContext(project, application, component string, kclient kclient.ClientInterface) *Context {
	return &Context{
		internalCxt{
			KClient:     kclient,
			project:     project,
			application: application,
			component:   component,
		},
	}
}
