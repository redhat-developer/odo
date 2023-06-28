package apiserver_impl

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
)

func (s *DefaultApiService) DevstateContainerPost(ctx context.Context, container openapi.DevstateContainerPostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddContainer(
		container.Name,
		container.Image,
		container.Command,
		container.Args,
		container.MemReq,
		container.MemLimit,
		container.CpuReq,
		container.CpuLimit,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the container: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}
