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
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/deploy"
	_init "github.com/redhat-developer/odo/pkg/init"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/watch"
)

const (
	// CATALOG instantiates client for pkg/catalog
	CATALOG = "DEP_CATALOG"
	// DEPLOY instantiates client for pkg/deploy
	DEPLOY = "DEP_DEPLOY"
	// DEV instantiates client for pkg/dev
	DEV = "DEP_DEV"
	// FILESYSTEM instantiates client for pkg/testingutil/filesystem
	FILESYSTEM = "DEP_FILESYSTEM"
	// INIT instantiates client for pkg/init
	INIT = "DEP_INIT"
	// KUBERNETES_NULLABLE instantiates client for pkg/kclient, can be nil
	KUBERNETES_NULLABLE = "DEP_KUBERNETES_NULLABLE"
	// KUBERNETES instantiates client for pkg/kclient
	KUBERNETES = "DEP_KUBERNETES"
	// PREFERENCE instantiates client for pkg/preference
	PREFERENCE = "DEP_PREFERENCE"
	// PROJECT instantiates client for pkg/project
	PROJECT = "DEP_PROJECT"
	// REGISTRY instantiates client for pkg/init/registry
	REGISTRY = "DEP_REGISTRY"
	// WATCH instantiates client for pkg/watch
	WATCH = "DEP_WATCH"

	/* Add key for new package here */
)

// subdeps defines the sub-dependencies
// Clients will be created only once and be reused for sub-dependencies
var subdeps map[string][]string = map[string][]string{
	CATALOG: {FILESYSTEM, PREFERENCE},
	DEPLOY:  {KUBERNETES},
	DEV:     {WATCH},
	INIT:    {FILESYSTEM, PREFERENCE, REGISTRY, CATALOG},
	PROJECT: {KUBERNETES_NULLABLE},
	/* Add sub-dependencies here, if any */
}

type Clientset struct {
	CatalogClient    catalog.Client
	DeployClient     deploy.Client
	DevClient        dev.Client
	FS               filesystem.Filesystem
	InitClient       _init.Client
	KubernetesClient kclient.ClientInterface
	PreferenceClient preference.Client
	ProjectClient    project.Client
	RegistryClient   registry.Client
	WatchClient      watch.Client
	/* Add client here */
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

func Fetch(command *cobra.Command) (*Clientset, error) {
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
	if isDefined(command, PREFERENCE) {
		dep.PreferenceClient, err = preference.NewClient()
		if err != nil {
			return nil, err
		}
	}
	if isDefined(command, REGISTRY) {
		dep.RegistryClient = registry.NewRegistryClient()
	}
	if isDefined(command, WATCH) {
		dep.WatchClient = watch.NewWatchClient()
	}

	/* With sub-dependencies */
	if isDefined(command, CATALOG) {
		dep.CatalogClient = catalog.NewCatalogClient(dep.FS, dep.PreferenceClient)
	}
	if isDefined(command, DEPLOY) {
		dep.DeployClient = deploy.NewDeployClient(dep.KubernetesClient)
	}
	if isDefined(command, INIT) {
		dep.InitClient = _init.NewInitClient(dep.FS, dep.PreferenceClient, dep.RegistryClient, dep.CatalogClient)
	}
	if isDefined(command, PROJECT) {
		dep.ProjectClient = project.NewClient(dep.KubernetesClient)
	}
	if isDefined(command, DEV) {
		dep.DevClient = dev.NewDevClient(dep.WatchClient)
	}
	/* Instantiate new clients here. Take care to instantiate after all sub-dependencies */
	return &dep, nil
}
