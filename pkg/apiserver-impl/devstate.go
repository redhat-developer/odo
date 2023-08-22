package apiserver_impl

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/apiserver-impl/devstate"
)

func (s *DevstateApiService) DevstateContainerPost(ctx context.Context, container openapi.DevstateContainerPostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddContainer(
		container.Name,
		container.Image,
		container.Command,
		container.Args,
		container.Env,
		container.MemReq,
		container.MemLimit,
		container.CpuReq,
		container.CpuLimit,
		container.VolumeMounts,
		container.ConfigureSources,
		container.MountSources,
		container.SourceMapping,
		container.Annotation,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the container: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateContainerContainerNameDelete(ctx context.Context, containerName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteContainer(containerName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the container: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateImagePost(ctx context.Context, image openapi.DevstateImagePostRequest) (openapi.ImplResponse, error) {
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

func (s *DevstateApiService) DevstateImageImageNameDelete(ctx context.Context, imageName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteImage(imageName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the image: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateResourcePost(ctx context.Context, resource openapi.DevstateResourcePostRequest) (openapi.ImplResponse, error) {
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

func (s *DevstateApiService) DevstateResourceResourceNameDelete(ctx context.Context, resourceName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteResource(resourceName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the resource: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateVolumePost(ctx context.Context, volume openapi.DevstateVolumePostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddVolume(
		volume.Name,
		volume.Ephemeral,
		volume.Size,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the volume: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateVolumeVolumeNameDelete(ctx context.Context, volumeName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteVolume(volumeName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the volume: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateApplyCommandPost(ctx context.Context, command openapi.DevstateApplyCommandPostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddApplyCommand(
		command.Name,
		command.Component,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the Apply command: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateCommandCommandNameDelete(ctx context.Context, commandName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.DeleteCommand(commandName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error deleting the command: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateCompositeCommandPost(ctx context.Context, command openapi.DevstateCompositeCommandPostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddCompositeCommand(
		command.Name,
		command.Parallel,
		command.Commands,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the Composite command: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil

}

func (s *DevstateApiService) DevstateExecCommandPost(ctx context.Context, command openapi.DevstateExecCommandPostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.AddExecCommand(
		command.Name,
		command.Component,
		command.CommandLine,
		command.WorkingDir,
		command.HotReloadCapable,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error adding the Exec command: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateMetadataPut(ctx context.Context, metadata openapi.MetadataRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.SetMetadata(
		metadata.Name,
		metadata.Version,
		metadata.DisplayName,
		metadata.Description,
		metadata.Tags,
		metadata.Architectures,
		metadata.Icon,
		metadata.GlobalMemoryLimit,
		metadata.ProjectType,
		metadata.Language,
		metadata.Website,
		metadata.Provider,
		metadata.SupportUrl,
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error updating the metadata: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateChartGet(context.Context) (openapi.ImplResponse, error) {
	chart, err := s.devfileState.GetFlowChart()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error building the Devfile cycle chart: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, openapi.DevstateChartGet200Response{
		Chart: chart,
	}), nil
}

func (s *DevstateApiService) DevstateCommandCommandNameMovePost(ctx context.Context, commandName string, params openapi.DevstateCommandCommandNameMovePostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.MoveCommand(
		params.FromGroup,
		params.ToGroup,
		int(params.FromIndex),
		int(params.ToIndex),
	)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error moving command to group %q index %d: %s", params.ToGroup, params.ToIndex, err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateCommandCommandNameSetDefaultPost(ctx context.Context, commandName string, params openapi.DevstateCommandCommandNameSetDefaultPostRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.SetDefaultCommand(commandName, params.Group)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error setting command %q as default for group %q: %s", commandName, params.Group, err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateCommandCommandNameUnsetDefaultPost(ctx context.Context, commandName string) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.UnsetDefaultCommand(commandName)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error unsetting command %q as default: %s", commandName, err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateEventsPut(ctx context.Context, params openapi.DevstateEventsPutRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.UpdateEvents(params.EventName, params.Commands)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error updating commands for event %q: %s", params.EventName, err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateQuantityValidPost(ctx context.Context, params openapi.DevstateQuantityValidPostRequest) (openapi.ImplResponse, error) {
	result := devstate.IsQuantityValid(params.Quantity)
	if !result {
		return openapi.Response(http.StatusBadRequest, openapi.GeneralError{
			Message: fmt.Sprintf("Quantity %q is not valid", params.Quantity),
		}), nil
	}
	return openapi.Response(http.StatusOK, openapi.GeneralSuccess{
		Message: fmt.Sprintf("Quantity %q is valid", params.Quantity),
	}), nil
}

func (s *DevstateApiService) DevstateDevfilePut(ctx context.Context, params openapi.DevstateDevfilePutRequest) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.SetDevfileContent(params.Content)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error setting new Devfile content: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateDevfileGet(context.Context) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.GetContent()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error getting new Devfile content: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}

func (s *DevstateApiService) DevstateDevfileDelete(context.Context) (openapi.ImplResponse, error) {
	newContent, err := s.devfileState.SetDevfileContent(`schemaVersion: 2.2.0`)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, openapi.GeneralError{
			Message: fmt.Sprintf("Error clearing Devfile content: %s", err),
		}), nil
	}
	return openapi.Response(http.StatusOK, newContent), nil
}
