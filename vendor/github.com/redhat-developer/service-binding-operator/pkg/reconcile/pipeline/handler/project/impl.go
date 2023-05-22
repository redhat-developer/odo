package project

import (
	"errors"
	"fmt"
	"strings"

	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func PreFlightCheck(mandatoryBindingKeys ...string) func(pipeline.Context) {
	return func(ctx pipeline.Context) {
		if ctx.PersistSecret() != nil {
			ctx.SetCondition(apis.Conditions().NotCollectionReady().Build())
			return
		}
		ctx.SetCondition(apis.Conditions().CollectionReady().DataCollected().Build())
		applications, err := ctx.Applications()
		if err != nil {
			ctx.RetryProcessing(err)
			ctx.SetCondition(apis.Conditions().NotInjectionReady().ApplicationNotFound().Msg(err.Error()).Build())
			return
		}
		if len(applications) == 0 {
			ctx.SetCondition(apis.Conditions().NotInjectionReady().Reason(apis.EmptyApplicationReason).Build())
			ctx.StopProcessing()
			return
		}
		items := ctx.BindingItems()
		if len(items) == 0 {
			err := errors.New("no binding data to project")
			ctx.RetryProcessing(err)
			ctx.SetCondition(apis.Conditions().NotInjectionReady().Reason(apis.NoBindingDataReason).Msg(err.Error()).Build())
			return
		}
		if len(mandatoryBindingKeys) > 0 {
			itemMap := items.AsMap()
			for _, bk := range mandatoryBindingKeys {
				if _, found := itemMap[bk]; !found {
					err := fmt.Errorf("Mandatory binding '%v' not found", bk)
					ctx.SetCondition(apis.Conditions().NotInjectionReady().Reason(apis.RequiredBindingNotFound).Msg(err.Error()).Build())
					ctx.Error(err)
					ctx.StopProcessing()
					return
				}
			}
		}
	}
}

func PostFlightCheck(ctx pipeline.Context) {
	ctx.SetCondition(apis.Conditions().InjectionReady().Reason("ApplicationUpdated").Build())
	if ctx.HasLabelSelector() {
		// Force periodic reprocessing of this service binding to catch new workloads
		ctx.DelayReprocessing(nil)
	}
}

func InjectSecretRef(ctx pipeline.Context) {
	applications, _ := ctx.Applications()
	for _, app := range applications {
		// Safety: In the spec API, secretPath is always "", so injecting secrets using this method
		// only happens in the coreos api.
		secretPath := app.SecretPath()
		if secretPath == "" {
			continue
		}

		err := unstructured.SetNestedField(app.Resource().Object, ctx.BindingSecretName(), strings.Split(secretPath, ".")...)
		if err != nil {
			stop(ctx, err)
			return
		}

	}
}

func BindingsAsEnv(ctx pipeline.Context) {
	envBindings := ctx.EnvBindings()
	if ctx.BindAsFiles() && len(envBindings) == 0 {
		return
	}

	secretName := ctx.BindingSecretName()
	var envVars []corev1.EnvVar
	if len(envBindings) > 0 {
		envVars = make([]corev1.EnvVar, 0, len(envBindings))
		for _, e := range envBindings {
			u := corev1.EnvVar{
				Name: e.Var,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: e.Name,
					},
				},
			}
			envVars = append(envVars, u)
		}
	}
	applications, _ := ctx.Applications()
	envFromSecret := corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretName,
			},
		},
	}

	for _, app := range applications {
		containerResources, err := app.BindablePods()
		if err != nil {
			if app.SecretPath() != "" {
				continue
			}
			stop(ctx, err)
			return
		}
		for _, container := range containerResources.Containers {
			if !ctx.BindAsFiles() {
				// Safety: EnvFrom is available when ctx.BindAsFiles() == false
				if err := container.AddEnvFromVar(envFromSecret); err != nil {
					stop(ctx, err)
					return
				}
			}
			if err := container.AddEnvVars(envVars); err != nil {
				stop(ctx, err)
				return
			}
		}
	}
}

func BindingsAsFiles(ctx pipeline.Context) {
	if !ctx.BindAsFiles() {
		return
	}
	secretName := ctx.BindingSecretName()
	bindingName := ctx.BindingName()
	applications, _ := ctx.Applications()
	for _, app := range applications {
		containerResources, err := app.BindablePods()
		if err != nil {
			if app.SecretPath() != "" {
				continue
			}
			stop(ctx, err)
			return
		}

		volume := corev1.Volume{
			Name: bindingName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}
		if err := containerResources.AddVolume(volume); err != nil {
			stop(ctx, err)
			return
		}

		for _, container := range containerResources.Containers {
			mountPath, err := container.MountPath(ctx.BindingName())
			if err != nil {
				stop(ctx, err)
				return
			}
			volumeMount := corev1.VolumeMount{
				Name:      bindingName,
				MountPath: mountPath,
			}

			if err := container.AddVolumeMount(volumeMount); err != nil {
				stop(ctx, err)
				return
			}
		}
	}
}

func Unbind(ctx pipeline.Context) {
	if !ctx.UnbindRequested() {
		return
	}
	applications, err := ctx.Applications()
	if err != nil || len(applications) == 0 {
		ctx.StopProcessing()
		return
	}
	secretName := ctx.BindingSecretName()
	bindingName := ctx.BindingName()
	envVars := ctx.EnvBindings()
	if secretName == "" {
		ctx.StopProcessing()
		return
	}
	for _, app := range applications {
		containerResources, err := app.BindablePods()
		if containerResources == nil && err == nil {
			ctx.StopProcessing()
			return
		}
		if err := containerResources.RemoveVolume(bindingName); err != nil {
			stop(ctx, err)
			return
		}
		for _, container := range containerResources.Containers {
			for _, env := range envVars {
				if err := container.RemoveEnvVars(env.Name); err != nil {
					stop(ctx, err)
					return
				}
			}
			if err := container.RemoveEnvFromVars(secretName); err != nil {
				stop(ctx, err)
				return
			}
			if err := container.RemoveVolumeMount(bindingName); err != nil {
				stop(ctx, err)
				return
			}
		}
	}
	ctx.StopProcessing()
}

func stop(ctx pipeline.Context, err error) {
	ctx.StopProcessing()
	ctx.Error(err)
	ctx.SetCondition(apis.Conditions().NotInjectionReady().Reason("Error").Msg(err.Error()).Build())
}
