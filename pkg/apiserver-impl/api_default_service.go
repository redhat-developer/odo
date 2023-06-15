package apiserver_impl

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/component/describe"
	"github.com/redhat-developer/odo/pkg/kclient"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/state"
)

// DefaultApiService is a service that implements the logic for the DefaultApiServicer
// This service should implement the business logic for every endpoint for the DefaultApi API.
// Include any external packages or services that will be required by this service.
type DefaultApiService struct {
	cancel       context.CancelFunc
	pushWatcher  chan<- struct{}
	kubeClient   kclient.ClientInterface
	podmanClient podman.Client
	stateClient  state.Client
}

// NewDefaultApiService creates a default api service
func NewDefaultApiService(
	cancel context.CancelFunc,
	pushWatcher chan<- struct{},
	kubeClient kclient.ClientInterface,
	podmanClient podman.Client,
	stateClient state.Client,
) openapi.DefaultApiServicer {
	return &DefaultApiService{
		cancel:       cancel,
		pushWatcher:  pushWatcher,
		kubeClient:   kubeClient,
		podmanClient: podmanClient,
		stateClient:  stateClient,
	}
}

// ComponentCommandPost -
func (s *DefaultApiService) ComponentCommandPost(ctx context.Context, componentCommandPostRequest openapi.ComponentCommandPostRequest) (openapi.ImplResponse, error) {
	switch componentCommandPostRequest.Name {
	case "push":
		s.pushWatcher <- struct{}{}
		return openapi.Response(http.StatusOK, nil), nil
	default:
		return openapi.Response(http.StatusBadRequest, nil), fmt.Errorf("command name %q not supported. Supported values are: %q", componentCommandPostRequest.Name, "push")
	}
}

// ComponentGet -
func (s *DefaultApiService) ComponentGet(ctx context.Context) (openapi.ImplResponse, error) {
	value, _, err := describe.DescribeDevfileComponent(ctx, s.kubeClient, s.podmanClient, s.stateClient)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, ""), fmt.Errorf("error getting the description of the component: %w", err)
	}
	return openapi.Response(http.StatusOK, value), nil
}

// InstanceDelete -
func (s *DefaultApiService) InstanceDelete(ctx context.Context) (openapi.ImplResponse, error) {
	s.cancel()
	return openapi.Response(http.StatusOK, openapi.GeneralSuccess{
		Message: fmt.Sprintf("'odo dev' instance with pid: %d is shutting down.", odocontext.GetPID(ctx)),
	}), nil
}

// InstanceGet -
func (s *DefaultApiService) InstanceGet(ctx context.Context) (openapi.ImplResponse, error) {
	response := openapi.InstanceGet200Response{
		Pid:                int32(odocontext.GetPID(ctx)),
		ComponentDirectory: odocontext.GetWorkingDirectory(ctx),
	}
	return openapi.Response(http.StatusOK, response), nil
}
