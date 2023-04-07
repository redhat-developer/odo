package podmandev

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/fatih/color"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/port"
	"github.com/redhat-developer/odo/pkg/watch"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog"
)

func (o *DevClient) reconcile(
	ctx context.Context,
	out io.Writer,
	errOut io.Writer,
	options dev.StartOptions,
	componentStatus *watch.ComponentStatus,
) error {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
	)

	o.warnAboutK8sComponents(*devfileObj)

	err := o.buildPushAutoImageComponents(ctx, *devfileObj)
	if err != nil {
		return err
	}

	pod, fwPorts, err := o.deployPod(ctx, options)
	if err != nil {
		return err
	}
	o.deployedPod = pod

	execRequired, err := o.syncFiles(ctx, options, pod, path)
	if err != nil {
		return err
	}

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	if !componentStatus.PostStartEventsDone && libdevfile.HasPostStartEvents(*devfileObj) {
		execHandler := component.NewExecHandler(
			o.podmanClient,
			o.execClient,
			appName,
			componentName,
			pod.Name,
			"Executing post-start command in container",
			false, /* TODO */
		)
		err = libdevfile.ExecPostStartEvents(*devfileObj, execHandler)
		if err != nil {
			return err
		}
	}
	componentStatus.PostStartEventsDone = true

	if execRequired {
		doExecuteBuildCommand := func() error {
			execHandler := component.NewExecHandler(
				o.podmanClient,
				o.execClient,
				appName,
				componentName,
				pod.Name,
				"Building your application in container",
				false, /* TODO */
			)
			return libdevfile.Build(*devfileObj, options.BuildCommand, execHandler)
		}
		err = doExecuteBuildCommand()
		if err != nil {
			return err
		}

		cmdKind := devfilev1.RunCommandGroupKind
		cmdName := options.RunCommand
		if options.Debug {
			cmdKind = devfilev1.DebugCommandGroupKind
			cmdName = options.DebugCommand
		}
		cmdHandler := commandHandler{
			ctx:             ctx,
			fs:              o.fs,
			execClient:      o.execClient,
			platformClient:  o.podmanClient,
			componentExists: componentStatus.RunExecuted,
			podName:         pod.Name,
			appName:         appName,
			componentName:   componentName,
		}
		err = libdevfile.ExecuteCommandByNameAndKind(*devfileObj, cmdName, cmdKind, &cmdHandler, false)
		if err != nil {
			return err
		}
		componentStatus.RunExecuted = true
	}

	// By default, Podman will not forward to container applications listening on the loopback interface.
	// So we are trying to detect such cases and act accordingly.
	// See https://github.com/redhat-developer/odo/issues/6510#issuecomment-1439986558
	err = o.handleLoopbackPorts(options, pod, fwPorts)
	if err != nil {
		return err
	}

	if options.ForwardLocalhost {
		// Port-forwarding is enabled by executing dedicated socat commands
		err = o.portForwardClient.StartPortForwarding(*devfileObj, componentName, options.Debug, options.RandomPorts, out, errOut, fwPorts)
		if err != nil {
			return adapters.NewErrPortForward(err)
		}
	} // else port-forwarding is done via the main container ports in the pod spec

	for _, fwPort := range fwPorts {
		s := fmt.Sprintf("Forwarding from %s:%d -> %d", fwPort.LocalAddress, fwPort.LocalPort, fwPort.ContainerPort)
		fmt.Fprintf(out, " -  %s", log.SboldColor(color.FgGreen, s))
	}
	err = o.stateClient.SetForwardedPorts(fwPorts)
	if err != nil {
		return err
	}

	componentStatus.State = watch.StateReady
	return nil
}

// warnAboutApplyComponents prints a warning if the Devfile contains standalone K8s components (not referenced by any Apply commands). These resources are currently applied when running in the cluster mode, but not on Podman.
func (o *DevClient) warnAboutK8sComponents(devfileObj parser.DevfileObj) {
	var components []string
	// get all standalone k8s components for a given commandGK
	k8sComponents, _ := libdevfile.GetK8sAndOcComponentsToPush(devfileObj, false)

	if len(k8sComponents) == 0 {
		return
	}

	for _, comp := range k8sComponents {
		components = append(components, comp.Name)
	}

	log.Warningf("Kubernetes components are not supported on Podman. Skipping: %v.", strings.Join(components, ", "))
}

