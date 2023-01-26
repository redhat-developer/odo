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

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(devfileObj, containers, buildCommand, runCommand, debugCommand)
	if err != nil {
		return nil, nil, err
	}
	utils.AddOdoProjectVolume(&containers)
	utils.AddOdoMandatoryVolume(&containers)

	// get the endpoint/port information for containers in devfile
	containerComponents, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType},
	})
	if err != nil {
		return nil, nil, err
	}
	ceMapping := libdevfile.GetContainerEndpointMapping(containerComponents, debug)
	fwPorts := addHostPorts(containers, ceMapping, debug, randomPorts, usedPorts)

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

func addHostPorts(containers []corev1.Container, ceMapping map[string][]v1alpha2.Endpoint, debug bool, randomPorts bool, usedPorts []int) []api.ForwardedPort {
	var result []api.ForwardedPort
	startPort := 20001
	endPort := startPort + 10000
	usedPortsCopy := make([]int, len(usedPorts))
	copy(usedPortsCopy, usedPorts)
	for i := range containers {
		var ports []corev1.ContainerPort
		for _, port := range containers[i].Ports {
			containerName := containers[i].Name
			portName := port.Name
			isDebugPort := libdevfile.IsDebugPort(portName)
			if !debug && isDebugPort {
				klog.V(4).Infof("not running in Debug mode, so skipping container Debug port: %v:%v:%v",
					containerName, portName, port.ContainerPort)
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
						if util.IsPortFree(freePort) {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
				}
			} else {
				var err error
				freePort, err = util.NextFreePort(startPort, endPort, usedPorts)
				if err != nil {
					klog.Infof("%s", err)
					continue
				}
				startPort = freePort + 1
			}
			// Find the endpoint in the container-endpoint mapping
			containerPort := int(port.ContainerPort)
			fp := api.ForwardedPort{
				Platform:      commonflags.PlatformPodman,
				PortName:      portName,
				IsDebug:       isDebugPort,
				ContainerName: containerName,
				LocalAddress:  "127.0.0.1",
				LocalPort:     freePort,
				ContainerPort: containerPort,
			}

			for _, ep := range ceMapping[containerName] {
				if ep.TargetPort == containerPort {
					fp.Exposure = string(ep.Exposure)
					break
				}
			}
			result = append(result, fp)
			port.HostPort = int32(freePort)
			ports = append(ports, port)
		}
		containers[i].Ports = ports
	}
	return result
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
