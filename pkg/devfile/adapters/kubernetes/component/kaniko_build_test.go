package component

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	dynamicfakeclient "k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"

	"encoding/json"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeUnstructured "k8s.io/apimachinery/pkg/runtime"
)

func getTestInitContainer() corev1.Container {

	return corev1.Container{
		Name:            "test-init-container",
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

func getTestBuilderContainer() corev1.Container {

	commandArgs := []string{"--dockerfile=" + buildContextMountPath + "/Dockerfile",
		"--context=dir://" + buildContextMountPath,
		"--destination=" + "test-image-tag"}

	envVars := []corev1.EnvVar{
		{Name: "DOCKER_CONFIG", Value: kanikoSecretMountPath},
		{Name: "AWS_ACCESS_KEY_ID", Value: "NOT_SET"},
		{Name: "AWS_SECRET_KEY", Value: "NOT_SET"},
	}

	return corev1.Container{

		Name:  "test-builder-container",
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
}

type objectMetaFunc func(om *metav1.ObjectMeta)

func TypeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

func ObjectMeta(name, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
}

func SecretObjectMeta(n types.NamespacedName, opts ...objectMetaFunc) metav1.ObjectMeta {
	om := metav1.ObjectMeta{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
	for _, o := range opts {
		o(&om)
	}
	return om
}

func TestGetServiceAccountSecret(t *testing.T) {

	secretString := "test-secret-data"
	secretBytes := []byte(secretString)
	secretData := make(map[string][]byte)
	secretData["secret"] = secretBytes
	testNs := "test-namespace"
	testSaSecretName := "test-sa-secret"
	testSaName := "test-service-account"

	testSa := &corev1.ServiceAccount{
		TypeMeta:   TypeMeta("serviceAccount", "v1"),
		ObjectMeta: ObjectMeta(testSaName, testNs),
		Secrets: []corev1.ObjectReference{
			{
				Kind:       "secret",
				APIVersion: "v1",
				Namespace:  testNs,
				Name:       testSaSecretName,
			},
		},
	}

	testSaSecret := &corev1.Secret{
		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: ObjectMeta(testSaSecretName, testNs),
		Data:       secretData,
		Type:       corev1.SecretTypeDockercfg,
	}

	fkclient, fkclientset := kclient.FakeNew()
	fkclientset.Kubernetes.PrependReactor("get", "serviceaccounts", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, testSa, nil
	})
	fkclientset.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, testSaSecret, nil
	})

	adapterCtx := adaptersCommon.AdapterContext{}
	testAdapter := New(adapterCtx, *fkclient)

	want := testSaSecret
	got, err := testAdapter.getServiceAccountSecret(testNs, testSaName, corev1.SecretTypeDockercfg)

	if err != nil {
		t.Error(err)
		t.Errorf("Error retrieving sa secret")
	}

	diff := cmp.Diff(got, want)
	if diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}

}

func TestCreateDockerConfigSecretFrom(t *testing.T) {
	testConfigJsonString := "{\"auths\":{\"image-registry.openshift-image-registry.svc:5000\":{\"auth\":\"test-auth-token\"}}}"
	testConfigJsonData := []byte(testConfigJsonString)

	testSaSecretString := "{ \"image-registry.openshift-image-registry.svc:5000\": { \"auth\": \"test-auth-token\" } }"
	testSaSecretData := []byte(testSaSecretString)
	testSecretName := "test-secret"
	testSaSecretName := "test-sa-secret"
	testNs := "test-namespace"

	testNamespacedNameSa := types.NamespacedName{
		Name:      testSaSecretName,
		Namespace: testNs,
	}

	testNamespacedName := types.NamespacedName{
		Name:      testSecretName,
		Namespace: testNs,
	}

	testSaSecret := &corev1.Secret{
		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: SecretObjectMeta(testNamespacedNameSa),
		Type:       corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigKey: testSaSecretData,
		},
	}

	testSecret := &corev1.Secret{
		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: SecretObjectMeta(testNamespacedName),
		Type:       corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: testConfigJsonData,
		},
	}

	testSecretData, err := runtimeUnstructured.DefaultUnstructuredConverter.ToUnstructured(testSecret)
	if err != nil {
		t.Error(err)
	}

	testSecretBytes, err := json.Marshal(testSecretData)
	if err != nil {
		t.Errorf("error while marshalling: %v", err)
	}

	var testSecretUnstructured *unstructured.Unstructured
	if err := json.Unmarshal(testSecretBytes, &testSecretUnstructured); err != nil {
		t.Errorf("error unmarshalling into unstructured: %v", err)
	}

	want := testSecretUnstructured
	fkclient, _ := kclient.FakeNew()
	scheme := runtime.NewScheme()
	fkclient.DynamicClient = dynamicfakeclient.NewSimpleDynamicClient(scheme)
	testAdapter := Adapter{
		Client: *fkclient,
	}

	err = testAdapter.createDockerConfigSecretFrom(testSaSecret, testSecretName)
	if err != nil {
		t.Error(err)
		t.Errorf("failed to retrieve dockerconfig secret bytes")
	}

	got, err := testAdapter.Client.DynamicClient.Resource(secretGroupVersionResource).
		Namespace(testNs).
		Get(testSecretName, metav1.GetOptions{})

	diff := cmp.Diff(got, want)
	if diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}
}

