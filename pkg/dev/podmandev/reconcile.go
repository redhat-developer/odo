package podmandev

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/fatih/color"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
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
			"",
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
			execClient:      o.execClient,
			platformClient:  o.podmanClient,
			componentExists: true, // TODO
			podName:         pod.Name,
			appName:         appName,
			componentName:   componentName,
		}
		err = libdevfile.ExecuteCommandByNameAndKind(*devfileObj, cmdName, cmdKind, &cmdHandler, false)
		if err != nil {
			return err
		}
	}

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
	k8sComponents, _ := devfile.GetKubernetesComponentsToPush(devfileObj, false)

	if len(k8sComponents) == 0 {
		return
	}

	for _, comp := range k8sComponents {
		components = append(components, comp.Name)
	}

	log.Warningf("Kubernetes components are not supported on Podman. Skipping: %v.", strings.Join(components, ", "))
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
		options.BuildCommand,
		options.RunCommand,
		"",
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
		return nil, nil, err
	}

	spinner.End(true)
	return pod, fwPorts, nil
}
