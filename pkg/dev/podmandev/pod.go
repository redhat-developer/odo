package podmandev

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

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

	fwPorts := addHostPorts(containers, debug, usedPorts)

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

func addHostPorts(containers []corev1.Container, debug bool, usedPorts []int) []api.ForwardedPort {
	var result []api.ForwardedPort
	startPort := 40001
	endPort := startPort + 10000
	for i := range containers {
		var ports []corev1.ContainerPort
		for _, port := range containers[i].Ports {
			portName := port.Name
			if !debug && libdevfile.IsDebugPort(portName) {
				klog.V(4).Infof("not running in Debug mode, so skipping container Debug port: %v:%v:%v",
					containers[i].Name, portName, port.ContainerPort)
				continue
			}
			freePort, err := util.NextFreePort(startPort, endPort, usedPorts)
			if err != nil {
				klog.Infof("%s", err)
				continue
			}
			result = append(result, api.ForwardedPort{
				Platform:      commonflags.PlatformPodman,
				ContainerName: containers[i].Name,
				LocalAddress:  "127.0.0.1",
				LocalPort:     freePort,
				ContainerPort: int(port.ContainerPort),
			})
			port.HostPort = int32(freePort)
			ports = append(ports, port)
			startPort = freePort + 1
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
