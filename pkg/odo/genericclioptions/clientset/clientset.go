// Package clientset is used to inject clients inside commands
//
// To use this package:
// From a command definition, use the `Add` function to declare the clients needed by the command
// Then, from the `SetClientset` method of the `Runnable` interface, you can access the clients
//
// To add a new client to this package:
// - add a new constant for the client
// - if the client has sub-dependencies, define a new entry in the map of sub-dependencies
// - add the packages's client to the Clientset structure
// - complete the Fetch function to instantiate the package's client
package clientset

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/dev/kubedev"
	"github.com/redhat-developer/odo/pkg/dev/podmandev"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/logs"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/sync"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/state"

	"github.com/redhat-developer/odo/pkg/binding"
	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/deploy"
	_init "github.com/redhat-developer/odo/pkg/init"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/watch"
)

const (
	// ALIZER instantiates client for pkg/alizer
	ALIZER = "DEP_ALIZER"
	// BINDING instantiates client for pkg/binding
	BINDING = "DEP_BINDING"
	// DELETE_COMPONENT instantiates client for pkg/component/delete
	DELETE_COMPONENT = "DEP_DELETE_COMPONENT"
	// DEPLOY instantiates client for pkg/deploy
	DEPLOY = "DEP_DEPLOY"
	// DEV instantiates client for pkg/dev
	DEV = "DEP_DEV"
	// EXEC instantiates client for pkg/exec
	EXEC = "DEP_EXEC"
	// FILESYSTEM instantiates client for pkg/testingutil/filesystem
	FILESYSTEM = "DEP_FILESYSTEM"
	// INIT instantiates client for pkg/init
	INIT = "DEP_INIT"
	// KUBERNETES_NULLABLE instantiates client for pkg/kclient, can be nil
	KUBERNETES_NULLABLE = "DEP_KUBERNETES_NULLABLE"
	// KUBERNETES instantiates client for pkg/kclient
	KUBERNETES = "DEP_KUBERNETES"
	// LOGS instantiates client for pkg/logs
	LOGS = "DEP_LOGS"
	// PODMAN instantiates client for pkg/podman
	PODMAN = "DEP_PODMAN"
	// PORT_FORWARD instantiates client for pkg/portForward
	PORT_FORWARD = "PORT_FORWARD"
	// PREFERENCE instantiates client for pkg/preference
	PREFERENCE = "DEP_PREFERENCE"
	// PROJECT instantiates client for pkg/project
	PROJECT = "DEP_PROJECT"
	// REGISTRY instantiates client for pkg/registry
	REGISTRY = "DEP_REGISTRY"
	// STATE instantiates client for pkg/state
	STATE = "DEP_STATE"
	// SYNC instantiates client for pkg/sync
	SYNC = "DEP_SYNC"
	// WATCH instantiates client for pkg/watch
	WATCH = "DEP_WATCH"
	/* Add key for new package here */
)

// subdeps defines the sub-dependencies
// Clients will be created only once and be reused for sub-dependencies
var subdeps map[string][]string = map[string][]string{
	ALIZER:           {REGISTRY},
	DELETE_COMPONENT: {KUBERNETES_NULLABLE, EXEC},
	DEPLOY:           {KUBERNETES, FILESYSTEM},
	DEV:              {BINDING, DELETE_COMPONENT, EXEC, FILESYSTEM, KUBERNETES, PODMAN, PORT_FORWARD, PREFERENCE, SYNC, WATCH},
	EXEC:             {KUBERNETES_NULLABLE},
	INIT:             {ALIZER, FILESYSTEM, PREFERENCE, REGISTRY},
	LOGS:             {KUBERNETES_NULLABLE},
	PORT_FORWARD:     {KUBERNETES_NULLABLE, STATE},
	PROJECT:          {KUBERNETES_NULLABLE},
	REGISTRY:         {FILESYSTEM, PREFERENCE},
	STATE:            {FILESYSTEM},
	SYNC:             {EXEC},
	WATCH:            {KUBERNETES_NULLABLE},
	BINDING:          {PROJECT, KUBERNETES_NULLABLE},
	/* Add sub-dependencies here, if any */
}

type Clientset struct {
	AlizerClient      alizer.Client
	BindingClient     binding.Client
	DeleteClient      _delete.Client
	DeployClient      deploy.Client
	DevClient         dev.Client
	ExecClient        exec.Client
	FS                filesystem.Filesystem
	InitClient        _init.Client
	KubernetesClient  kclient.ClientInterface
	LogsClient        logs.Client
	PodmanClient      podman.Client
	PortForwardClient portForward.Client
	PreferenceClient  preference.Client
	ProjectClient     project.Client
	RegistryClient    registry.Client
	StateClient       state.Client
	SyncClient        sync.Client
	WatchClient       watch.Client
	/* Add client by alphabetic order */
}

