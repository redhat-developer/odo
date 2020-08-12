package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/sync"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

const (
	kanikoSecret          = "kaniko-secret"
	buildContext          = "build-context"
	buildContextMountPath = "/root/build-context"
	kanikoSecretMountPath = "/root/.docker"
	completionFile        = "/tmp/complete"
	builderServiceAccount = "builder"
	containerName         = "build"
	initContainerName     = "init"
)

var (
	buildContextVolumeMount = corev1.VolumeMount{Name: buildContext, MountPath: buildContextMountPath}
	kanikoSecretVolumeMount = corev1.VolumeMount{Name: kanikoSecret, MountPath: kanikoSecretMountPath}

	secretGroupVersionResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	defaultId                  = int64(0)
)

func (a Adapter) runKaniko(parameters common.BuildParameters, isImageRegistryInternal bool) error {
	if isImageRegistryInternal {
		if err := a.createDockerCfgSecretForInternalRegistry(parameters.EnvSpecificInfo.GetNamespace()); err != nil {
			return errors.Wrap(err, "failed to create dockerconfig secret")
		}
	}

	labels := map[string]string{
		"component": a.ComponentName,
	}

	if err := a.createKanikoBuilderPod(labels, initContainer(initContainerName), builderContainer(containerName, parameters.Tag, isImageRegistryInternal), regcredName); err != nil {
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

	defer func() {
		// This will clean up the builder pod after build is complete
		if err := a.Client.KubeClient.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{}); err != nil {
			log.Errorf("Failed to delete pod '%s': %v", pod.Name, err)
		}

		// This will clean up the created secret after build is complete
		if err := a.Client.KubeClient.CoreV1().Secrets(parameters.EnvSpecificInfo.GetNamespace()).Delete(regcredName, &metav1.DeleteOptions{}); err != nil {
			log.Errorf("Failed to delete secret '%s': %v", regcredName, err)
		}

	}()

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

	// Executing remote command to trigger closing of init container
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

	log.Successf("Started builder pod %s using Kaniko Build strategy", pod.GetName())

	reader, _ := io.Pipe()
	controlC := make(chan os.Signal, 1)

	var cmdOutput string

	go utils.PipeStdOutput(cmdOutput, reader, controlC)

	s := log.Spinner("Waiting for builder pod to complete")

	if _, err := a.Client.WaitAndGetPod(watchOptions, corev1.PodSucceeded, "Waiting for builder pod to complete", false); err != nil {
		s.End(false)
		return errors.Wrapf(err, "unable to build image using Kaniko, error: %s", cmdOutput)
	}

	s.End(true)
	// Stop listening for a ^C so it doesnt perform terminateBuild during any later stages
	signal.Stop(controlC)
	log.Successf("Successfully built container image: %s", parameters.Tag)
	return nil
}

func (a Adapter) createDockerCfgSecretForInternalRegistry(ns string) error {
	// The builder serviceaccount's dockercfg secret as it does not work when it is used
	// as internal registry pull secret.  However, if the secret only contains
	// auth token (no username/password), then it works.  So, what we are doing here
	// is to retrieve the auth token from builder's serviceaccount dockercfg secret.
	// Then, we use the auth token to construct a new docker confg secret to be used as
	// internal registry credentials.
	secret, err := a.getServiceAccountSecret(ns, builderServiceAccount, corev1.SecretTypeDockercfg)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve service account credentials")
	}

	if err := a.createDockerConfigSecretFrom(secret, regcredName); err != nil {
		return errors.Wrap(err, "failed to create docker secret from service account credentials")
	}
	return nil
}

func (a Adapter) getServiceAccountSecret(ns, saName string, secretType corev1.SecretType) (*corev1.Secret, error) {
	sa, err := a.Client.KubeClient.CoreV1().ServiceAccounts(ns).Get(saName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get serviceaccount '%s': %v", saName, err)
	}
	for _, secretRef := range sa.Secrets {
		secret, err := a.Client.KubeClient.CoreV1().Secrets(ns).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve secret '%s': %v", secretRef.Name, err)
		}
		if secret.Type == secretType {
			return secret, nil
		}
	}
	return nil, fmt.Errorf("serviceacccount '%s' does not have '%s' type secret", saName, secretType)
}

type dockerCfg struct {
	Auths map[string]*types.AuthConfig `json:"auths,omitempty"`
}

func (a Adapter) createDockerConfigSecretFrom(source *corev1.Secret, newSecretName string) error {

	token, err := getAuthTokenFromDockerCfgSecret(source)
	if err != nil {
		return errors.Wrap(err, "failed to extract auth token from service account secret")
	}
	outCfg := &dockerCfg{
		Auths: map[string]*types.AuthConfig{
			internalRegistryHost: {Auth: token},
		},
	}
	dockerConfigSecretBytes, err := json.Marshal(&outCfg)
	if err != nil {
		return errors.Wrap(err, "failed to convert created dockerconfig to byte format")
	}

	secretUnstructured, err := utils.CreateSecret(newSecretName, source.GetNamespace(), dockerConfigSecretBytes)
	if err != nil {
		return errors.Wrap(err, "failed to convert created secret to unstructured format")
	}
	if _, err := a.Client.DynamicClient.Resource(secretGroupVersionResource).
		Namespace(source.GetNamespace()).
		Create(secretUnstructured, metav1.CreateOptions{}); err != nil {
		return errors.Wrap(err, "failed to create secret on cluster")
	}
	return nil
}

func (a Adapter) createKanikoBuilderPod(labels map[string]string, init, builder *corev1.Container, secretName string) error {
	objectMeta := kclient.CreateObjectMeta(a.ComponentName+"-build", a.Client.Namespace, labels, nil)
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
							SecretName: secretName,
							Items: []corev1.KeyToPath{
								{
									Key:  corev1.DockerConfigJsonKey,
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

func builderContainer(name, imageTag string, isImageRegistryInternal bool) *corev1.Container {
	commandArgs := []string{"--dockerfile=" + buildContextMountPath + "/Dockerfile",
		"--context=dir://" + buildContextMountPath,
		"--destination=" + imageTag}

	if isImageRegistryInternal {
		commandArgs = append(commandArgs, "--skip-tls-verify")
	}
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

type authCfg struct {
	InternalRegistryAuthToken *token `json:"image-registry.openshift-image-registry.svc:5000"`
}
type token struct {
	Auth string `json:"auth,omitempty"`
}

// NOTE: we assume internal registry host is image-registry.openshift-image-registry.svc:5000
func getAuthTokenFromDockerCfgSecret(source *corev1.Secret) (string, error) {
	data := fmt.Sprintf("%v", source.Data[corev1.DockerConfigKey])

	var bytes []byte
	for _, ps := range strings.Split(strings.Trim(data, "[]"), " ") {
		pi, _ := strconv.Atoi(ps)
		bytes = append(bytes, byte(pi))
	}
	var cfg authCfg
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return "", err
	}
	return cfg.InternalRegistryAuthToken.Auth, nil
}
