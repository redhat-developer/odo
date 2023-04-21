package kubedev

import (
	"context"
	"fmt"
	"io"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/binding"
	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/watch"
)

const (
	promptMessage = `
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
`
)

type DevClient struct {
	kubernetesClient      kclient.ClientInterface
	prefClient            preference.Client
	portForwardClient     portForward.Client
	watchClient           watch.Client
	bindingClient         binding.Client
	syncClient            sync.Client
	filesystem            filesystem.Filesystem
	execClient            exec.Client
	deleteClient          _delete.Client
	configAutomountClient configAutomount.Client

	adapter component.Adapter
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	watchClient watch.Client,
	bindingClient binding.Client,
	syncClient sync.Client,
	filesystem filesystem.Filesystem,
	execClient exec.Client,
	deleteClient _delete.Client,
	configAutomountClient configAutomount.Client,
) *DevClient {
	adapter := component.NewKubernetesAdapter(
		kubernetesClient,
		prefClient,
		portForwardClient,
		bindingClient,
		syncClient,
		execClient,
		configAutomountClient,
		filesystem,
	)

	return &DevClient{
		kubernetesClient:      kubernetesClient,
		prefClient:            prefClient,
		portForwardClient:     portForwardClient,
		watchClient:           watchClient,
		bindingClient:         bindingClient,
		syncClient:            syncClient,
		filesystem:            filesystem,
		execClient:            execClient,
		deleteClient:          deleteClient,
		configAutomountClient: configAutomountClient,
		adapter:               adapter,
	}
}

func (o *DevClient) Start(
	ctx context.Context,
	out io.Writer,
	errOut io.Writer,
	options dev.StartOptions,
) error {
	klog.V(4).Infoln("Creating new adapter")

	var (
		devfileObj = odocontext.GetDevfileObj(ctx)
	)

	pushParameters := adapters.PushParameters{
		IgnoredFiles:         options.IgnorePaths,
		Debug:                options.Debug,
		DevfileBuildCmd:      options.BuildCommand,
		DevfileRunCmd:        options.RunCommand,
		RandomPorts:          options.RandomPorts,
		CustomForwardedPorts: options.CustomForwardedPorts,
		ErrOut:               errOut,
		Devfile:              *devfileObj,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	componentStatus := watch.ComponentStatus{
		ImageComponentsAutoApplied: make(map[string]v1alpha2.ImageComponent),
	}
	err := o.adapter.Push(ctx, pushParameters, &componentStatus)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resources")

	watchParameters := watch.WatchParameters{
		DevfileWatchHandler:  o.regenerateAdapterAndPush,
		FileIgnores:          options.IgnorePaths,
		Debug:                options.Debug,
		DevfileBuildCmd:      options.BuildCommand,
		DevfileRunCmd:        options.RunCommand,
		Variables:            options.Variables,
		RandomPorts:          options.RandomPorts,
		CustomForwardedPorts: options.CustomForwardedPorts,
		WatchFiles:           options.WatchFiles,
		WatchCluster:         true,
		ErrOut:               errOut,
		PromptMessage:        promptMessage,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx, componentStatus)
}

// RegenerateAdapterAndPush get the new devfile and pushes the files to remote pod
func (o *DevClient) regenerateAdapterAndPush(ctx context.Context, pushParams adapters.PushParameters, watchParams watch.WatchParameters, componentStatus *watch.ComponentStatus) error {

	devObj, err := devfile.ParseAndValidateFromFileWithVariables(location.DevfileLocation(""), watchParams.Variables)
	if err != nil {
		return fmt.Errorf("unable to generate component from watch parameters: %w", err)
	}

	pushParams.Devfile = devObj

	err = o.adapter.Push(ctx, pushParams, componentStatus)
	if err != nil {
		return fmt.Errorf("watch command was unable to push component: %w", err)
	}
	return nil
}