func TestCreateKanikoBuilderPod(t *testing.T) {

	testInitContainer := getTestInitContainer()
	testBuiderContainer := getTestBuilderContainer()
	testSecretName := "test-secret"

	labels := map[string]string{
		"component": "test-component",
	}

	fkclient, _ := kclient.FakeNew()
	adapterContext := common.AdapterContext{
		ComponentName: "test-component-name",
	}

	testAdapter := New(adapterContext, *fkclient)

	testAdapter.Client.Namespace = "test-namespace"
	kanikoBuilderPodPorted := corev1.Pod{}

	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-component-name-build",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"component": "test-component",
			},
			Annotations: nil,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: &defaultId,
			},
			InitContainers: []corev1.Container{getTestInitContainer()},
			Containers:     []corev1.Container{getTestBuilderContainer()},
			Volumes: []corev1.Volume{
				{Name: kanikoSecret,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret",
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

	want := testPod

	err := testAdapter.createKanikoBuilderPod(labels, &testInitContainer, &testBuiderContainer, testSecretName)
	if err != nil {
		t.Errorf("failed to deploy pod using fake client")
	}
	got, err := testAdapter.Client.KubeClient.CoreV1().Pods(testAdapter.Client.Namespace).Get("test-component-name-build", metav1.GetOptions{})
	if err != nil {
		t.Errorf("failed to deploy pod using fake client")
	}

	kanikoBuilderPodPorted = *got
	if diff := cmp.Diff(kanikoBuilderPodPorted.ObjectMeta, want.ObjectMeta); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	} else if diff := cmp.Diff(kanikoBuilderPodPorted.Spec.RestartPolicy, want.Spec.RestartPolicy); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	} else if diff := cmp.Diff(kanikoBuilderPodPorted.Spec.SecurityContext, want.Spec.SecurityContext); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	} else if diff := cmp.Diff(kanikoBuilderPodPorted.Spec.InitContainers, want.Spec.InitContainers); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	} else if diff := cmp.Diff(kanikoBuilderPodPorted.Spec.Containers, want.Spec.Containers); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	} else if diff := cmp.Diff(kanikoBuilderPodPorted.Spec.Volumes, want.Spec.Volumes); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}

}

func TestBuilderContainer(t *testing.T) {

	imageTag := "dummy-image-tag"
	containerName := "test-container"
	image := "gcr.io/kaniko-project/executor:latest"

	imagePullPolicy := corev1.PullAlways
	resources := corev1.ResourceRequirements{}
	env := []corev1.EnvVar{
		{Name: "DOCKER_CONFIG", Value: kanikoSecretMountPath},
		{Name: "AWS_ACCESS_KEY_ID", Value: "NOT_SET"},
		{Name: "AWS_SECRET_KEY", Value: "NOT_SET"},
	}
	command := []string{}
	argsExternal := []string{"--dockerfile=" + buildContextMountPath + "/Dockerfile",
		"--context=dir://" + buildContextMountPath,
		"--destination=" + imageTag}

	argsInternal := []string{"--dockerfile=" + buildContextMountPath + "/Dockerfile",
		"--context=dir://" + buildContextMountPath,
		"--destination=" + imageTag,
		"--skip-tls-verify"}
	volumeMounts := []corev1.VolumeMount{
		buildContextVolumeMount,
		kanikoSecretVolumeMount,
	}

	buildContainerPorted := corev1.Container{}

	tests := []struct {
		name                 string
		containerName        string
		testBuilderContainer corev1.Container
		imageTag             string
		isInternalRegistry   bool
	}{
		{
			name:          "Case: Builder container pushes to internal registry",
			containerName: containerName,
			testBuilderContainer: corev1.Container{
				Name:            containerName,
				Image:           image,
				ImagePullPolicy: imagePullPolicy,
				Resources:       resources,
				Env:             env,
				Command:         command,
				Args:            argsInternal,
				VolumeMounts:    volumeMounts,
			},
			imageTag:           imageTag,
			isInternalRegistry: true,
		},

		{
			name:          "Case: Builder container pushes to external registry",
			containerName: containerName,
			testBuilderContainer: corev1.Container{
				Name:            containerName,
				Image:           image,
				ImagePullPolicy: imagePullPolicy,
				Resources:       resources,
				Env:             env,
				Command:         command,
				Args:            argsExternal,
				VolumeMounts:    volumeMounts,
			},
			imageTag:           imageTag,
			isInternalRegistry: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BuilderContainer := builderContainer(tt.containerName, tt.imageTag, tt.isInternalRegistry)
			buildContainerPorted = *BuilderContainer

			diff := cmp.Diff(buildContainerPorted, tt.testBuilderContainer)
			if diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}
		})
	}
}

func TestInitContainer(t *testing.T) {

	testInitContainer := &corev1.Container{
		Name:            "test-container",
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

	initContainerPorted := corev1.Container{}

	want := testInitContainer
	got := initContainer("test-container")
	initContainerPorted = *got

	diff := cmp.Diff(initContainerPorted, *want)
	if diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}
}

func TestGetAuthTokenFromDockerCfgSecret(t *testing.T) {

	testData := "{ \"image-registry.openshift-image-registry.svc:5000\": { \"auth\": \"test-auth-token\" } }"
	testDockerConfigData := []byte(testData)
	testSecretName := "test-secret"
	testNs := "test-namespace"

	testNamespacedName := types.NamespacedName{
		Name:      testSecretName,
		Namespace: testNs,
	}

	testSecret := &corev1.Secret{
		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: SecretObjectMeta(testNamespacedName),
		Type:       corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigKey: testDockerConfigData,
		},
	}

	want := "test-auth-token"

	got, err := getAuthTokenFromDockerCfgSecret(testSecret)
	if err != nil {
		t.Errorf("failed to retrieve auth token: %v", err)
	}
	diff := cmp.Diff(got, want)
	if diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}
}
