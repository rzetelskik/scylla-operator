// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

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
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	"github.com/scylladb/scylla-operator/pkg/kubeinterfaces"
	"github.com/scylladb/scylla-operator/pkg/scheme"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	batchv1informers "k8s.io/client-go/informers/batch/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	ControllerName = "ScyllaDBManagerTaskController"

	// maxSyncDuration enforces preemption. Do not raise the value! Controllers shouldn't actively wait,
	// but rather use the queue.
	// Unfortunately, Scylla Manager calls are synchronous, internally retried and can take ages.
	// Contrary to what it should be, this needs to be quite high.
	maxSyncDuration = 2 * time.Minute
)

var (
	keyFunc                          = cache.DeletionHandlingMetaNamespaceKeyFunc
	scyllaDBManagerTaskControllerGVK = scyllav1alpha1.GroupVersion.WithKind("ScyllaDBManagerTask")
)

type Controller struct {
	kubeClient   kubernetes.Interface
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface

	scyllaDBManagerTaskLister scyllav1alpha1listers.ScyllaDBManagerTaskLister
	//scyllaDBManagerClusterRegistrationLister scyllav1alpha1listers.ScyllaDBManagerClusterRegistrationLister
	scyllaDBDatacenterLister scyllav1alpha1listers.ScyllaDBDatacenterLister
	secretLister             corev1listers.SecretLister
	jobLister                batchv1listers.JobLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue    workqueue.RateLimitingInterface
	handlers *controllerhelpers.Handlers[*scyllav1alpha1.ScyllaDBManagerTask]
}

func NewController(
	kubeClient kubernetes.Interface,
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface,
	scyllaDBManagerTaskInformer scyllav1alpha1informers.ScyllaDBManagerTaskInformer,
	scyllaDBDatacenterInformer scyllav1alpha1informers.ScyllaDBDatacenterInformer,
	secretInformer corev1informers.SecretInformer,
	jobInformer batchv1informers.JobInformer,
) (*Controller, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	smtc := &Controller{
		kubeClient:   kubeClient,
		scyllaClient: scyllaClient,

		scyllaDBManagerTaskLister: scyllaDBManagerTaskInformer.Lister(),
		scyllaDBDatacenterLister:  scyllaDBDatacenterInformer.Lister(),
		secretLister:              secretInformer.Lister(),
		jobLister:                 jobInformer.Lister(),

		cachesToSync: []cache.InformerSynced{
			scyllaDBManagerTaskInformer.Informer().HasSynced,
			scyllaDBDatacenterInformer.Informer().HasSynced,
			secretInformer.Informer().HasSynced,
			jobInformer.Informer().HasSynced,
		},

		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "scylladbmanagertask-controller"}),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "scylladbmanagertask"),
	}

	var err error
	smtc.handlers, err = controllerhelpers.NewHandlers[*scyllav1alpha1.ScyllaDBManagerTask](
		smtc.queue,
		keyFunc,
		scheme.Scheme,
		scyllaDBManagerTaskControllerGVK,
		kubeinterfaces.NamespacedGetList[*scyllav1alpha1.ScyllaDBManagerTask]{
			GetFunc: func(namespace, name string) (*scyllav1alpha1.ScyllaDBManagerTask, error) {
				return smtc.scyllaDBManagerTaskLister.ScyllaDBManagerTasks(namespace).Get(name)
			},
			ListFunc: func(namespace string, selector labels.Selector) (ret []*scyllav1alpha1.ScyllaDBManagerTask, err error) {
				return smtc.scyllaDBManagerTaskLister.ScyllaDBManagerTasks(namespace).List(selector)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create handlers: %w", err)
	}

	scyllaDBManagerTaskInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smtc.addScyllaDBManagerTask,
		UpdateFunc: smtc.updateScyllaDBManagerTask,
		DeleteFunc: smtc.deleteScyllaDBManagerTask,
	})

	scyllaDBDatacenterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smtc.addScyllaDBDatacenter,
		UpdateFunc: smtc.updateScyllaDBDatacenter,
		DeleteFunc: smtc.deleteScyllaDBDatacenter,
	})

	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smtc.addSecret,
		UpdateFunc: smtc.updateSecret,
		DeleteFunc: smtc.deleteSecret,
	})

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    smtc.addJob,
		UpdateFunc: smtc.updateJob,
		DeleteFunc: smtc.deleteJob,
	})

	return smtc, nil
}

func (smtc *Controller) processNextItem(ctx context.Context) bool {
	key, quit := smtc.queue.Get()
	if quit {
		return false
	}
	defer smtc.queue.Done(key)

	ctx, cancel := context.WithTimeout(ctx, maxSyncDuration)
	defer cancel()
	err := smtc.sync(ctx, key.(string))
	// TODO: Do smarter filtering then just Reduce to handle cases like 2 conflict errors.
	err = utilerrors.Reduce(err)
	switch {
	case err == nil:
		smtc.queue.Forget(key)
		return true

	case apierrors.IsConflict(err):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Key", key, "Error", err)

	case apierrors.IsAlreadyExists(err):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Key", key, "Error", err)

	default:
		if controllertools.IsNonRetriable(err) {
			klog.InfoS("Hit non-retriable error. Dropping the item from the queue.", "Error", err)
			smtc.queue.Forget(key)
			return true
		}

		utilruntime.HandleError(fmt.Errorf("syncing key '%v' failed: %v", key, err))

	}

	smtc.queue.AddRateLimited(key)

	return true
}

