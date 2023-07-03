package describe

import (
	"context"
	"errors"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/state"
)

// DescribeDevfileComponent describes the component defined by the devfile in the current directory
func DescribeDevfileComponent(
	ctx context.Context,
	kubeClient kclient.ClientInterface,
	podmanClient podman.Client,
	stateClient state.Client,
) (result api.Component, devfile *parser.DevfileObj, err error) {
	var (
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		componentName = odocontext.GetComponentName(ctx)
	)

	devfileData, err := api.GetDevfileData(*devfileObj)
	if err != nil {
		return api.Component{}, nil, err
	}

	isPlatformFeatureEnabled := feature.IsEnabled(ctx, feature.GenericPlatformFlag)
	platform := fcontext.GetPlatform(ctx, "")
	switch platform {
	case "":
		if kubeClient == nil {
			log.Warning(kclient.NewNoConnectionError())
		}
		if isPlatformFeatureEnabled && podmanClient == nil {
			log.Warning(podman.NewPodmanNotFoundError(nil))
		}
	case commonflags.PlatformCluster:
		if kubeClient == nil {
			return api.Component{}, nil, kclient.NewNoConnectionError()
		}
		podmanClient = nil
	case commonflags.PlatformPodman:
		if podmanClient == nil {
			return api.Component{}, nil, podman.NewPodmanNotFoundError(nil)
		}
		kubeClient = nil
	}

	// TODO(feloy) Pass PID with `--pid` flag
	allFwdPorts, err := stateClient.GetForwardedPorts(ctx)
	if err != nil {
		return api.Component{}, nil, err
	}
	if isPlatformFeatureEnabled {
		for i := range allFwdPorts {
			if allFwdPorts[i].Platform == "" {
				allFwdPorts[i].Platform = commonflags.PlatformCluster
			}
		}
	}
	var forwardedPorts []api.ForwardedPort
	switch platform {
	case "":
		if isPlatformFeatureEnabled {
			// Read ports from all platforms
			forwardedPorts = allFwdPorts
		} else {
			// Limit to cluster ports only
			for _, p := range allFwdPorts {
				if p.Platform == "" || p.Platform == commonflags.PlatformCluster {
					forwardedPorts = append(forwardedPorts, p)
				}
			}
		}
	case commonflags.PlatformCluster:
		for _, p := range allFwdPorts {
			if p.Platform == "" || p.Platform == commonflags.PlatformCluster {
				forwardedPorts = append(forwardedPorts, p)
			}
		}
	case commonflags.PlatformPodman:
		for _, p := range allFwdPorts {
			if p.Platform == commonflags.PlatformPodman {
				forwardedPorts = append(forwardedPorts, p)
			}
		}
	}

	runningOn, err := GetRunningOn(ctx, componentName, kubeClient, podmanClient)
	if err != nil {
		return api.Component{}, nil, err
	}

	var ingresses []api.ConnectionData
	var routes []api.ConnectionData
	if kubeClient != nil {
		ingresses, routes, err = component.ListRoutesAndIngresses(kubeClient, componentName, odocontext.GetApplication(ctx))
		if err != nil {
			err = clierrors.NewWarning("failed to get ingresses/routes", err)
			// Do not return the error yet, as it is only a warning
		}
	}

	cmp := api.Component{
		DevfilePath:       devfilePath,
		DevfileData:       devfileData,
		DevForwardedPorts: forwardedPorts,
		RunningIn:         api.MergeRunningModes(runningOn),
		RunningOn:         runningOn,
		ManagedBy:         "odo",
		Ingresses:         ingresses,
		Routes:            routes,
	}
	if !isPlatformFeatureEnabled {
		// Display RunningOn field only if the feature is enabled
		cmp.RunningOn = nil
	}
	updateWithRemoteSourceLocation(&cmp)
	return cmp, devfileObj, err
}

