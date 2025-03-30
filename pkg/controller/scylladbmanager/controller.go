// Copyright (C) 2025 ScyllaDB

package scylladbmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	scyllav1alpha1client "github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned/typed/scylla/v1alpha1"
	scyllav1alpha1informers "github.com/scylladb/scylla-operator/pkg/client/scylla/informers/externalversions/scylla/v1alpha1"
	scyllav1alpha1listers "github.com/scylladb/scylla-operator/pkg/client/scylla/listers/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/kubeinterfaces"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/scheme"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	ControllerName = "ScyllaDBManagerController"
)

var (
	keyFunc                      = cache.DeletionHandlingMetaNamespaceKeyFunc
	scyllaDBManagerControllerGVK = scyllav1alpha1.GroupVersion.WithKind("ScyllaDBManager")
)

type Controller struct {
	kubeClient   kubernetes.Interface
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface

	scyllaDBManagerLister                    scyllav1alpha1listers.ScyllaDBManagerLister
	scyllaDBManagerClusterRegistrationLister scyllav1alpha1listers.ScyllaDBManagerClusterRegistrationLister
	scyllaDBDatacenterLister                 scyllav1alpha1listers.ScyllaDBDatacenterLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue    workqueue.RateLimitingInterface
	handlers *controllerhelpers.Handlers[*scyllav1alpha1.ScyllaDBManager]
}

func NewController(
	kubeClient kubernetes.Interface,
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface,
	scyllaDBManagerInformer scyllav1alpha1informers.ScyllaDBManagerInformer,
	scyllaDBManagerClusterRegistrationInformer scyllav1alpha1informers.ScyllaDBManagerClusterRegistrationInformer,
	scyllaDBDatacenterInformer scyllav1alpha1informers.ScyllaDBDatacenterInformer,
) (*Controller, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	smc := &Controller{
		kubeClient:   kubeClient,
		scyllaClient: scyllaClient,

		scyllaDBManagerLister:                    scyllaDBManagerInformer.Lister(),
		scyllaDBManagerClusterRegistrationLister: scyllaDBManagerClusterRegistrationInformer.Lister(),
		scyllaDBDatacenterLister:                 scyllaDBDatacenterInformer.Lister(),

		cachesToSync: []cache.InformerSynced{
			scyllaDBManagerInformer.Informer().HasSynced,
			scyllaDBManagerClusterRegistrationInformer.Informer().HasSynced,
			scyllaDBDatacenterInformer.Informer().HasSynced,
		},

		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "scylladbmanager-controller"}),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "scylladbmanager"),
	}

	var err error
	smc.handlers, err = controllerhelpers.NewHandlers[*scyllav1alpha1.ScyllaDBManager](
		smc.queue,
		keyFunc,
		scheme.Scheme,
		scyllaDBManagerControllerGVK,
		kubeinterfaces.NamespacedGetList[*scyllav1alpha1.ScyllaDBManager]{
			GetFunc: func(namespace, name string) (*scyllav1alpha1.ScyllaDBManager, error) {
				return smc.scyllaDBManagerLister.ScyllaDBManagers(namespace).Get(name)
			},
			ListFunc: func(namespace string, selector labels.Selector) (ret []*scyllav1alpha1.ScyllaDBManager, err error) {
				return smc.scyllaDBManagerLister.ScyllaDBManagers(namespace).List(selector)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create handlers: %w", err)
	}

	scyllaDBManagerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smc.addScyllaDBManager,
		UpdateFunc: smc.updateScyllaDBManager,
		DeleteFunc: smc.deleteScyllaDBManager,
	})

	scyllaDBManagerClusterRegistrationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smc.addScyllaDBManagerClusterRegistration,
		UpdateFunc: smc.updateScyllaDBManagerClusterRegistration,
		DeleteFunc: smc.deleteScyllaDBManagerClusterRegistration,
	})

	scyllaDBDatacenterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smc.addScyllaDBDatacenter,
		UpdateFunc: smc.updateScyllaDBDatacenter,
		DeleteFunc: smc.deleteScyllaDBDatacenter,
	})

	return smc, nil
}

