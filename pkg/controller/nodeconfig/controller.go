// Copyright (C) 2021 ScyllaDB

package nodeconfig

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	apimachineryutilerrors "k8s.io/apimachinery/pkg/util/errors"
	apimachineryutilruntime "k8s.io/apimachinery/pkg/util/runtime"
	apimachineryutilwait "k8s.io/apimachinery/pkg/util/wait"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	rbacv1informers "k8s.io/client-go/informers/rbac/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	ControllerName = "NodeConfigController"
	// maxSyncDuration enforces preemption. Do not raise the value! Controllers shouldn't actively wait,
	// but rather use the queue.
	maxSyncDuration = 30 * time.Second
)

var (
	keyFunc                 = cache.DeletionHandlingMetaNamespaceKeyFunc
	nodeConfigControllerGVK = scyllav1alpha1.GroupVersion.WithKind("NodeConfig")
)

type Controller struct {
	kubeClient   kubernetes.Interface
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface

	nodeConfigLister           scyllav1alpha1listers.NodeConfigLister
	scyllaOperatorConfigLister scyllav1alpha1listers.ScyllaOperatorConfigLister
	clusterRoleLister          rbacv1listers.ClusterRoleLister
	clusterRoleBindingLister   rbacv1listers.ClusterRoleBindingLister
	roleLister                 rbacv1listers.RoleLister
	roleBindingLister          rbacv1listers.RoleBindingLister
	daemonSetLister            appsv1listers.DaemonSetLister
	namespaceLister            corev1listers.NamespaceLister
	nodeLister                 corev1listers.NodeLister
	serviceAccountLister       corev1listers.ServiceAccountLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue    workqueue.TypedRateLimitingInterface[string]
	handlers *controllerhelpers.Handlers[*scyllav1alpha1.NodeConfig]

	operatorImage string
}

func isManagedByNodeConfigController(obj kubeinterfaces.ObjectInterface) bool {
	return obj.GetLabels()[naming.NodeConfigNameLabel] == naming.NodeConfigAppName
}

