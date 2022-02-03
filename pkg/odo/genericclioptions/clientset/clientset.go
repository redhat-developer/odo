// clientset package is used to inject clients inside commands
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

	"github.com/redhat-developer/odo/pkg/application"
	_init "github.com/redhat-developer/odo/pkg/init"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

const (
	// pkg/application
	APPLICATION = "DEP_APPLICATION"
	// pkg/testingutil/filesystem
	FILESYSTEM = "DEP_FILESYSTEM"
	// pkg/init
	INIT = "DEP_INIT"
	// pkg/kclient, can be nil
	KUBERNETES_NULLABLE = "DEP_KUBERNETES_NULLABLE"
	// pkg/kclient
	KUBERNETES = "DEP_KUBERNETES"
	// pkg/preference
	PREFERENCE = "DEP_PREFERENCE"
	// pkg/project
	PROJECT = "DEP_PROJECT"
	// pkg/init/registry
	REGISTRY = "DEP_REGISTRY"

	/* Add key for new package here */
)

// subdeps defines the sub-dependencies
// Clients will be created only once and be reused for sub-dependencies
var subdeps map[string][]string = map[string][]string{
	APPLICATION: {KUBERNETES},
	INIT:        {FILESYSTEM, PREFERENCE, REGISTRY},
	PROJECT:     {KUBERNETES},
	/* Add sub-dependencies here, if any */
}

type Clientset struct {
	ApplicationClient application.Client
	FS                filesystem.Filesystem
	InitClient        _init.Client
	KubernetesClient  kclient.ClientInterface
	PreferenceClient  preference.Client
	ProjectClient     project.Client
	RegistryClient    registry.Client
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

	/* With sub-dependencies */
	if isDefined(command, APPLICATION) {
		dep.ApplicationClient = application.NewClient(dep.KubernetesClient)
	}
	if isDefined(command, INIT) {
		dep.InitClient = _init.NewInitClient(dep.FS, dep.PreferenceClient, dep.RegistryClient)
	}
	if isDefined(command, PROJECT) {
		dep.ProjectClient = project.NewClient(dep.KubernetesClient)
	}

	/* Instantiate new clients here. Take care to instantiate after all sub-dependencies */
	return &dep, nil
}