// DescribeNamedComponent describes a component given its name
func DescribeNamedComponent(
	ctx context.Context,
	name string,
	kubeClient kclient.ClientInterface,
	podmanClient podman.Client,
) (result api.Component, devfileObj *parser.DevfileObj, err error) {

	isPlatformFeatureEnabled := feature.IsEnabled(ctx, feature.GenericPlatformFlag)
	platform := fcontext.GetPlatform(ctx, "")
	switch platform {
	case "":
		if isPlatformFeatureEnabled {
			//Get info from all platforms
			if kubeClient == nil {
				log.Warning(kclient.NewNoConnectionError())
			}
			if podmanClient == nil {
				log.Warning(podman.NewPodmanNotFoundError(nil))
			}
		} else {
			if kubeClient == nil {
				return api.Component{}, nil, kclient.NewNoConnectionError()
			}
			podmanClient = nil
		}
	case commonflags.PlatformCluster:
		if kubeClient == nil {
			return api.Component{}, nil, kclient.NewNoConnectionError()
		}
		podmanClient = nil
	case commonflags.PlatformPodman:
		if podmanClient == nil {
			return api.Component{}, nil, podman.NewPodmanNotFoundError(nil)
		}
		kubeClient = nil
	}

	runningOn, err := GetRunningOn(ctx, name, kubeClient, podmanClient)
	if err != nil {
		return api.Component{}, nil, err
	}

	devfile, err := component.GetDevfileInfo(ctx, kubeClient, podmanClient, name)
	if err != nil {
		return api.Component{}, nil, err
	}

	var ingresses []api.ConnectionData
	var routes []api.ConnectionData
	if kubeClient != nil {
		ingresses, routes, err = component.ListRoutesAndIngresses(kubeClient, name, odocontext.GetApplication(ctx))
		if err != nil {
			return api.Component{}, nil, fmt.Errorf("failed to get ingresses/routes: %w", err)
		}
	}

	cmp := api.Component{
		DevfileData: &api.DevfileData{
			Devfile: devfile.Data,
		},
		RunningIn: api.MergeRunningModes(runningOn),
		RunningOn: runningOn,
		ManagedBy: "odo",
		Ingresses: ingresses,
		Routes:    routes,
	}
	if !feature.IsEnabled(ctx, feature.GenericPlatformFlag) {
		// Display RunningOn field only if the feature is enabled
		cmp.RunningOn = nil
	}

	return cmp, &devfile, nil
}

func GetRunningOn(ctx context.Context, n string, kubeClient kclient.ClientInterface, podmanClient podman.Client) (map[string]api.RunningModes, error) {
	var runningOn map[string]api.RunningModes
	runningModesMap, err := component.GetRunningModes(ctx, kubeClient, podmanClient, n)
	if err != nil {
		if !errors.As(err, &component.NoComponentFoundError{}) {
			return nil, err
		}
		// it is ok if the component is not deployed
		runningModesMap = nil
	}
	if runningModesMap != nil {
		runningOn = make(map[string]api.RunningModes, len(runningModesMap))
		if kubeClient != nil && runningModesMap[kubeClient] != nil {
			runningOn[commonflags.PlatformCluster] = runningModesMap[kubeClient]
		}
		if podmanClient != nil && runningModesMap[podmanClient] != nil {
			runningOn[commonflags.PlatformPodman] = runningModesMap[podmanClient]
		}
	}
	return runningOn, nil
}

func updateWithRemoteSourceLocation(cmp *api.Component) {
	components, err := cmp.DevfileData.Devfile.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType},
	})
	if err != nil {
		return
	}
	for _, comp := range components {
		if comp.Container.GetMountSources() {
			if comp.Container.SourceMapping == "" {
				comp.Container.SourceMapping = generator.DevfileSourceVolumeMount
				err = cmp.DevfileData.Devfile.UpdateComponent(comp)
				if err != nil {
					klog.V(2).Infof("error occurred while updating the component %s; cause: %s", comp.Name, err)
				}
			}
		}
	}
}