func NewController(
	kubeClient kubernetes.Interface,
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface,
	nodeConfigInformer scyllav1alpha1informers.NodeConfigInformer,
	scyllaOperatorConfigInformer scyllav1alpha1informers.ScyllaOperatorConfigInformer,
	clusterRoleInformer rbacv1informers.ClusterRoleInformer,
	clusterRoleBindingInformer rbacv1informers.ClusterRoleBindingInformer,
	roleInformer rbacv1informers.RoleInformer,
	roleBindingInformer rbacv1informers.RoleBindingInformer,
	daemonSetInformer appsv1informers.DaemonSetInformer,
	namespaceInformer corev1informers.NamespaceInformer,
	nodeInformer corev1informers.NodeInformer,
	serviceAccountInformer corev1informers.ServiceAccountInformer,
	operatorImage string,
) (*Controller, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	ncc := &Controller{
		kubeClient:   kubeClient,
		scyllaClient: scyllaClient,

		nodeConfigLister:           nodeConfigInformer.Lister(),
		scyllaOperatorConfigLister: scyllaOperatorConfigInformer.Lister(),
		clusterRoleLister:          clusterRoleInformer.Lister(),
		clusterRoleBindingLister:   clusterRoleBindingInformer.Lister(),
		roleLister:                 roleInformer.Lister(),
		roleBindingLister:          roleBindingInformer.Lister(),
		daemonSetLister:            daemonSetInformer.Lister(),
		namespaceLister:            namespaceInformer.Lister(),
		nodeLister:                 nodeInformer.Lister(),
		serviceAccountLister:       serviceAccountInformer.Lister(),

		cachesToSync: []cache.InformerSynced{
			nodeConfigInformer.Informer().HasSynced,
			scyllaOperatorConfigInformer.Informer().HasSynced,
			clusterRoleInformer.Informer().HasSynced,
			clusterRoleBindingInformer.Informer().HasSynced,
			roleInformer.Informer().HasSynced,
			roleBindingInformer.Informer().HasSynced,
			daemonSetInformer.Informer().HasSynced,
			namespaceInformer.Informer().HasSynced,
			nodeInformer.Informer().HasSynced,
			serviceAccountInformer.Informer().HasSynced,
		},

		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "NodeConfig-controller"}),

		queue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{
				Name: "NodeConfig",
			},
		),

		operatorImage: operatorImage,
	}

	var err error
	ncc.handlers, err = controllerhelpers.NewHandlers[*scyllav1alpha1.NodeConfig](
		ncc.queue,
		keyFunc,
		scheme.Scheme,
		nodeConfigControllerGVK,
		kubeinterfaces.GlobalGetList[*scyllav1alpha1.NodeConfig]{
			GetFunc: func(name string) (*scyllav1alpha1.NodeConfig, error) {
				return ncc.nodeConfigLister.Get(name)
			},
			ListFunc: func(selector labels.Selector) (ret []*scyllav1alpha1.NodeConfig, err error) {
				return ncc.nodeConfigLister.List(selector)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create handlers: %w", err)
	}

	nodeConfigInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addNodeConfig,
		UpdateFunc: ncc.updateNodeConfig,
		DeleteFunc: ncc.deleteNodeConfig,
	})

	// TODO: react to label changes on nodes
	// nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	AddFunc:    ncc.addNode,
	// 	UpdateFunc: ncc.updateNode,
	// 	DeleteFunc: ncc.deleteNode,
	// })

	scyllaOperatorConfigInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addScyllaOperatorConfig,
		UpdateFunc: ncc.updateScyllaOperatorConfig,
		DeleteFunc: ncc.deleteScyllaOperatorConfig,
	})

	clusterRoleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addClusterRole,
		UpdateFunc: ncc.updateClusterRole,
		DeleteFunc: ncc.deleteClusterRole,
	})

	clusterRoleBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addClusterRoleBinding,
		UpdateFunc: ncc.updateClusterRoleBinding,
		DeleteFunc: ncc.deleteClusterRoleBinding,
	})

	roleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addRole,
		UpdateFunc: ncc.updateRole,
		DeleteFunc: ncc.deleteRole,
	})

	roleBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addRoleBinding,
		UpdateFunc: ncc.updateRoleBinding,
		DeleteFunc: ncc.deleteRoleBinding,
	})

	serviceAccountInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addServiceAccount,
		UpdateFunc: ncc.updateServiceAccount,
		DeleteFunc: ncc.deleteServiceAccount,
	})

	daemonSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addDaemonSet,
		UpdateFunc: ncc.updateDaemonSet,
		DeleteFunc: ncc.deleteDaemonSet,
	})

	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ncc.addNamespace,
		UpdateFunc: ncc.updateNamespace,
		DeleteFunc: ncc.deleteNamespace,
	})

	return ncc, nil
}

func (ncc *Controller) addDaemonSet(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*appsv1.DaemonSet),
		ncc.handlers.EnqueueOwner,
	)
}

func (ncc *Controller) updateDaemonSet(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*appsv1.DaemonSet),
		cur.(*appsv1.DaemonSet),
		ncc.handlers.EnqueueOwner,
		ncc.deleteDaemonSet,
	)
}

func (ncc *Controller) deleteDaemonSet(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueOwner,
	)
}

func (ncc *Controller) addNamespace(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*corev1.Namespace),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) updateNamespace(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*corev1.Namespace),
		cur.(*corev1.Namespace),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
		ncc.deleteNamespace,
	)
}

func (ncc *Controller) deleteNamespace(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) addServiceAccount(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*corev1.ServiceAccount),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) updateServiceAccount(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*corev1.ServiceAccount),
		cur.(*corev1.ServiceAccount),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
		ncc.deleteServiceAccount,
	)
}

func (ncc *Controller) deleteServiceAccount(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) addClusterRoleBinding(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*rbacv1.ClusterRoleBinding),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) updateClusterRoleBinding(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*rbacv1.ClusterRoleBinding),
		cur.(*rbacv1.ClusterRoleBinding),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
		ncc.deleteClusterRoleBinding,
	)
}

