package apiserver_impl

import (
	"context"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/apiserver-impl/devstate"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/state"
)

// DefaultApiService is a service that implements the logic for the DefaultApiServicer
// This service should implement the business logic for every endpoint for the DefaultApi API.
// Include any external packages or services that will be required by this service.
type DevstateApiService struct {
	cancel           context.CancelFunc
	pushWatcher      chan<- struct{}
	kubeClient       kclient.ClientInterface
	podmanClient     podman.Client
	stateClient      state.Client
	preferenceClient preference.Client

	devfileState devstate.DevfileState
}

// NewDefaultApiService creates a default api service
func NewDevstateApiService(
	cancel context.CancelFunc,
	pushWatcher chan<- struct{},
	kubeClient kclient.ClientInterface,
	podmanClient podman.Client,
	stateClient state.Client,
	preferenceClient preference.Client,
) openapi.DevstateApiServicer {
	return &DevstateApiService{
		cancel:           cancel,
		pushWatcher:      pushWatcher,
		kubeClient:       kubeClient,
		podmanClient:     podmanClient,
		stateClient:      stateClient,
		preferenceClient: preferenceClient,

		devfileState: devstate.NewDevfileState(),
	}
}
