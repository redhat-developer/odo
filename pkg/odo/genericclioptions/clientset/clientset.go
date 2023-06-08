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
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/dev/kubedev"
	"github.com/redhat-developer/odo/pkg/dev/podmandev"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/logs"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/portForward/kubeportforward"
	"github.com/redhat-developer/odo/pkg/portForward/podmanportforward"
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
	// CONFIG_AUTOMOUNT instantiates client for pkg/configAutomount
	CONFIG_AUTOMOUNT = "DEP_CONFIG_AUTOMOUNT"
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
	// PODMAN_NULLABLE instantiates client for pkg/podman, can be nil
	PODMAN_NULLABLE = "DEP_PODMAN_NULLABLE"
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
	CONFIG_AUTOMOUNT: {KUBERNETES_NULLABLE, PODMAN_NULLABLE},
	DELETE_COMPONENT: {KUBERNETES_NULLABLE, PODMAN_NULLABLE, EXEC, CONFIG_AUTOMOUNT},
	DEPLOY:           {KUBERNETES, FILESYSTEM, CONFIG_AUTOMOUNT},
	DEV: {
		BINDING,
		DELETE_COMPONENT,
		CONFIG_AUTOMOUNT,
		EXEC,
		FILESYSTEM,
		KUBERNETES_NULLABLE,
		PODMAN_NULLABLE,
		PORT_FORWARD,
		PREFERENCE,
		STATE,
		SYNC,
		WATCH,
	},
	EXEC:         {KUBERNETES_NULLABLE, PODMAN_NULLABLE},
	INIT:         {ALIZER, FILESYSTEM, PREFERENCE, REGISTRY},
	LOGS:         {KUBERNETES_NULLABLE, PODMAN_NULLABLE},
	PORT_FORWARD: {KUBERNETES_NULLABLE, EXEC, STATE},
	PROJECT:      {KUBERNETES},
	REGISTRY:     {FILESYSTEM, PREFERENCE, KUBERNETES_NULLABLE},
	STATE:        {FILESYSTEM},
	SYNC:         {EXEC},
	WATCH:        {KUBERNETES_NULLABLE},
	BINDING:      {PROJECT, KUBERNETES_NULLABLE},
	/* Add sub-dependencies here, if any */
}

