package apiserver_impl

import (
	"context"
	"net/http"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
)

func (s *DefaultApiService) DevstateContainerPost(context.Context, openapi.DevstateContainerPostRequest) (openapi.ImplResponse, error) {
	return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
		Message: "Not implemented",
	}), nil
}
