package apiserver_impl

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
)

// DefaultAPIService is a service that implements the logic for the DefaultApiServicer
// This service should implement the business logic for every endpoint for the DefaultApi API.
// Include any external packages or services that will be required by this service.
type DefaultAPIService struct {
	cancel context.CancelFunc
}

// NewDefaultApiService creates a default api service
func NewDefaultApiService(cancel context.CancelFunc) openapi.DefaultAPIServicer {
	return &DefaultAPIService{
		cancel: cancel,
	}
}

// ComponentCommandPost -
func (s *DefaultAPIService) ComponentCommandPost(ctx context.Context, componentCommandPostRequest openapi.ComponentCommandPostRequest) (openapi.ImplResponse, error) {
	// TODO - update ComponentCommandPost with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, GeneralSuccess{}) or use other options such as http.Ok ...
	// return Response(200, GeneralSuccess{}), nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("ComponentCommandPost method not implemented")
}

// ComponentGet -
func (s *DefaultAPIService) ComponentGet(ctx context.Context) (openapi.ImplResponse, error) {
	// TODO - update ComponentGet with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, ComponentGet200Response{}) or use other options such as http.Ok ...
	// return Response(200, ComponentGet200Response{}), nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("ComponentGet method not implemented")
}

// InstanceDelete -
func (s *DefaultAPIService) InstanceDelete(ctx context.Context) (openapi.ImplResponse, error) {
	s.cancel()
	return openapi.Response(http.StatusOK, openapi.GeneralSuccess{
		Message: fmt.Sprintf("'odo dev' instance with pid: %d is shuting down.", odocontext.GetPID(ctx)),
	}), nil
}

// InstanceGet -
func (s *DefaultAPIService) InstanceGet(ctx context.Context) (openapi.ImplResponse, error) {
	response := openapi.InstanceGet200Response{
		Pid:                int32(odocontext.GetPID(ctx)),
		ComponentDirectory: odocontext.GetWorkingDirectory(ctx),
	}
	return openapi.Response(http.StatusOK, response), nil
}
