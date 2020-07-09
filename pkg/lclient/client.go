package lclient

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	volumeTypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

const errorMsg = `
Please ensure that Docker is currently running on your machine.
`

// MinDockerAPIVersion is the minimum Docker API version to use
// 1.30 corresponds to Docker 17.05, which should be sufficiently old enough
const (
	MinDockerAPIVersion     = "1.30" // MinDockerAPIVersion is the minimum Docker API version to use 1.30 corresponds to Docker 17.05, which should be sufficiently old enough to support most systems
	DockerStorageDriver     = ""
	OdoSourceVolumeMount    = "/projects"
	ProjectSourceVolumeName = "odo-project-source"
)

// DockerClient requires functions called on the docker client package
// By abstracting these functions into an interface, it makes creating mock clients for unit testing much easier
type DockerClient interface {
	ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error)
	ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error)
	VolumeCreate(ctx context.Context, options volumeTypes.VolumeCreateBody) (types.Volume, error)
	VolumeList(ctx context.Context, filter filters.Args) (volumeTypes.VolumeListOKBody, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
}

// Client is a collection of fields used for client configuration and interaction
type Client struct {
	Context context.Context
	Client  DockerClient
}

// New creates a new instances of Docker client, with the minimum API version set to
//   to the value of MinDockerAPIVersion.
func New() (*Client, error) {
	// Create the context and client variables for docker
	ctx := context.Background()

	// Create a new Docker client instance
	client, err := client.NewClientWithOpts(client.WithVersion(MinDockerAPIVersion))
	if err != nil {
		// Unable to create a Docker client likely means that Docker isn't running on the user's system.
		return nil, errors.Wrapf(err, errorMsg)
	}

	dockerClient := Client{
		Context: ctx,
		Client:  client,
	}

	return &dockerClient, nil
}
