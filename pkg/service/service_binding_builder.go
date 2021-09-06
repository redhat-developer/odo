package service

// This file is a fork of github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/collect/impl.go
// This file was created to deal with the permission issue that comes up when user does not access to clusterwide resources,
// and they want to use odo link without the Service Binding Operator.
// The only difference between the original and forked implementation is that we check if the user is forbidden from accessing
// CRD, and if they are, then we simply ignore checking CRD while linking.
// For more information, see issue: https://github.com/openshift/odo/issues/4965
// In case there is a need to revert the changes, or we figure out an alternate way of allowing forbidden users to link without SBO,
// we can go back to using the builder.DefaultBuilder instead of the OdoDefaultBuilder in getPipeline.

import (
	"fmt"
	"strings"

	"k8s.io/klog"

	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/binding"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/collect"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/mapping"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/naming"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/project"
	"github.com/redhat-developer/service-binding-operator/pkg/util"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultFlow = []pipeline.Handler{
	pipeline.HandlerFunc(project.Unbind),
	pipeline.HandlerFunc(collect.PreFlight),
	pipeline.HandlerFunc(ProvisionedService),
	pipeline.HandlerFunc(collect.DirectSecretReference),
	pipeline.HandlerFunc(BindingDefinitions),
	pipeline.HandlerFunc(collect.BindingItems),
	pipeline.HandlerFunc(collect.OwnedResources),
	pipeline.HandlerFunc(mapping.Handle),
	pipeline.HandlerFunc(naming.Handle),
	pipeline.HandlerFunc(project.PreFlightCheck()),
	pipeline.HandlerFunc(project.InjectSecretRef),
	pipeline.HandlerFunc(project.BindingsAsEnv),
	pipeline.HandlerFunc(project.BindingsAsFiles),
	pipeline.HandlerFunc(project.PostFlightCheck),
}

var OdoDefaultBuilder = builder.Builder().WithHandlers(defaultFlow...)

func ProvisionedService(ctx pipeline.Context) {
	services, _ := ctx.Services()

	for _, service := range services {
		res := service.Resource()
		secretName, found, err := unstructured.NestedString(res.Object, "status", "binding", "name")
		if err != nil {
			requestRetry(ctx, collect.ErrorReadingBindingReason, err)
			return
		}
		if found {
			if secretName != "" {
				secret, err := ctx.ReadSecret(res.GetNamespace(), secretName)
				if err != nil {
					requestRetry(ctx, collect.ErrorReadingSecret, err)
					return
				}
				ctx.AddBindings(&pipeline.SecretBackedBindings{Service: service, Secret: secret})
			}
		} else {
			crd, err := service.CustomResourceDefinition()
			if err != nil {
				// If the user does not have permission to access CRD, ignore
				if !kerrors.IsForbidden(err) {
					requestRetry(ctx, collect.ErrorReadingCRD, err)
					return
				} else {
					klog.V(4).Infof("Skipping the check for CRD, user does not have access")
				}
			}
			if crd == nil {
				continue
			}
			v, ok := crd.Resource().GetAnnotations()[collect.ProvisionedServiceAnnotationKey]
			if ok && v == "true" {
				requestRetry(ctx, collect.ErrorReadingBindingReason, fmt.Errorf("CRD of service %v/%v indicates provisioned service, but no secret name provided under .status.binding.name", res.GetNamespace(), res.GetName()))
				return
			}
		}
	}
}

func BindingDefinitions(ctx pipeline.Context) {
	services, _ := ctx.Services()

	for _, service := range services {
		anns := make(map[string]string)
		crd, err := service.CustomResourceDefinition()
		if err != nil {
			// If the user does not have permission to access CRD, ignore
			if !kerrors.IsForbidden(err) {
				requestRetry(ctx, collect.ErrorReadingCRD, err)
				return
			} else {
				klog.V(4).Infof("Skipping the check for CRD, user does not have access")
			}
		}
		if crd != nil {
			descr, err := crd.Descriptor()
			if err != nil {
				requestRetry(ctx, collect.ErrorReadingDescriptorReason, err)
				return
			}
			if descr != nil {
				util.MergeMaps(anns, bindingAnnotations(descr))
			}
			util.MergeMaps(anns, crd.Resource().GetAnnotations())
		}

		util.MergeMaps(anns, service.Resource().GetAnnotations())

		for k, v := range anns {
			definition, err := makeBindingDefinition(k, v, ctx)
			if err != nil {
				continue
			}
			service.AddBindingDef(definition)
		}
	}
}

func requestRetry(ctx pipeline.Context, reason string, err error) {
	ctx.RetryProcessing(err)
	ctx.SetCondition(notCollectionReadyCond(reason, err))
}

func notCollectionReadyCond(reason string, err error) *metav1.Condition {
	return apis.Conditions().NotCollectionReady().Reason(reason).Msg(err.Error()).Build()
}

func makeBindingDefinition(key string, value string, ctx pipeline.Context) (binding.Definition, error) {
	return binding.NewDefinitionBuilder(key,
		value,
		func(namespace string, name string) (*unstructured.Unstructured, error) {
			return ctx.ReadConfigMap(namespace, name)
		},
		func(namespace string, name string) (*unstructured.Unstructured, error) {
			return ctx.ReadSecret(namespace, name)
		}).Build()
}

func bindingAnnotations(crdDescription *olmv1alpha1.CRDDescription) map[string]string {
	anns := make(map[string]string)
	for _, sd := range crdDescription.StatusDescriptors {
		objectType := getObjectType(sd.XDescriptors)
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "status", objectType)
		}
	}

	for _, sd := range crdDescription.SpecDescriptors {
		objectType := getObjectType(sd.XDescriptors)
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "spec", objectType)
		}
	}
	return anns
}

func getObjectType(descriptors []string) string {
	typeAnno := "urn:alm:descriptor:io.kubernetes:"
	for _, desc := range descriptors {
		if strings.HasPrefix(desc, typeAnno) {
			return strings.TrimPrefix(desc, typeAnno)
		}
	}
	return ""
}

func loadDescriptor(anns map[string]string, path string, descriptor string, root string, objectType string) {
	if !strings.HasPrefix(descriptor, binding.AnnotationPrefix) {
		return
	}

	keys := strings.Split(descriptor, ":")
	key := binding.AnnotationPrefix
	value := ""

	if len(keys) > 1 {
		key += "/" + keys[1]
	} else {
		key += "/" + path
	}

	p := []string{fmt.Sprintf("path={.%s.%s}", root, path)}
	if len(keys) > 1 {
		p = append(p, keys[2:]...)
	}
	if objectType != "" {
		p = append(p, []string{fmt.Sprintf("objectType=%s", objectType)}...)
	}

	value += strings.Join(p, ",")
	anns[key] = value
}
