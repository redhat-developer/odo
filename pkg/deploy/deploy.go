package deploy

import (
	"context"
	"path/filepath"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type DeployClient struct {
	kubeClient            kclient.ClientInterface
	configAutomountClient configAutomount.Client
	fs                    filesystem.Filesystem
}

var _ Client = (*DeployClient)(nil)

func NewDeployClient(kubeClient kclient.ClientInterface, configAutomountClient configAutomount.Client, fs filesystem.Filesystem) *DeployClient {
	return &DeployClient{
		kubeClient:            kubeClient,
		configAutomountClient: configAutomountClient,
		fs:                    fs,
	}
}

func (o *DeployClient) Deploy(ctx context.Context) error {
	var (
		devfileObj  = odocontext.GetEffectiveDevfileObj(ctx)
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
	)

	_, err := libdevfile.ValidateAndGetCommand(*devfileObj, "", v1alpha2.DeployCommandGroupKind)
	if err != nil {
		return err
	}

	handler := component.NewRunHandler(
		ctx,
		o.kubeClient,
		nil,
		o.configAutomountClient,
		o.fs,
		image.SelectBackend(ctx),
		component.HandlerOptions{
			Devfile: *devfileObj,
			Path:    path,
		},
	)

	err = o.buildPushAutoImageComponents(handler, *devfileObj)
	if err != nil {
		return err
	}

	err = o.applyAutoK8sOrOcComponents(handler, *devfileObj)
	if err != nil {
		return err
	}

	return libdevfile.Deploy(ctx, *devfileObj, handler)
}

func (o *DeployClient) buildPushAutoImageComponents(handler libdevfile.Handler, devfileObj parser.DevfileObj) error {
	components, err := libdevfile.GetImageComponentsToPushAutomatically(devfileObj)
	if err != nil {
		return err
	}

	for _, c := range components {
		err = handler.ApplyImage(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *DeployClient) applyAutoK8sOrOcComponents(handler libdevfile.Handler, devfileObj parser.DevfileObj) error {
	components, err := libdevfile.GetK8sAndOcComponentsToPush(devfileObj, false)
	if err != nil {
		return err
	}

	for _, c := range components {
		var f func(component2 v1alpha2.Component, kind v1alpha2.CommandGroupKind) error
		if c.Kubernetes != nil {
			f = handler.ApplyKubernetes
		} else if c.Openshift != nil {
			f = handler.ApplyOpenShift
		}
		if f == nil {
			continue
		}
		if err = f(c, v1alpha2.DeployCommandGroupKind); err != nil {
			return err
		}
	}
	return nil
}
