/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventlistener

import (
	"context"
	"os"
	"time"

	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggersclient "github.com/tektoncd/triggers/pkg/client/injection/client"
	eventlistenerinformer "github.com/tektoncd/triggers/pkg/client/injection/informers/triggers/v1alpha1/eventlistener"
	"github.com/tektoncd/triggers/pkg/reconciler"
	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deployinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const (
	resyncPeriod = 10 * time.Hour
)

// NewController creates a new instance of an EventListener controller.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)
	kubeclientset := kubeclient.Get(ctx)
	triggersclientset := triggersclient.Get(ctx)
	eventListenerInformer := eventlistenerinformer.Get(ctx)
	deploymentInformer := deployinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)

	opt := reconciler.Options{
		KubeClientSet:     kubeclientset,
		TriggersClientSet: triggersclientset,
		ConfigMapWatcher:  cmw,
		Logger:            logger,
		ResyncPeriod:      resyncPeriod,
	}

	c := &Reconciler{
		Base:                reconciler.NewBase(opt, eventListenerAgentName),
		eventListenerLister: eventListenerInformer.Lister(),
		systemNamespace:     os.Getenv("SYSTEM_NAMESPACE"),
	}
	impl := controller.NewImpl(c, c.Logger, eventListenerControllerName)

	c.Logger.Info("Setting up event handlers")
	eventListenerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    impl.Enqueue,
		UpdateFunc: controller.PassNew(impl.Enqueue),
		DeleteFunc: impl.Enqueue,
	})

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("EventListener")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("EventListener")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
