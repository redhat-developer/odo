package component

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	"github.com/openshift/odo/pkg/lclient"
)

// New instantiantes a component adapter
func New(adapterContext common.AdapterContext, client lclient.Client) Adapter {
	return Adapter{
		Client:         client,
		AdapterContext: adapterContext,
	}
}

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	Client lclient.Client
	common.AdapterContext
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
func (a Adapter) Push(path string, ignoredFiles []string, forceBuild bool, globExps []string) (err error) {
	componentExists := utils.ComponentExists(a.Client, a.ComponentName)

	err = a.createOrUpdateComponent(componentExists)
	if err != nil {
		return errors.Wrap(err, "unable to create or update component")
	}

	return nil
}

func (a Adapter) createOrUpdateComponent(componentExists bool) (err error) {
	componentName := a.ComponentName

	labels := map[string]string{
		"component": componentName,
	}

	supportedComponents := adaptersCommon.GetSupportedComponents(a.Devfile.Data)
	if len(supportedComponents) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	if componentExists {
		glog.V(3).Info("The component already exists, attempting to update it")
	} else {
		// Create a docker volume to store the project source code
		a.Client.CreateVolume(a.ComponentName, labels)
		for _, comp := range supportedComponents {
			envVars := utils.ConvertEnvs(comp.Env)

			glog.V(3).Info("Pulling image: %s", *comp.Image)
			// Pull the image (as the Docker daemon requires it to be on the system before starting it)
			err = a.Client.PullImage(*comp.Image)
			if err != nil {
				return errors.Wrapf(err, "Unable to pull %s image", *comp.Image)
			}
			containerLabels := map[string]string{
				"component": componentName,
				"container": *comp.Alias,
			}
			containerConfig := a.Client.GenerateContainerConfig(*comp.Image, comp.Command, comp.Args, envVars, containerLabels)
			hostConfig := container.HostConfig{}
			// If the component set `mountSources` to true, add the source volume to it
			if comp.MountSources {
				mounts := []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: a.ComponentName,
						Target: "/projects",
					},
				}
				hostConfig.Mounts = mounts
			}

			// Create the docker container
			err = a.Client.StartContainer(&containerConfig, &hostConfig, nil)
			if err != nil {
				return err
			}
			glog.V(3).Infof("Successfully created container %s for component %s", *comp.Image, componentName)
		}
		glog.V(3).Infof("Successfully created all containers for component %s", componentName)
	}

	return nil
}
