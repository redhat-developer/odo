package project

import (
	"errors"
	"fmt"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"path"
	"reflect"
	"strings"
)

func PreFlightCheck(mandatoryBindingKeys ...string) func(pipeline.Context) {
	return func(ctx pipeline.Context) {
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
}

func InjectSecretRef(ctx pipeline.Context) {
	applications, _ := ctx.Applications()
	for _, app := range applications {
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
	var envVars []interface{}
	if len(envBindings) > 0 {
		envVars = make([]interface{}, 0, len(envBindings))
		for _, e := range envBindings {
			u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.EnvVar{
				Name: e.Var,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: e.Name,
					},
				},
			})
			if err != nil {
				stop(ctx, err)
				return
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
		if app.SecretPath() != "" {
			continue
		}
		containerResources, err := app.BindableContainers()
		if containerResources == nil && err == nil {
			err = errors.New("Containers not found in app resource")
		}
		if err != nil {
			stop(ctx, err)
			return
		}
		for _, container := range containerResources {
			if !ctx.BindAsFiles() {
				envFrom, found := container["envFrom"]
				if !found {

					u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&envFromSecret)
					if err != nil {
						stop(ctx, err)
						return
					}
					container["envFrom"] = []interface{}{u}
					continue
				}
				envFromSlice, ok := envFrom.([]interface{})
				if !ok {
					stop(ctx, errors.New("envFrom not a slice"))
					return
				}
				u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&envFromSecret)
				if err != nil {
					stop(ctx, err)
					return
				}
				container["envFrom"] = append(envFromSlice, u)
				continue
			}
			env, found := container["env"]
			if !found {
				container["env"] = envVars
				continue
			}
			envSlice, ok := env.([]interface{})
			if !ok {
				stop(ctx, errors.New("env not a slice"))
				return
			}
			container["env"] = append(envSlice, envVars...)
		}
	}
}

var volumesPath = []string{"spec", "template", "spec", "volumes"}