type Clientset struct {
	AlizerClient          alizer.Client
	BindingClient         binding.Client
	ConfigAutomountClient configAutomount.Client
	DeleteClient          _delete.Client
	DeployClient          deploy.Client
	DevClient             dev.Client
	ExecClient            exec.Client
	FS                    filesystem.Filesystem
	InitClient            _init.Client
	KubernetesClient      kclient.ClientInterface
	LogsClient            logs.Client
	PodmanClient          podman.Client
	PortForwardClient     portForward.Client
	PreferenceClient      preference.Client
	ProjectClient         project.Client
	RegistryClient        registry.Client
	StateClient           state.Client
	SyncClient            sync.Client
	WatchClient           watch.Client
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

func Fetch(command *cobra.Command, platform string, testClientset Clientset) (*Clientset, error) {
	var (
		err error
		dep = Clientset{}
		ctx = command.Context()
	)

	/* Without sub-dependencies */
	if isDefined(command, FILESYSTEM) {
		if testClientset.FS != nil {
			dep.FS = testClientset.FS
		} else {
			dep.FS = filesystem.DefaultFs{}
		}
	}
	if isDefined(command, KUBERNETES) || isDefined(command, KUBERNETES_NULLABLE) {
		dep.KubernetesClient, err = kclient.New()
		if err != nil {
			// only return error is KUBERNETES_NULLABLE is not defined in combination with KUBERNETES
			if isDefined(command, KUBERNETES) && !isDefined(command, KUBERNETES_NULLABLE) {
				return nil, err
			}
			klog.V(3).Infof("no Kubernetes client initialized: %v", err)
			dep.KubernetesClient = nil
		}

	}
	if isDefined(command, PODMAN) || isDefined(command, PODMAN_NULLABLE) {
		dep.PodmanClient, err = podman.NewPodmanCli(ctx)
		if err != nil {
			// send error in case the command is to run on podman platform or if PODMAN clientset is required.
			if isDefined(command, PODMAN) || platform == commonflags.PlatformPodman {
				return nil, podman.NewPodmanNotFoundError(err)
			}
			klog.V(3).Infof("no Podman client initialized: %v", err)
			dep.PodmanClient = nil
		}
	}
	if isDefined(command, PREFERENCE) {
		dep.PreferenceClient, err = preference.NewClient(ctx)
		if err != nil {
			return nil, err
		}
	}
	if isDefined(command, REGISTRY) {
		dep.RegistryClient = registry.NewRegistryClient(dep.FS, dep.PreferenceClient, dep.KubernetesClient)
	}

	/* With sub-dependencies */
	if isDefined(command, ALIZER) {
		if testClientset.AlizerClient != nil {
			dep.AlizerClient = testClientset.AlizerClient
		} else {
			dep.AlizerClient = alizer.NewAlizerClient(dep.RegistryClient)
		}
	}
	if isDefined(command, EXEC) {
		switch platform {
		case commonflags.PlatformPodman:
			dep.ExecClient = exec.NewExecClient(dep.PodmanClient)
		default:
			dep.ExecClient = exec.NewExecClient(dep.KubernetesClient)
		}
	}
	if isDefined(command, CONFIG_AUTOMOUNT) {
		switch platform {
		case commonflags.PlatformPodman:
			dep.ConfigAutomountClient = nil // Not supported
		default:
			dep.ConfigAutomountClient = configAutomount.NewKubernetesClient(dep.KubernetesClient)
		}
	}
	if isDefined(command, DELETE_COMPONENT) {
		dep.DeleteClient = _delete.NewDeleteComponentClient(dep.KubernetesClient, dep.PodmanClient, dep.ExecClient, dep.ConfigAutomountClient)
	}
	if isDefined(command, DEPLOY) {
		dep.DeployClient = deploy.NewDeployClient(dep.KubernetesClient, dep.ConfigAutomountClient, dep.FS)
	}
	if isDefined(command, INIT) {
		dep.InitClient = _init.NewInitClient(dep.FS, dep.PreferenceClient, dep.RegistryClient, dep.AlizerClient)
	}
	if isDefined(command, LOGS) {
		switch platform {
		case commonflags.PlatformPodman:
			dep.LogsClient = logs.NewLogsClient(dep.PodmanClient)
		default:
			dep.LogsClient = logs.NewLogsClient(dep.KubernetesClient)
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
		case commonflags.PlatformPodman:
			dep.SyncClient = sync.NewSyncClient(dep.PodmanClient, dep.ExecClient)
		default:
			dep.SyncClient = sync.NewSyncClient(dep.KubernetesClient, dep.ExecClient)
		}
	}
	if isDefined(command, WATCH) {
		dep.WatchClient = watch.NewWatchClient(dep.KubernetesClient)
	}
	if isDefined(command, BINDING) {
		dep.BindingClient = binding.NewBindingClient(dep.ProjectClient, dep.KubernetesClient)
	}
	if isDefined(command, PORT_FORWARD) {
		switch platform {
		case commonflags.PlatformPodman:
			dep.PortForwardClient = podmanportforward.NewPFClient(dep.ExecClient)
		default:
			dep.PortForwardClient = kubeportforward.NewPFClient(dep.KubernetesClient, dep.StateClient)
		}
	}
	if isDefined(command, DEV) {
		switch platform {
		case commonflags.PlatformPodman:
			dep.DevClient = podmandev.NewDevClient(
				dep.FS,
				dep.PodmanClient,
				dep.PreferenceClient,
				dep.PortForwardClient,
				dep.SyncClient,
				dep.ExecClient,
				dep.StateClient,
				dep.WatchClient,
			)
		default:
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
				dep.ConfigAutomountClient,
			)
		}
	}

	/* Instantiate new clients here. Take care to instantiate after all sub-dependencies */
	return &dep, nil
}
