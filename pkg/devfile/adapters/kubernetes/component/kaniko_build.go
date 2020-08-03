package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/secret"
	"github.com/openshift/odo/pkg/sync"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeUnstructured "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

const (
	regcredName           = "regcred"
	kanikoSecret          = "kaniko-secret"
	buildContext          = "build-context"
	buildContextMountPath = "/root/build-context"
	kanikoSecretMountPath = "/root/.docker"
	completionFile        = "/tmp/complete"
)

var (
	buildContextVolumeMount = corev1.VolumeMount{Name: buildContext, MountPath: buildContextMountPath}
	kanikoSecretVolumeMount = corev1.VolumeMount{Name: kanikoSecret, MountPath: kanikoSecretMountPath}

	secretGroupVersionResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	defaultId                  = int64(0)
)

func (a Adapter) runKaniko(parameters common.BuildParameters) error {
	if err := a.createSecret(parameters.EnvSpecificInfo.GetNamespace(), parameters.DockerConfigJSONFilename); err != nil {
		return err
	}
	containerName := "build"
	initContainerName := "init"
	labels := map[string]string{
		"component": a.ComponentName,
	}

	if err := a.createKanikoBuilderPod(labels, initContainer(initContainerName), builderContainer(containerName, parameters.Tag)); err != nil {
		return errors.Wrap(err, "error while creating kaniko builder pod")
	}

	podSelector := fmt.Sprintf("component=%s", a.ComponentName)
	watchOptions := metav1.ListOptions{
		LabelSelector: podSelector,
	}
	// Wait for Pod to be in running state otherwise we can't sync data to it.
	pod, err := a.Client.WaitAndGetPodOnInitContainerStarted(watchOptions, initContainerName, "Waiting for component to start", false)
	if err != nil {
		return errors.Wrapf(err, "error while waiting for pod %s", podSelector)
	}

	// Sync files to volume
	log.Infof("\nSyncing to component %s", a.ComponentName)
	// Get a sync adapter. Check if project files have changed and sync accordingly
	syncAdapter := sync.New(a.AdapterContext, &a.Client)
	compInfo := common.ComponentInfo{
		ContainerName: initContainerName,
		PodName:       pod.GetName(),
	}

	syncFolder, err := syncAdapter.SyncFilesBuild(parameters, dockerfilePath)

	if err != nil {
		return errors.Wrapf(err, "failed to sync to component with name %s", a.ComponentName)
	}

	klog.V(4).Infof("Copying files to pod")
	if err := a.Client.ExtractProjectToComponent(compInfo, buildContextMountPath, syncFolder); err != nil {
		return errors.Wrapf(err, "failed to stream tarball into file transfer container")
	}

	cmd := []string{"touch", completionFile}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := a.Client.ExecCMDInContainer(compInfo, cmd, &stdout, &stderr, nil, false); err != nil {
		log.Errorf("Command '%s' in container failed.\n", strings.Join(cmd, " "))
		log.Errorf("stdout: %s\n", stdout.String())
		log.Errorf("stderr: %s\n", stderr.String())
		log.Errorf("err: %s\n", err.Error())
		return err
	}
	return errors.New("WIP: Need to redirect log output the stdout and wait for build to complete")
}

func (a Adapter) createKanikoBuilderPod(labels map[string]string, init, builder *corev1.Container) error {
	objectMeta := kclient.CreateObjectMeta(a.ComponentName, a.Client.Namespace, labels, nil)
	pod := &corev1.Pod{
		ObjectMeta: objectMeta,
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: &defaultId,
			},
			InitContainers: []corev1.Container{*init},
			Containers:     []corev1.Container{*builder},
			Volumes: []corev1.Volume{
				{Name: kanikoSecret,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: regcredName,
							Items: []corev1.KeyToPath{
								{
									Key:  ".dockerconfigjson",
									Path: "config.json",
								},
							},
						},
					},
				},
				{Name: buildContext,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	klog.V(3).Infof("Creating build pod %v", pod.GetName())
	p, err := a.Client.KubeClient.CoreV1().Pods(a.Client.Namespace).Create(pod)
	if err != nil {
		return err
	}
	klog.V(5).Infof("Successfully created pod %v", p.GetName())
	return nil
}

func builderContainer(name, imageTag string) *corev1.Container {
	commandArgs := []string{"--dockerfile=" + buildContextMountPath + "/Dockerfile",
		"--context=dir://" + buildContextMountPath,
		"--destination=" + imageTag}
	envVars := []corev1.EnvVar{
		{Name: "DOCKER_CONFIG", Value: kanikoSecretMountPath},
		{Name: "AWS_ACCESS_KEY_ID", Value: "NOT_SET"},
		{Name: "AWS_SECRET_KEY", Value: "NOT_SET"},
	}
	container := &corev1.Container{
		Name:  name,
		Image: "gcr.io/kaniko-project/executor:latest",

		ImagePullPolicy: corev1.PullAlways,
		Resources:       corev1.ResourceRequirements{},
		Env:             envVars,
		Command:         []string{},
		Args:            commandArgs,
		VolumeMounts: []corev1.VolumeMount{
			buildContextVolumeMount,
			kanikoSecretVolumeMount,
		},
	}
	return container
}

func initContainer(name string) *corev1.Container {
	return &corev1.Container{
		Name:            name,
		Image:           "busybox",
		ImagePullPolicy: corev1.PullAlways,
		Resources:       corev1.ResourceRequirements{},
		Env:             []corev1.EnvVar{},
		Command:         []string{"/bin/sh", "-c"},
		Args:            []string{"while true; do sleep 1; if [ -f " + completionFile + " ]; then break; fi done"},
		VolumeMounts: []corev1.VolumeMount{
			buildContextVolumeMount,
		},
	}
}

func (a Adapter) createSecret(ns, dcokerConfigFile string) error {
	filename, err := homedir.Expand(dcokerConfigFile)
	if err != nil {
		return fmt.Errorf("failed to generate path to file for %s: %v", dcokerConfigFile, err)
	}

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to read Docker config %#v : %s", filename, err)
	}
	defer f.Close()

	secret, err := secret.CreateDockerConfigSecret(types.NamespacedName{
		Name:      regcredName,
		Namespace: ns,
	}, f)
	if err != nil {
		return err
	}

	secretData, err := runtimeUnstructured.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return err
	}

	secretBytes, err := json.Marshal(secretData)
	if err != nil {
		return err
	}

	var secretUnstructured *unstructured.Unstructured
	if err := json.Unmarshal(secretBytes, &secretUnstructured); err != nil {
		return err
	}

	if _, err = a.Client.DynamicClient.Resource(secretGroupVersionResource).
		Namespace(ns).
		Create(secretUnstructured, metav1.CreateOptions{}); err != nil {
		if errors.Cause(err).Error() != "secrets \""+regcredName+"\" already exists" {
			return err
		}
	}
	return nil
}