func Add(command *cobra.Command, dependencies ...string) {
	if command.Annotations == nil {
		command.Annotations = map[string]string{}
	}
	for _, dependency := range dependencies {
		_, ok := command.Annotations[dependency]
		// prevent infinite loop with circular dependencies
		if !ok {
			command.Annotations[dependency] = "true"
			Add(command, subdeps[dependency]...)
		}
	}
}

func isDefined(command *cobra.Command, dependency string) bool {
	_, ok := command.Annotations[dependency]
	return ok
}

func Fetch(command *cobra.Command, platform string) (*Clientset, error) {
	dep := Clientset{}
	var err error

	/* Without sub-dependencies */
	if isDefined(command, FILESYSTEM) {
		dep.FS = filesystem.DefaultFs{}
	}
	if isDefined(command, KUBERNETES) || isDefined(command, KUBERNETES_NULLABLE) {
		dep.KubernetesClient, err = kclient.New()
		if err != nil && isDefined(command, KUBERNETES) {
			return nil, err
		}
	}
	if isDefined(command, PODMAN) {
		dep.PodmanClient = podman.NewPodmanCli()
	}
	if isDefined(command, PREFERENCE) {
		dep.PreferenceClient, err = preference.NewClient(command.Context())
		if err != nil {
			return nil, err
		}
	}
	if isDefined(command, REGISTRY) {
		dep.RegistryClient = registry.NewRegistryClient(dep.FS, dep.PreferenceClient)
	}

	/* With sub-dependencies */
	if isDefined(command, ALIZER) {
		dep.AlizerClient = alizer.NewAlizerClient(dep.RegistryClient)
	}
	if isDefined(command, EXEC) {
		switch platform {
		case commonflags.RunOnCluster:
			dep.ExecClient = exec.NewExecClient(dep.KubernetesClient)
		case commonflags.RunOnPodman:
			dep.ExecClient = exec.NewExecClient(dep.PodmanClient)
		}
	}
	if isDefined(command, DELETE_COMPONENT) {
		dep.DeleteClient = _delete.NewDeleteComponentClient(dep.KubernetesClient, dep.ExecClient)
	}
	if isDefined(command, DEPLOY) {
		dep.DeployClient = deploy.NewDeployClient(dep.KubernetesClient, dep.FS)
	}
	if isDefined(command, INIT) {
		dep.InitClient = _init.NewInitClient(dep.FS, dep.PreferenceClient, dep.RegistryClient, dep.AlizerClient)
	}
	if isDefined(command, LOGS) {
		switch platform {
		case commonflags.RunOnCluster:
			dep.LogsClient = logs.NewLogsClient(dep.KubernetesClient)
		default:
			panic(fmt.Sprintf("not implemented yet for platform %q", platform))
		}
	}
	if isDefined(command, PROJECT) {
		dep.ProjectClient = project.NewClient(dep.KubernetesClient)
	}
	if isDefined(command, STATE) {
		dep.StateClient = state.NewStateClient(dep.FS)
	}
	if isDefined(command, SYNC) {
		switch platform {
		case commonflags.RunOnCluster:
			dep.SyncClient = sync.NewSyncClient(dep.KubernetesClient, dep.ExecClient)
		case commonflags.RunOnPodman:
			dep.SyncClient = sync.NewSyncClient(dep.PodmanClient, dep.ExecClient)
		}
	}
	if isDefined(command, WATCH) {
		dep.WatchClient = watch.NewWatchClient(dep.KubernetesClient)
	}
	if isDefined(command, BINDING) {
		dep.BindingClient = binding.NewBindingClient(dep.ProjectClient, dep.KubernetesClient)
	}
	if isDefined(command, PORT_FORWARD) {
		dep.PortForwardClient = portForward.NewPFClient(dep.KubernetesClient, dep.StateClient)
	}
	if isDefined(command, DEV) {
		switch platform {
		case commonflags.RunOnCluster:
			dep.DevClient = kubedev.NewDevClient(
				dep.KubernetesClient,
				dep.PreferenceClient,
				dep.PortForwardClient,
				dep.WatchClient,
				dep.BindingClient,
				dep.SyncClient,
				dep.FS,
				dep.ExecClient,
				dep.DeleteClient,
			)
		case commonflags.RunOnPodman:
			dep.DevClient = podmandev.NewDevClient(
				dep.PodmanClient,
				dep.SyncClient,
				dep.ExecClient,
			)
		}
	}

	/* Instantiate new clients here. Take care to instantiate after all sub-dependencies */
	return &dep, nil
}