func (smtc *Controller) runWorker(ctx context.Context) {
	for smtc.processNextItem(ctx) {
	}
}

func (smtc *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	klog.InfoS("Starting controller", "controller", ControllerName)

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Shutting down controller", "controller", ControllerName)
		smtc.queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "controller", ControllerName)
	}()

	if !cache.WaitForNamedCacheSync(ControllerName, ctx.Done(), smtc.cachesToSync...) {
		return
	}

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.UntilWithContext(ctx, smtc.runWorker, time.Second)
		}()
	}

	<-ctx.Done()
}

func (smtc *Controller) addScyllaDBManagerTask(obj interface{}) {
	smtc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.ScyllaDBManagerTask),
		smtc.handlers.Enqueue,
	)
}

func (smtc *Controller) updateScyllaDBManagerTask(old, cur interface{}) {
	smtc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.ScyllaDBManagerTask),
		cur.(*scyllav1alpha1.ScyllaDBManagerTask),
		smtc.handlers.Enqueue,
		smtc.deleteScyllaDBManagerTask,
	)
}

func (smtc *Controller) deleteScyllaDBManagerTask(obj interface{}) {
	smtc.handlers.HandleDelete(
		obj,
		smtc.handlers.Enqueue,
	)
}

func (smtc *Controller) addScyllaDBDatacenter(obj interface{}) {
	smtc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.ScyllaDBDatacenter),
		smtc.handlers.EnqueueAllFunc(smtc.enqueueThroughScyllaDBDatacenter(obj.(*scyllav1alpha1.ScyllaDBDatacenter))),
	)
}

func (smtc *Controller) updateScyllaDBDatacenter(old, cur interface{}) {
	smtc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.ScyllaDBDatacenter),
		cur.(*scyllav1alpha1.ScyllaDBDatacenter),
		smtc.handlers.EnqueueAllFunc(smtc.enqueueThroughScyllaDBDatacenter(cur.(*scyllav1alpha1.ScyllaDBDatacenter))), smtc.deleteScyllaDBDatacenter,
	)
}

func (smtc *Controller) deleteScyllaDBDatacenter(obj interface{}) {
	smtc.handlers.HandleDelete(
		obj,
		smtc.handlers.EnqueueAllFunc(smtc.enqueueThroughScyllaDBDatacenter(obj.(*scyllav1alpha1.ScyllaDBDatacenter))))
}

func (smtc *Controller) addSecret(obj interface{}) {
	smtc.handlers.HandleAdd(
		obj.(*corev1.Secret),
		smtc.enqueueThroughOwner,
	)
}

func (smtc *Controller) updateSecret(old, cur interface{}) {
	smtc.handlers.HandleUpdate(
		old.(*corev1.Secret),
		cur.(*corev1.Secret),
		smtc.enqueueThroughOwner,
		smtc.deleteSecret,
	)
}

func (smtc *Controller) deleteSecret(obj interface{}) {
	smtc.handlers.HandleDelete(
		obj,
		smtc.enqueueThroughOwner,
	)
}

func (smtc *Controller) addJob(obj interface{}) {
	smtc.handlers.HandleAdd(
		obj.(*batchv1.Job),
		smtc.handlers.EnqueueOwner,
	)
}

func (smtc *Controller) updateJob(old, cur interface{}) {
	smtc.handlers.HandleUpdate(
		old.(*batchv1.Job),
		cur.(*batchv1.Job),
		smtc.handlers.EnqueueOwner,
		smtc.deleteJob,
	)
}

func (smtc *Controller) deleteJob(obj interface{}) {
	smtc.handlers.HandleDelete(
		obj,
		smtc.handlers.EnqueueOwner,
	)
}

func (smtc *Controller) enqueueThroughOwner(depth int, obj kubeinterfaces.ObjectInterface, op controllerhelpers.HandlerOperationType) {
	controllerRef := metav1.GetControllerOf(obj)
	if controllerRef == nil {
		return
	}

	switch controllerRef.Kind {
	case scyllav1alpha1.ScyllaDBDatacenterGVK.Kind:
		sdc, err := smtc.scyllaDBDatacenterLister.ScyllaDBDatacenters(obj.GetNamespace()).Get(controllerRef.Name)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}

		smtc.handlers.EnqueueAllFunc(smtc.enqueueThroughScyllaDBDatacenter(sdc))(depth+1, obj, op)
		return

	default:
		// Nothing to do.
		return

	}
}

func (smtc *Controller) enqueueThroughScyllaDBDatacenter(sdc *scyllav1alpha1.ScyllaDBDatacenter) controllerhelpers.EnqueueFuncType {
	return smtc.handlers.EnqueueAllFunc(smtc.handlers.EnqueueWithFilterFunc(func(smt *scyllav1alpha1.ScyllaDBManagerTask) bool {
		switch smt.Spec.ScyllaDBClusterRef.Kind {
		case scyllav1alpha1.ScyllaDBDatacenterGVK.Kind:
			return smt.Spec.ScyllaDBClusterRef.Name == sdc.Name

		default:
			return false

		}
	}))
}
