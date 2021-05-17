package occlient

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/library-go/pkg/apps/appsutil"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/kclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

// HasTag checks to see if there is a tag in a list of over tags..
func HasTag(tags []string, requiredTag string) bool {
	for _, tag := range tags {
		if tag == requiredTag {
			return true
		}
	}
	return false
}

// getDeploymentCondition returns the condition with the provided type.
// Borrowed from https://github.com/openshift/origin/blob/64349ed036ed14808124c5b4d8538b3856783b54/pkg/oc/originpolymorphichelpers/deploymentconfigs/status.go
func getDeploymentCondition(status appsv1.DeploymentConfigStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// IsDCRolledOut indicates whether the deployment config is rolled out or not
// Borrowed from https://github.com/openshift/origin/blob/64349ed036ed14808124c5b4d8538b3856783b54/pkg/oc/originpolymorphichelpers/deploymentconfigs/status.go
func IsDCRolledOut(config *appsv1.DeploymentConfig, desiredRevision int64) bool {

	latestRevision := config.Status.LatestVersion

	if latestRevision == 0 {
		switch {
		case appsutil.HasImageChangeTrigger(config):
			klog.V(3).Infof("Deployment config %q waiting on image update", config.Name)
			return false

		case len(config.Spec.Triggers) == 0:
			fmt.Printf("Deployment config %q waiting on manual update (use 'oc rollout latest %s')", config.Name, config.Name)
			return false
		}
	}

	// We use `<` due to OpenShift at times (in rare cases) updating the DeploymentConfig multiple times via ImageTrigger
	if desiredRevision > 0 && latestRevision < desiredRevision {
		klog.V(3).Infof("Desired revision (%d) is different from the running revision (%d)", desiredRevision, latestRevision)
		return false
	}

	// Check the current condition of the deployment config
	cond := getDeploymentCondition(config.Status, appsv1.DeploymentProgressing)
	if config.Generation <= config.Status.ObservedGeneration {
		switch {
		case cond != nil && cond.Reason == "NewReplicationControllerAvailable":
			return true

		case cond != nil && cond.Reason == "ProgressDeadlineExceeded":
			return true

		case cond != nil && cond.Reason == "RolloutCancelled":
			return true

		case cond != nil && cond.Reason == "DeploymentConfigPaused":
			return true

		case config.Status.UpdatedReplicas < config.Spec.Replicas:
			klog.V(3).Infof("Waiting for rollout to finish: %d out of %d new replicas have been updated...", config.Status.UpdatedReplicas, config.Spec.Replicas)
			return false

		case config.Status.Replicas > config.Status.UpdatedReplicas:
			klog.V(3).Infof("Waiting for rollout to finish: %d old replicas are pending termination...", config.Status.Replicas-config.Status.UpdatedReplicas)
			return false

		case config.Status.AvailableReplicas < config.Status.UpdatedReplicas:
			klog.V(3).Infof("Waiting for rollout to finish: %d of %d updated replicas are available...", config.Status.AvailableReplicas, config.Status.UpdatedReplicas)
			return false
		}
	}
	return false
}

// GetS2IEnvForDevfile gets environment variable for builder image to be added in devfiles
func GetS2IEnvForDevfile(sourceType string, env config.EnvVarList, imageStreamImage imagev1.ImageStreamImage) (config.EnvVarList, error) {
	klog.V(2).Info("Get S2I environment variables to be added in devfile")

	s2iPaths, err := getS2IMetaInfoFromBuilderImg(&imageStreamImage)
	if err != nil {
		return nil, err
	}

	inputEnvs, err := kclient.GetInputEnvVarsFromStrings(env.ToStringSlice())
	if err != nil {
		return nil, err
	}
	// Append s2i related parameters extracted above to env
	inputEnvs = injectS2IPaths(inputEnvs, s2iPaths)

	if sourceType == string(config.LOCAL) {
		inputEnvs = uniqueAppendOrOverwriteEnvVars(
			inputEnvs,
			corev1.EnvVar{
				Name:  EnvS2ISrcBackupDir,
				Value: s2iPaths.SrcBackupPath,
			},
		)
	}

	var configEnvs config.EnvVarList

	for _, env := range inputEnvs {
		configEnv := config.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}

		configEnvs = append(configEnvs, configEnv)
	}

	return configEnvs, nil
}

// generateServiceSpec generates the service spec for s2i components
func generateServiceSpec(commonObjectMeta metav1.ObjectMeta, containerPorts []corev1.ContainerPort) corev1.ServiceSpec {
	// generate the Service spec
	var svcPorts []corev1.ServicePort
	for _, containerPort := range containerPorts {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			Protocol:   containerPort.Protocol,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}

	return corev1.ServiceSpec{
		Ports: svcPorts,
		Selector: map[string]string{
			"deploymentconfig": commonObjectMeta.Name,
		},
	}
}
