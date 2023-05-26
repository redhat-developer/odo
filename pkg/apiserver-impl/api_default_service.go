package apiserver_impl

import (
	"context"
	"errors"
	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"net/http"
)

// DefaultApiService is a service that implements the logic for the DefaultApiServicer
// This service should implement the business logic for every endpoint for the DefaultApi API.
// Include any external packages or services that will be required by this service.
type DefaultApiService struct {
}

// NewDefaultApiService creates a default api service
func NewDefaultApiService() openapi.DefaultApiServicer {
	return &DefaultApiService{}
}

// ComponentCommandPost -
func (s *DefaultApiService) ComponentCommandPost(ctx context.Context, componentCommandPostRequest openapi.ComponentCommandPostRequest) (openapi.ImplResponse, error) {
	// TODO - update ComponentCommandPost with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, GeneralSuccess{}) or use other options such as http.Ok ...
	//return Response(200, GeneralSuccess{}), nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("ComponentCommandPost method not implemented")
}

// ComponentGet -
func (s *DefaultApiService) ComponentGet(ctx context.Context) (openapi.ImplResponse, error) {
	// TODO - update ComponentGet with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, ComponentGet200Response{}) or use other options such as http.Ok ...
	//return Response(200, ComponentGet200Response{}), nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("ComponentGet method not implemented")
}

// InstanceDelete -
func (s *DefaultApiService) InstanceDelete(ctx context.Context) (openapi.ImplResponse, error) {
	// TODO - update InstanceDelete with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, GeneralSuccess{}) or use other options such as http.Ok ...
	//return Response(200, GeneralSuccess{}), nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("InstanceDelete method not implemented")
}

// InstanceGet -
func (s *DefaultApiService) InstanceGet(ctx context.Context) (openapi.ImplResponse, error) {
	// TODO - update InstanceGet with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, InstanceGet200Response{}) or use other options such as http.Ok ...
	//return Response(200, InstanceGet200Response{}), nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("InstanceGet method not implemented")
}