func BindingsAsFiles(ctx pipeline.Context) {
	if !ctx.BindAsFiles() {
		return
	}
	secretName := ctx.BindingSecretName()
	bindingName := ctx.BindingName()
	applications, _ := ctx.Applications()
	for _, app := range applications {
		if app.SecretPath() != "" {
			continue
		}
		appResource := app.Resource()
		volumerResources, found, err := converter.NestedResources(&corev1.Volume{}, appResource.Object, volumesPath...)
		if err != nil {
			stop(ctx, err)
			return
		}

		volume, err := runtime.DefaultUnstructuredConverter.ToUnstructured(
			&corev1.Volume{
				Name: bindingName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			})
		if err != nil {
			stop(ctx, err)
			return
		}
		if found {
			var newVolumes []interface{}
			exist := false
			for _, v := range volumerResources {
				if v["name"] == volume["name"] {
					exist = true
					if !reflect.DeepEqual(v["secret"], volume["secret"]) {
						newVolumes = append(newVolumes, volume)
					} else {
						newVolumes = append(newVolumes, v)
					}
				} else {
					newVolumes = append(newVolumes, v)
				}
			}
			if !exist {
				newVolumes = append(newVolumes, volume)
			}
			if err = unstructured.SetNestedSlice(appResource.Object, newVolumes, volumesPath...); err != nil {
				stop(ctx, err)
				return
			}
		} else {
			if err = unstructured.SetNestedSlice(appResource.Object, []interface{}{volume}, volumesPath...); err != nil {
				stop(ctx, err)
				return
			}
		}

		containerResources, err := app.BindableContainers()
		if containerResources == nil && err == nil {
			err = errors.New("Containers not found in app resource")
		}
		if err != nil {
			stop(ctx, err)
			return
		}
		for _, container := range containerResources {
			mountPath, err := mountPath(container, ctx)
			if err != nil {
				stop(ctx, err)
				return
			}
			volumeMounts, found, err := converter.NestedResources(&corev1.VolumeMount{}, container, "volumeMounts")
			if err != nil {
				stop(ctx, err)
				return
			}
			u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.VolumeMount{
				Name:      bindingName,
				MountPath: mountPath,
			})
			if err != nil {
				stop(ctx, err)
				return
			}
			if found {
				var newVolumeMounts []interface{}
				exist := false
				for _, vm := range volumeMounts {
					if vm["name"] == u["name"] {
						exist = true
						if !reflect.DeepEqual(vm, u) {
							newVolumeMounts = append(newVolumeMounts, u)
						} else {
							newVolumeMounts = append(newVolumeMounts, vm)
						}
					} else {
						newVolumeMounts = append(newVolumeMounts, vm)
					}
				}
				if !exist {
					newVolumeMounts = append(newVolumeMounts, u)
				}
				container["volumeMounts"] = newVolumeMounts
			} else {
				if err := unstructured.SetNestedField(container, []interface{}{u}, "volumeMounts"); err != nil {
					stop(ctx, err)
					return
				}
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
	if secretName == "" {
		ctx.StopProcessing()
		return
	}
	for _, app := range applications {
		appResource := app.Resource()
		podSpec, found, err := unstructured.NestedFieldNoCopy(appResource.Object, volumesPath[:len(volumesPath)-1]...)
		if !found || err != nil {
			continue
		}
		podSpecMap, ok := podSpec.(map[string]interface{})
		if !ok {
			continue
		}
		volumeResources, found, _ := converter.NestedResources(&corev1.Volume{}, podSpecMap, "volumes")
		if found {
			for i, vol := range volumeResources {
				if val, found, err := unstructured.NestedString(vol, "name"); found && err == nil && val == bindingName {
					s := append(volumeResources[:i], volumeResources[i+1:]...)
					if len(s) == 0 {
						delete(podSpecMap, "volumes")
					} else {
						podSpecMap["volumes"] = s
					}
					break
				}
			}
		}
		containerResources, err := app.BindableContainers()
		if containerResources == nil && err == nil {
			ctx.StopProcessing()
			return
		}
		for _, container := range containerResources {
			envFrom, found, _ := converter.NestedResources(&corev1.EnvFromSource{}, container, "envFrom")
			if found {
				for i, envSource := range envFrom {
					if val, found, err := unstructured.NestedString(envSource, "secretRef", "name"); found && err == nil && val == secretName {
						s := append(envFrom[:i], envFrom[i+1:]...)
						if len(s) == 0 {
							delete(container, "envFrom")
						} else {
							container["envFrom"] = s
						}
						break
					}
				}
			}
			volumeMounts, found, _ := converter.NestedResources(&corev1.VolumeMount{}, container, "volumeMounts")
			if found {
				for i, vm := range volumeMounts {
					if val, found, err := unstructured.NestedString(vm, "name"); found && err == nil && val == bindingName {
						s := append(volumeMounts[:i], volumeMounts[i+1:]...)
						if len(s) == 0 {
							delete(container, "volumeMounts")
						} else {
							container["volumeMounts"] = s
						}
						break
					}
				}
			}
		}
	}
	ctx.StopProcessing()
}

const bindingRootEnvVar = "SERVICE_BINDING_ROOT"

func mountPath(container map[string]interface{}, ctx pipeline.Context) (string, error) {
	envs, found, err := converter.NestedResources(&corev1.EnvVar{}, container, "env")
	if err != nil {
		return "", err
	}
	bindingRoot := ""
	if found {
		for _, e := range envs {
			if e["name"] == bindingRootEnvVar {
				bindingRoot = fmt.Sprintf("%v", e["value"])
				return path.Join(bindingRoot, ctx.BindingName()), nil
			}
		}
	}

	bindingRoot = "/bindings"
	mp := path.Join(bindingRoot, ctx.BindingName())

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.EnvVar{
		Name:  bindingRootEnvVar,
		Value: bindingRoot,
	})
	if err != nil {
		return "", err
	}

	if found {
		container["env"] = append(envs, u)
	} else {
		if err := unstructured.SetNestedField(container, []interface{}{u}, "env"); err != nil {
			return "", err
		}
	}
	return mp, nil

}

func stop(ctx pipeline.Context, err error) {
	ctx.StopProcessing()
	ctx.Error(err)
	ctx.SetCondition(apis.Conditions().NotInjectionReady().Reason("Error").Msg(err.Error()).Build())
}
