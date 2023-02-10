package podmandev

import (
	"fmt"
	"math/rand" //#nosec
	"time"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// See https://github.com/devfile/developer-images and https://quay.io/repository/devfile/base-developer-image?tab=tags
const (
	portForwardingHelperContainerName = "odo-helper-port-forwarding"
	portForwardingHelperImage         = "quay.io/devfile/base-developer-image@sha256:27d5ce66a259decb84770ea0d1ce8058a806f39dfcfeed8387f9cf2f29e76480"
)

func createPodFromComponent(
	devfileObj parser.DevfileObj,
	componentName string,
	appName string,
	debug bool,
	buildCommand string,
	runCommand string,
	debugCommand string,
	randomPorts bool,
	usedPorts []int,
) (*corev1.Pod, []api.ForwardedPort, error) {
	containers, err := generator.GetContainers(devfileObj, common.DevfileOptions{})
	if err != nil {
		return nil, nil, err
	}
	if len(containers) == 0 {
		return nil, nil, fmt.Errorf("no valid components found in the devfile")
	}

	fwPorts, err := getPortMapping(devfileObj, debug, randomPorts, usedPorts)
	if err != nil {
		return nil, nil, err
	}

	utils.AddOdoProjectVolume(&containers)
	utils.AddOdoMandatoryVolume(&containers)

	volumes := []corev1.Volume{
		{
			Name: storage.OdoSourceVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: getVolumeName(storage.OdoSourceVolume, componentName, appName),
				},
			},
		},
		{
			Name: storage.SharedDataVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: getVolumeName(storage.SharedDataVolumeName, componentName, appName),
				},
			},
		},
	}

	devfileVolumes, err := storage.ListStorage(devfileObj)
	if err != nil {
		return nil, nil, err
	}

	for _, devfileVolume := range devfileVolumes {
		volumes = append(volumes, corev1.Volume{
			Name: devfileVolume.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: getVolumeName(devfileVolume.Name, componentName, appName),
				},
			},
		})
		err = addVolumeMountToContainer(containers, devfileVolume)
		if err != nil {
			return nil, nil, err
		}
	}

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(devfileObj, containers, buildCommand, runCommand, debugCommand)
	if err != nil {
		return nil, nil, err
	}

	// Remove all containerPorts, as they will be set afterwards in the helper container
	for i := range containers {
		containers[i].Ports = nil
	}
	// Add helper container for port-forwarding
	pfHelperContainer := corev1.Container{
		Name:    portForwardingHelperContainerName,
		Image:   portForwardingHelperImage,
		Command: []string{"tail"},
		Args:    []string{"-f", "/dev/null"},
	}
	for _, fwPort := range fwPorts {
		pfHelperContainer.Ports = append(pfHelperContainer.Ports, corev1.ContainerPort{
			// It is intentional here to use the same port as ContainerPort and HostPort, for simplicity.
			// In the helper container, a process will be run afterwards and will be listening on this port;
			// this process will leverage socat to forward requests to the actual application port.
			Name:          fwPort.PortName,
			ContainerPort: int32(fwPort.LocalPort),
			HostPort:      int32(fwPort.LocalPort),
		})
	}
	containers = append(containers, pfHelperContainer)

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: containers,
			Volumes:    volumes,
		},
	}

	pod.APIVersion, pod.Kind = corev1.SchemeGroupVersion.WithKind("Pod").ToAPIVersionAndKind()
	name, err := util.NamespaceKubernetesObject(componentName, appName)
	if err != nil {
		return nil, nil, err
	}
	pod.SetName(name)

	runtime := component.GetComponentRuntimeFromDevfileMetadata(devfileObj.Data.GetMetadata())
	pod.SetLabels(labels.GetLabels(componentName, appName, runtime, labels.ComponentDevMode, true))
	labels.SetProjectType(pod.GetLabels(), component.GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))

	return &pod, fwPorts, nil
}

func getVolumeName(volume string, componentName string, appName string) string {
	return volume + "-" + componentName + "-" + appName
}

func getPortMapping(devfileObj parser.DevfileObj, debug bool, randomPorts bool, usedPorts []int) ([]api.ForwardedPort, error) {
	containerComponents, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType},
	})
	if err != nil {
		return nil, err
	}
	ceMapping := libdevfile.GetContainerEndpointMapping(containerComponents, debug)

	var existingContainerPorts []int
	for _, endpoints := range ceMapping {
		for _, ep := range endpoints {
			existingContainerPorts = append(existingContainerPorts, ep.TargetPort)
		}
	}

	isPortUsedInContainer := func(p int) bool {
		for _, port := range existingContainerPorts {
			if p == port {
				return true
			}
		}
		return false
	}

	var result []api.ForwardedPort
	startPort := 20001
	endPort := startPort + 10000
	usedPortsCopy := make([]int, len(usedPorts))
	copy(usedPortsCopy, usedPorts)
	for containerName, endpoints := range ceMapping {
	epLoop:
		for _, ep := range endpoints {
			portName := ep.Name
			isDebugPort := libdevfile.IsDebugPort(portName)
			if !debug && isDebugPort {
				klog.V(4).Infof("not running in Debug mode, so skipping Debug endpoint %s (%d) for container %q",
					portName, ep.TargetPort, containerName)
				continue
			}
			var freePort int
			if randomPorts {
				if len(usedPortsCopy) != 0 {
					freePort = usedPortsCopy[0]
					usedPortsCopy = usedPortsCopy[1:]
				} else {
					rand.Seed(time.Now().UnixNano()) //#nosec
					for {
						freePort = rand.Intn(endPort-startPort+1) + startPort //#nosec
						if !isPortUsedInContainer(freePort) && util.IsPortFree(freePort) {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
				}
			} else {
				for {
					freePort, err = util.NextFreePort(startPort, endPort, usedPorts)
					if err != nil {
						klog.Infof("%s", err)
						continue epLoop
					}
					if !isPortUsedInContainer(freePort) {
						break
					}
					startPort = freePort + 1
					time.Sleep(100 * time.Millisecond)
				}
				startPort = freePort + 1
			}
			fp := api.ForwardedPort{
				Platform:      commonflags.PlatformPodman,
				PortName:      portName,
				IsDebug:       isDebugPort,
				ContainerName: containerName,
				LocalAddress:  "127.0.0.1",
				LocalPort:     freePort,
				ContainerPort: ep.TargetPort,
				Exposure:      string(ep.Exposure),
			}
			result = append(result, fp)
		}
	}
	return result, nil
}

func addVolumeMountToContainer(containers []corev1.Container, devfileVolume storage.LocalStorage) error {
	for i := range containers {
		if containers[i].Name == devfileVolume.Container {
			containers[i].VolumeMounts = append(containers[i].VolumeMounts, corev1.VolumeMount{
				Name:      devfileVolume.Name,
				MountPath: devfileVolume.Path,
			})
			return nil
		}
	}
	return fmt.Errorf("container %q not found", devfileVolume.Container)
}

func getUsedPorts(ports []api.ForwardedPort) []int {
	res := make([]int, 0, len(ports))
	for _, port := range ports {
		res = append(res, port.LocalPort)
	}
	return res
}
