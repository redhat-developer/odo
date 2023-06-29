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

func (s *DefaultApiService) DevstateContainerContainerNameDelete(ctx context.Context, containerName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteContainer(containerName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the container: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DefaultApiService) DevstateImagePost(ctx context.Context, image openapi.DevstateImagePostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddImage(
		image.Name,
		image.ImageName,
		image.Args,
		image.BuildContext,
		image.RootRequired,
		image.Uri,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the image: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DefaultApiService) DevstateImageImageNameDelete(ctx context.Context, imageName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteImage(imageName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the image: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DefaultApiService) DevstateResourcePost(ctx context.Context, resource openapi.DevstateResourcePostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddResource(
		resource.Name,
		resource.Inlined,
		resource.Uri,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the resource: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil

}

func (s *DefaultApiService) DevstateResourceResourceNameDelete(ctx context.Context, resourceName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteResource(resourceName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the resource: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}