func (o *DevClient) buildPushAutoImageComponents(ctx context.Context, devfileObj parser.DevfileObj) error {
	components, err := libdevfile.GetImageComponentsToPushAutomatically(devfileObj)
	if err != nil {
		return err
	}

	for _, c := range components {
		err = image.BuildPushSpecificImage(ctx, o.fs, c, envcontext.GetEnvConfig(ctx).PushImages)
		if err != nil {
			return err
		}
	}
	return nil
}

// deployPod deploys the component as a Pod in podman
func (o *DevClient) deployPod(ctx context.Context, options dev.StartOptions) (*corev1.Pod, []api.ForwardedPort, error) {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetDevfileObj(ctx)
	)

	spinner := log.Spinner("Deploying pod")
	defer spinner.End(false)

	pod, fwPorts, err := createPodFromComponent(
		*devfileObj,
		componentName,
		appName,
		options.Debug,
		options.BuildCommand,
		options.RunCommand,
		options.DebugCommand,
		options.ForwardLocalhost,
		options.RandomPorts,
		o.usedPorts,
	)
	if err != nil {
		return nil, nil, err
	}
	o.usedPorts = getUsedPorts(fwPorts)

	if equality.Semantic.DeepEqual(o.deployedPod, pod) {
		klog.V(4).Info("pod is already deployed as required")
		spinner.End(true)
		return o.deployedPod, fwPorts, nil
	}

	err = o.checkVolumesFree(pod)
	if err != nil {
		return nil, nil, err
	}

	err = o.podmanClient.PlayKube(pod)
	if err != nil {
		// there are cases when pod is created even if there is an error with the pod def; for e.g. incorrect image
		if podMap, _ := o.podmanClient.PodLs(); podMap[pod.Name] {
			o.deployedPod = &corev1.Pod{}
			o.deployedPod.SetName(pod.Name)
		}
		return nil, nil, err
	}

	spinner.End(true)
	return pod, fwPorts, nil
}

// handleLoopbackPorts tries to detect if any of the ports to forward (in fwPorts) is actually bound to the loopback interface within the specified pod.
// If that is the case, it will either return an error if options.IgnoreLocalhost is false, or no error otherwise.
//
// Note that this method should be called after the process representing the application (run or debug command) is actually started in the pod.
func (o *DevClient) handleLoopbackPorts(options dev.StartOptions, pod *corev1.Pod, fwPorts []api.ForwardedPort) error {
	if len(pod.Spec.Containers) == 0 {
		return nil
	}

	loopbackPorts, err := port.DetectRemotePortsBoundOnLoopback(o.execClient, pod.Name, pod.Spec.Containers[0].Name, fwPorts)
	if err != nil {
		return fmt.Errorf("unable to detect container ports bound on the loopback interface: %w", err)
	}

	if len(loopbackPorts) == 0 {
		return nil
	}

	klog.V(5).Infof("detected %d ports bound on the loopback interface in the pod: %v", len(loopbackPorts), loopbackPorts)
	list := make([]string, 0, len(loopbackPorts))
	for _, p := range loopbackPorts {
		list = append(list, fmt.Sprintf("%s (%d)", p.PortName, p.ContainerPort))
	}
	msg := fmt.Sprintf(`Detected that the following port(s) can be reached only via the container loopback interface: %s.
Port forwarding on Podman currently does not work with applications listening on the loopback interface.
Either change the application to make those port(s) reachable on all interfaces (0.0.0.0), or rerun 'odo dev' with `, strings.Join(list, ", "))
	if options.IgnoreLocalhost {
		msg += "'--forward-localhost' to make port-forwarding work with such ports."
	} else {
		msg += `any of the following options:
- --ignore-localhost: no error will be returned by odo, but forwarding to those ports might not work on Podman.
- --forward-localhost: odo will inject a dedicated side container to redirect traffic to such ports.`
	}
	if options.IgnoreLocalhost {
		// ForwardLocalhost should not be true at this point.
		log.Warningf(msg)
	} else if !options.ForwardLocalhost {
		log.Errorf(msg)
		return errors.New("cannot make port forwarding work with ports bound to the loopback interface only")
	}

	return nil
}
