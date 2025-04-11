// Copyright (C) 2025 ScyllaDB

package globalscylladbmanager

import (
	"fmt"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	scyllav1alpha1client "github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned/typed/scylla/v1alpha1"
	scyllav1alpha1informers "github.com/scylladb/scylla-operator/pkg/client/scylla/informers/externalversions/scylla/v1alpha1"
	scyllav1alpha1listers "github.com/scylladb/scylla-operator/pkg/client/scylla/listers/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type Controller struct {
	*controllertools.Observer

	kubeClient   kubernetes.Interface
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface

	globalScyllaDBManagerDeploymentLister    appsv1listers.DeploymentLister
	scyllaDBManagerClusterRegistrationLister scyllav1alpha1listers.ScyllaDBManagerClusterRegistrationLister
	scyllaDBDatacenterLister                 scyllav1alpha1listers.ScyllaDBDatacenterLister
}

func NewController(
	kubeClient kubernetes.Interface,
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface,
	globalScyllaDBManagerDeploymentInformer appsv1informers.DeploymentInformer,
	scyllaDBManagerClusterRegistrationInformer scyllav1alpha1informers.ScyllaDBManagerClusterRegistrationInformer,
	scyllaDBDatacenterInformer scyllav1alpha1informers.ScyllaDBDatacenterInformer,
) (*Controller, error) {
	gsmc := &Controller{
		kubeClient:   kubeClient,
		scyllaClient: scyllaClient,

		globalScyllaDBManagerDeploymentLister:    globalScyllaDBManagerDeploymentInformer.Lister(),
		scyllaDBManagerClusterRegistrationLister: scyllaDBManagerClusterRegistrationInformer.Lister(),
		scyllaDBDatacenterLister:                 scyllaDBDatacenterInformer.Lister(),
	}

	observer := controllertools.NewObserver(
		"globalscylladbmanager-controller",
		kubeClient.CoreV1().Events(corev1.NamespaceAll),
		gsmc.sync,
	)

	//globalScyllaDBManagerDeploymentHandler, err := globalScyllaDBManagerDeploymentInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	//if err != nil {
	//	return nil, fmt.Errorf("can't add global ScyllaDB Manager Deployment handler: %w", err)
	//}
	//observer.AddCachesToSync(globalScyllaDBManagerDeploymentHandler.HasSynced)

	scyllaDBManagerClusterRegistrationHandler, err := scyllaDBManagerClusterRegistrationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    gsmc.addScyllaDBManagerClusterRegistration,
		UpdateFunc: gsmc.updateScyllaDBManagerClusterRegistration,
		DeleteFunc: gsmc.deleteScyllaDBManagerClusterRegistration,
	})
	if err != nil {
		return nil, fmt.Errorf("can't add ScyllaDBManagerClusterRegistration handler: %w", err)
	}
	observer.AddCachesToSync(scyllaDBManagerClusterRegistrationHandler.HasSynced)

	scyllaDBDatacenterHandler, err := scyllaDBDatacenterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    gsmc.addScyllaDBDatacenter,
		UpdateFunc: gsmc.updateScyllaDBDatacenter,
		DeleteFunc: gsmc.deleteScyllaDBDatacenter,
	})
	if err != nil {
		return nil, fmt.Errorf("can't add ScyllaDBDatacenter handler: %w", err)
	}
	observer.AddCachesToSync(scyllaDBDatacenterHandler.HasSynced)

	gsmc.Observer = observer

	// TODO: remove this?
	// Start immediately, global ScyllaDB Manager might already be deployed.
	gsmc.Enqueue()

	return gsmc, nil
}

func (gsmc *Controller) addScyllaDBManagerClusterRegistration(obj interface{}) {
	smcr := obj.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration)

	//owned, err := gsmc.isOwnedByGlobalScyllaDBManager(smcr)
	//if err != nil {
	//	utilruntime.HandleError(err)
	//	return
	//}
	//
	//if !owned {
	//	klog.V(5).InfoS("Not enqueueing ScyllaDBManagerClusterRegistration not owned by global ScyllaDB Manager", "ScyllaDBManagerClusterRegistration", klog.KObj(smcr), "RV", smcr.ResourceVersion)
	//	return
	//}

	klog.V(4).InfoS(
		"Observed addition of ScyllaDBManagerClusterRegistration",
		"ScyllaDBManagerClusterRegistration", klog.KObj(smcr),
		"RV", smcr.ResourceVersion,
	)
	gsmc.Enqueue()
}