func (ncc *Controller) deleteClusterRoleBinding(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) addClusterRole(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*rbacv1.ClusterRole),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) updateClusterRole(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*rbacv1.ClusterRole),
		cur.(*rbacv1.ClusterRole),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
		ncc.deleteClusterRole,
	)
}

func (ncc *Controller) deleteClusterRole(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) addRoleBinding(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*rbacv1.RoleBinding),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) updateRoleBinding(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*rbacv1.RoleBinding),
		cur.(*rbacv1.RoleBinding),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
		ncc.deleteRoleBinding,
	)
}

func (ncc *Controller) deleteRoleBinding(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) addRole(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*rbacv1.Role),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) updateRole(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*rbacv1.Role),
		cur.(*rbacv1.Role),
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
		ncc.deleteRole,
	)
}

func (ncc *Controller) deleteRole(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAllWithUntypedFilterFunc(isManagedByNodeConfigController),
	)
}

func (ncc *Controller) addNodeConfig(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.NodeConfig),
		ncc.handlers.Enqueue,
	)
}

func (ncc *Controller) updateNodeConfig(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.NodeConfig),
		cur.(*scyllav1alpha1.NodeConfig),
		ncc.handlers.Enqueue,
		ncc.deleteNodeConfig,
	)
}

func (ncc *Controller) deleteNodeConfig(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.Enqueue,
	)
}

func (ncc *Controller) addScyllaOperatorConfig(obj interface{}) {
	ncc.handlers.HandleAdd(
		obj.(*scyllav1alpha1.ScyllaOperatorConfig),
		ncc.handlers.EnqueueAll,
	)
}

func (ncc *Controller) updateScyllaOperatorConfig(old, cur interface{}) {
	ncc.handlers.HandleUpdate(
		old.(*scyllav1alpha1.ScyllaOperatorConfig),
		cur.(*scyllav1alpha1.ScyllaOperatorConfig),
		ncc.handlers.EnqueueAll,
		ncc.deleteScyllaOperatorConfig,
	)
}

func (ncc *Controller) deleteScyllaOperatorConfig(obj interface{}) {
	ncc.handlers.HandleDelete(
		obj,
		ncc.handlers.EnqueueAll,
	)
}

func (ncc *Controller) processNextItem(ctx context.Context) bool {
	key, quit := ncc.queue.Get()
	if quit {
		return false
	}
	defer ncc.queue.Done(key)

	ctx, cancel := context.WithTimeout(ctx, maxSyncDuration)
	defer cancel()
	err := ncc.sync(ctx, key)
	// TODO: Do smarter filtering then just Reduce to handle cases like 2 conflict errors.
	err = apimachineryutilerrors.Reduce(err)
	switch {
	case err == nil:
		ncc.queue.Forget(key)
		return true

	case apierrors.IsConflict(err):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Key", key, "Error", err)

	case apierrors.IsAlreadyExists(err):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Key", key, "Error", err)

	default:
		apimachineryutilruntime.HandleError(fmt.Errorf("syncing key '%v' failed: %v", key, err))
	}

	ncc.queue.AddRateLimited(key)

	return true
}

func (ncc *Controller) runWorker(ctx context.Context) {
	for ncc.processNextItem(ctx) {
	}
}

func (ncc *Controller) Run(ctx context.Context, workers int) {
	defer apimachineryutilruntime.HandleCrash()

	klog.InfoS("Starting controller", "controller", ControllerName)

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Shutting down controller", "controller", ControllerName)
		ncc.queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "controller", ControllerName)
	}()

	if !cache.WaitForNamedCacheSync(ControllerName, ctx.Done(), ncc.cachesToSync...) {
		return
	}

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			apimachineryutilwait.UntilWithContext(ctx, ncc.runWorker, time.Second)
		}()
	}

	<-ctx.Done()
}