func (smc *Controller) processNextItem(ctx context.Context) bool {
	key, quit := smc.queue.Get()
	if quit {
		return false
	}
	defer smc.queue.Done(key)

	err := smc.sync(ctx, key.(string))
	// TODO: Do smarter filtering then just Reduce to handle cases like 2 conflict errors.
	err = utilerrors.Reduce(err)
	switch {
	case err == nil:
		smc.queue.Forget(key)
		return true

	case apierrors.IsConflict(err):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Key", key, "Error", err)

	case apierrors.IsAlreadyExists(err):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Key", key, "Error", err)

	default:
		utilruntime.HandleError(fmt.Errorf("syncing key '%v' failed: %v", key, err))
	}

	smc.queue.AddRateLimited(key)

	return true
}

func (smc *Controller) runWorker(ctx context.Context) {
	for smc.processNextItem(ctx) {
	}
}

func (smc *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	klog.InfoS("Starting controller", "controller", ControllerName)

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Shutting down controller", "controller", ControllerName)
		smc.queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "controller", ControllerName)
	}()

	if !cache.WaitForNamedCacheSync(ControllerName, ctx.Done(), smc.cachesToSync...) {
		return
	}

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.UntilWithContext(ctx, smc.runWorker, time.Second)
		}()
	}

	<-ctx.Done()
}

func (smc *Controller) addScyllaDBManager(obj interface{}) {
	smc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.ScyllaDBManager),
		smc.handlers.Enqueue,
	)
}

func (smc *Controller) updateScyllaDBManager(old, cur interface{}) {
	smc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.ScyllaDBManager),
		cur.(*scyllav1alpha1.ScyllaDBManager),
		smc.handlers.Enqueue,
		smc.deleteScyllaDBManager,
	)
}

func (smc *Controller) deleteScyllaDBManager(obj interface{}) {
	smc.handlers.HandleDelete(
		obj,
		smc.handlers.Enqueue,
	)
}

func (smc *Controller) addScyllaDBManagerClusterRegistration(obj interface{}) {
	smc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration),
		smc.handlers.EnqueueOwner,
	)
}

func (smc *Controller) updateScyllaDBManagerClusterRegistration(old, cur interface{}) {
	smc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration),
		cur.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration),
		smc.handlers.EnqueueOwner,
		smc.deleteScyllaDBManagerClusterRegistration,
	)
}

func (smc *Controller) deleteScyllaDBManagerClusterRegistration(obj interface{}) {
	smc.handlers.HandleDelete(
		obj,
		smc.handlers.EnqueueOwner,
	)
}

func (smc *Controller) addScyllaDBDatacenter(obj interface{}) {
	smc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.ScyllaDBDatacenter),
		smc.enqueueGlobalScyllaDBManager,
	)
}

func (smc *Controller) updateScyllaDBDatacenter(old, cur interface{}) {
	smc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.ScyllaDBDatacenter),
		cur.(*scyllav1alpha1.ScyllaDBDatacenter),
		smc.enqueueGlobalScyllaDBManager,
		smc.deleteScyllaDBManagerClusterRegistration,
	)
}

func (smc *Controller) deleteScyllaDBDatacenter(obj interface{}) {
	smc.handlers.HandleDelete(
		obj,
		smc.enqueueGlobalScyllaDBManager,
	)
}

// TODO: fix this to check if manager is global
func (smc *Controller) enqueueGlobalScyllaDBManager(depth int, obj kubeinterfaces.ObjectInterface, op controllerhelpers.HandlerOperationType) {
	sdc := obj.(*scyllav1alpha1.ScyllaDBDatacenter)

	isRegisteringWithGlobalScyllaDBManager, ok := sdc.GetLabels()[naming.GlobalScyllaDBManagerRegistrationLabel]
	if !ok || isRegisteringWithGlobalScyllaDBManager != naming.LabelValueTrue {
		return
	}

	// TODO: manager namespace
	allScyllaDBManagers, err := smc.scyllaDBManagerLister.List(labels.Everything())
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("can't list ScyllaDBManagers: %w", err))
		return
	}

	// TODO: check if global manager
	klog.V(4).InfoSDepth(depth, "Enqueuing all global ScyllaDBManagers")
	for _, sm := range allScyllaDBManagers {
		smc.handlers.Enqueue(depth+1, sm, op)
	}
}