func (gsmc *Controller) updateScyllaDBManagerClusterRegistration(old, cur interface{}) {
	oldSMCR := old.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration)
	currentSMCR := cur.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration)

	if currentSMCR.UID != oldSMCR.UID {
		key, err := keyFunc(oldSMCR)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("can't get key for object %#v: %w", oldSMCR, err))
			return
		}

		gsmc.deleteScyllaDBManagerClusterRegistration(cache.DeletedFinalStateUnknown{
			Key: key,
			Obj: oldSMCR,
		})
	}

	//owned, err := gsmc.isOwnedByGlobalScyllaDBManager(currentSMCR)
	//if err != nil {
	//	utilruntime.HandleError(err)
	//	return
	//}
	//
	//if !owned {
	//	klog.V(5).InfoS("Not enqueueing ScyllaDBManagerClusterRegistration not owned by global ScyllaDB Manager", "ScyllaDBManagerClusterRegistration", klog.KObj(currentSMCR), "RV", currentSMCR.ResourceVersion)
	//	return
	//}

	klog.V(4).InfoS(
		"Observed update of ScyllaDBManagerClusterRegistration",
		"ScyllaDBManagerClusterRegistration", klog.KObj(currentSMCR),
		"RV", fmt.Sprintf("%s->%s", oldSMCR.ResourceVersion, currentSMCR.ResourceVersion),
		"UID", fmt.Sprintf("%s->%s", oldSMCR.UID, currentSMCR.UID),
	)
	gsmc.Enqueue()
}

func (gsmc *Controller) deleteScyllaDBManagerClusterRegistration(obj interface{}) {
	smcr, ok := obj.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration)
	if !ok {
		var tombstone cache.DeletedFinalStateUnknown
		tombstone, ok = obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("can't get object from tombstone %#v", obj))
			return
		}
		smcr, ok = tombstone.Obj.(*scyllav1alpha1.ScyllaDBManagerClusterRegistration)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contains an object that is not a ScyllaDBManagerClusterRegistration %#v", obj))
			return
		}
	}

	//owned, err := gsmc.isOwnedByGlobalScyllaDBManager(smcr)
	//if err != nil {
	//	utilruntime.HandleError(err)
	//	return
	//}
	//
	//if !owned {
	//	klog.V(5).InfoS("Not enqueueing ScyllaDBManagerClusterRegistration not owned by global ScyllaDB Manager", "ScyllaDBManagerClusterRegistration", klog.KObj(smcr), "RV", smcr.ResourceVersion)
	//	return
	//}

	klog.V(4).InfoS(
		"Observed deletion of ScyllaDBManagerClusterRegistration",
		"ScyllaDBManagerClusterRegistration", klog.KObj(smcr),
		"RV", smcr.ResourceVersion,
	)
	gsmc.Enqueue()
}

func (gsmc *Controller) addScyllaDBDatacenter(obj interface{}) {
	sdc := obj.(*scyllav1alpha1.ScyllaDBDatacenter)

	//if !isScyllaDBDatacenterRegisteringWithGlobalScyllaDBManager(sdc) {
	//	return
	//}

	klog.V(4).InfoS(
		"Observed addition of ScyllaDBDatacenter",
		"ScyllaDBDatacenter", klog.KObj(sdc),
		"RV", sdc.ResourceVersion,
	)
	gsmc.Enqueue()
}

func (gsmc *Controller) updateScyllaDBDatacenter(old, cur interface{}) {
	oldSDC := old.(*scyllav1alpha1.ScyllaDBDatacenter)
	currentSDC := cur.(*scyllav1alpha1.ScyllaDBDatacenter)

	if currentSDC.UID != oldSDC.UID {
		key, err := keyFunc(oldSDC)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("can't get key for object %#v: %w", oldSDC, err))
			return
		}

		gsmc.deleteScyllaDBDatacenter(cache.DeletedFinalStateUnknown{
			Key: key,
			Obj: oldSDC,
		})
	}

	//if !isScyllaDBDatacenterRegisteringWithGlobalScyllaDBManager(currentSDC) {
	//	return
	//}

	klog.V(4).InfoS(
		"Observed update of ScyllaDBDatacenter",
		"ScyllaDBDatacenter", klog.KObj(currentSDC),
		"RV", fmt.Sprintf("%s->%s", oldSDC.ResourceVersion, currentSDC.ResourceVersion),
		"UID", fmt.Sprintf("%s->%s", oldSDC.UID, currentSDC.UID),
	)
	gsmc.Enqueue()
}

func (gsmc *Controller) deleteScyllaDBDatacenter(obj interface{}) {
	sdc, ok := obj.(*scyllav1alpha1.ScyllaDBDatacenter)
	if !ok {
		var tombstone cache.DeletedFinalStateUnknown
		tombstone, ok = obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("can't get object from tombstone %#v", obj))
			return
		}
		sdc, ok = tombstone.Obj.(*scyllav1alpha1.ScyllaDBDatacenter)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contains an object that is not a ScyllaDBDatacenter %#v", obj))
			return
		}
	}

	//if !isScyllaDBDatacenterRegisteringWithGlobalScyllaDBManager(sdc) {
	//	return
	//}

	klog.V(4).InfoS(
		"Observed deletion of ScyllaDBDatacenter",
		"ScyllaDBDatacenter", klog.KObj(sdc),
		"RV", sdc.ResourceVersion,
	)
	gsmc.Enqueue()
}

//func (gsmc *Controller) isOwnedByGlobalScyllaDBManager(obj metav1.Object) (bool, error) {
//	globalScyllaDBManagerRef, err := gsmc.newOwningDeploymentControllerRef()
//	if err != nil {
//		return false, fmt.Errorf("can't get owning deployment controller ref: %w", err)
//	}
//
//	objControllerRef := metav1.GetControllerOfNoCopy(obj)
//	return apiequality.Semantic.DeepEqual(objControllerRef, globalScyllaDBManagerRef), nil
//}
