package kclient

import (
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// errorMsg is the message for user when invalid configuration error occurs
const errorMsg = `
Please ensure you have an active kubernetes context to your cluster. 
Consult your Kubernetes distribution's documentation for more details
`

// Client is a collection of fields used for client configuration and interaction
type Client struct {
	KubeClient       kubernetes.Interface
	KubeConfig       clientcmd.ClientConfig
	KubeClientConfig *rest.Config
	Namespace        string
}

// New creates a new client
func New() (*Client, error) {
	var client Client
	var err error

	// initialize client-go clients
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	client.KubeClientConfig, err = client.KubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, errorMsg)
	}

	client.KubeClient, err = kubernetes.NewForConfig(client.KubeClientConfig)
	if err != nil {
		return nil, err
	}

	client.Namespace, _, err = client.KubeConfig.Namespace()
	if err != nil {
		return nil, err
	}

	return &client, nil
}

// CreateObjectMeta creates a common object meta
func CreateObjectMeta(name, namespace string, labels, annotations map[string]string) metav1.ObjectMeta {

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	return objectMeta
}

// GetPVCsFromSelector returns the PVCs based on the given selector
func (c *Client) GetPVCsFromSelector(selector string) ([]corev1.PersistentVolumeClaim, error) {
	pvcList, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get PVCs for selector: %v", selector)
	}

	return pvcList.Items, nil
}

// AddPVCAndVolumeMountToPod adds PVC and volume mount to the pod spec
// volumePVCMap is a map of volume name to the PVC created
// containerVolumesMap is a map of the Devfile container alias to the Devfile Volumes
// func AddPVCAndVolumeMountToPod(pod *corev1.Pod, volumePVCMap map[string]*corev1.PersistentVolumeClaim, containerVolumesMap map[string][]devfile.DockerimageVolume) error {
// 	for vol, pvc := range volumePVCMap {
// 		pvcName := pvc.Name
// 		generatedVolumeName := generateVolumeNameFromPVC(pvcName)
// 		AddPVCToPodSpec(pod, pvcName, generatedVolumeName)

// 		// containerMountPathsMap is a map of the Devfile container alias to their Devfile Volume Mount Paths for a given Volume Name
// 		containerMountPathsMap := make(map[string][]string)
// 		for containerName, volumes := range containerVolumesMap {
// 			for _, volume := range volumes {
// 				if vol == *volume.Name {
// 					containerMountPathsMap[containerName] = append(containerMountPathsMap[containerName], *volume.ContainerPath)
// 				}
// 			}
// 		}

// 		err := AddVolumeMountToPodContainerSpec(pod, generatedVolumeName, pvcName, containerMountPathsMap)
// 		if err != nil {
// 			return errors.New("Unable to add volumes mounts to the pod: " + err.Error())
// 		}
// 	}
// 	return nil
// }
