package delete

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/component"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

type DeleteComponentClient struct {
	kubeClient kclient.ClientInterface
}

func NewDeleteComponentClient(kubeClient kclient.ClientInterface) *DeleteComponentClient {
	return &DeleteComponentClient{
		kubeClient: kubeClient,
	}
}

func (o *DeleteComponentClient) UnDeploy(devfileObj parser.DevfileObj, path string) error {
	undeployHandler := newUndeployHandler(path, o.kubeClient)
	return libdevfile.Deploy(devfileObj, undeployHandler)
}

// DevfileComponentDelete deletes the devfile component
func (do *DeleteComponentClient) DeleteComponent(devfileObj parser.DevfileObj, componentName string) error {
	labels := componentlabels.GetLabels(componentName, "app", false)
	return component.Delete(do.kubeClient, devfileObj, componentName, "app", labels, false, false)
}
