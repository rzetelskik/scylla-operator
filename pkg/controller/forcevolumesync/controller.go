// Copyright (C) 2024 ScyllaDB

package forcevolumesync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	hashutil "github.com/scylladb/scylla-operator/pkg/util/hash"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type Controller struct {
	*controllertools.Observer

	namespace     string
	podName       string
	volumesToSync sets.Set[string]

	kubeClient kubernetes.Interface

	podLister       corev1listers.PodLister
	configMapLister corev1listers.ConfigMapLister
	secretLister    corev1listers.SecretLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue workqueue.RateLimitingInterface
	key   string
}

func NewController(
	namespace string,
	podName string,
	volumesToSync []string,
	kubeClient kubernetes.Interface,
	podInformer corev1informers.PodInformer,
	configMapInformer corev1informers.ConfigMapInformer,
	secretInformer corev1informers.SecretInformer,
) (*Controller, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	// Sanity check.
	if len(namespace) == 0 {
		return nil, fmt.Errorf("pod namespace can't be empty")
	}
	if len(podName) == 0 {
		return nil, fmt.Errorf("pod name can't be empty")
	}

	controller := &Controller{
		namespace:       namespace,
		podName:         podName,
		volumesToSync:   sets.New(volumesToSync...),
		kubeClient:      kubeClient,
		podLister:       podInformer.Lister(),
		configMapLister: configMapInformer.Lister(),
		secretLister:    secretInformer.Lister(),
	}

	observer := controllertools.NewObserver(
		"force-volume-sync",
		kubeClient.CoreV1().Events(corev1.NamespaceAll),
		controller.sync,
	)

	podHandler, err := podInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add pod event handler: %w", err)
	}
	observer.AddCachesToSync(podHandler.HasSynced)

	configMapHandler, err := configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addConfigMap,
		UpdateFunc: controller.updateConfigMap,
		DeleteFunc: controller.deleteConfigMap,
	})
	if err != nil {
		return nil, fmt.Errorf("can't add config map event handler: %w", err)
	}
	observer.AddCachesToSync(configMapHandler.HasSynced)

	secretHandler, err := secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addSecret,
		UpdateFunc: controller.updateSecret,
		DeleteFunc: controller.deleteSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("can't add secret event handler: %w", err)
	}
	observer.AddCachesToSync(secretHandler.HasSynced)

	controller.Observer = observer

	return controller, nil
}

func (c *Controller) sync(ctx context.Context) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing", "Name", c.Observer.Name(), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing", "Name", c.Observer.Name(), "duration", time.Since(startTime))
	}()

	pod, err := c.podLister.Pods(c.namespace).Get(c.podName)
	if err != nil {
		return fmt.Errorf("can't get pod %q: %w", naming.ManualRef(c.namespace, c.podName), err)
	}

	var objectsToHash []interface{}
	for _, volume := range pod.Spec.Volumes {
		if !c.volumesToSync.Has(volume.Name) {
			continue
		}

		if volume.ConfigMap != nil {
			cm, err := c.configMapLister.ConfigMaps(c.namespace).Get(volume.ConfigMap.Name)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("can't get configmap %q: %w", naming.ManualRef(c.namespace, volume.ConfigMap.Name), err)
				}

				klog.ErrorS(err, "configmap has not been created yet", "Pod", klog.KObj(pod), "Volume", volume.Name, "ConfigMap", klog.KRef(c.namespace, volume.ConfigMap.Name))
				continue
			}

			objectsToHash = append(objectsToHash, cm.Data)
		} else if volume.Secret != nil {
			secret, err := c.secretLister.Secrets(c.namespace).Get(volume.Secret.SecretName)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("can't get secret %q: %w", naming.ManualRef(c.namespace, volume.Secret.SecretName), err)
				}

				klog.ErrorS(err, "secret has not been created yet", "Pod", klog.KObj(pod), "Volume", volume.Name, "Secret", klog.KRef(c.namespace, volume.Secret.SecretName))
				continue
			}

			objectsToHash = append(objectsToHash, secret.Data)
		} else {
			klog.ErrorS(nil, "volume source is not supported", "Pod", klog.KObj(pod), "Volume", volume.Name, "SupportedVolumeSources", strings.Join([]string{"secret", "configMap"}, ","))
		}
	}

	hash, err := hashutil.HashObjects(objectsToHash)
	if err != nil {
		return fmt.Errorf("can't hash objects: %w", err)
	}

	if controllerhelpers.HasMatchingAnnotation(pod, naming.ForceVolumeSyncManagedHashAnnotation, hash) {
		return nil
	}

	patch, err := controllerhelpers.PrepareSetAnnotationPatch(pod, naming.ForceVolumeSyncManagedHashAnnotation, pointer.Ptr(hash))
	if err != nil {
		return fmt.Errorf("can't prepate set annotation patch: %w", err)
	}

	_, err = c.kubeClient.CoreV1().Pods(pod.Namespace).Patch(ctx, pod.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch pod %q: %w", naming.ObjRef(pod), err)
	}

	return nil
}

func (c *Controller) addConfigMap(obj interface{}) {
	cm := obj.(*corev1.ConfigMap)
	c.enqueueIfConfigMapIsUsedAsVolume(cm)
}

func (c *Controller) updateConfigMap(old, cur interface{}) {
	oldCM := old.(*corev1.ConfigMap)
	curCM := cur.(*corev1.ConfigMap)

	if curCM.UID != oldCM.UID {
		key, err := keyFunc(oldCM)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", oldCM, err))
			return
		}

		c.deleteConfigMap(cache.DeletedFinalStateUnknown{
			Key: key,
			Obj: oldCM,
		})
	}

	c.enqueueIfConfigMapIsUsedAsVolume(curCM)
}

func (c *Controller) deleteConfigMap(obj interface{}) {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		var tombstone cache.DeletedFinalStateUnknown
		tombstone, ok = obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		cm, ok = tombstone.Obj.(*corev1.ConfigMap)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a ConfigMap %#v", obj))
			return
		}
	}

	c.enqueueIfConfigMapIsUsedAsVolume(cm)
}

func (c *Controller) enqueueIfConfigMapIsUsedAsVolume(cm *corev1.ConfigMap) {
	c.enqueueWithIsUsedAsVolumeFunc(func(volume corev1.Volume) bool {
		if volume.ConfigMap == nil {
			return false
		}

		if volume.ConfigMap.Name != cm.Name {
			return false
		}

		return true
	})
}

func (c *Controller) addSecret(obj interface{}) {
	secret := obj.(*corev1.Secret)
	c.enqueueIfSecretIsUsedAsVolume(secret)
}

func (c *Controller) updateSecret(old, cur interface{}) {
	oldSecret := old.(*corev1.Secret)
	curSecret := cur.(*corev1.Secret)

	if curSecret.UID != oldSecret.UID {
		key, err := keyFunc(oldSecret)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", oldSecret, err))
			return
		}

		c.deleteConfigMap(cache.DeletedFinalStateUnknown{
			Key: key,
			Obj: oldSecret,
		})
	}

	c.enqueueIfSecretIsUsedAsVolume(curSecret)
}

func (c *Controller) deleteSecret(obj interface{}) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		var tombstone cache.DeletedFinalStateUnknown
		tombstone, ok = obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		secret, ok = tombstone.Obj.(*corev1.Secret)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Secret %#v", obj))
			return
		}
	}

	c.enqueueIfSecretIsUsedAsVolume(secret)
}

func (c *Controller) enqueueIfSecretIsUsedAsVolume(secret *corev1.Secret) {
	c.enqueueWithIsUsedAsVolumeFunc(func(volume corev1.Volume) bool {
		if volume.Secret == nil {
			return false
		}

		if volume.Secret.SecretName != secret.Name {
			return false
		}

		return true
	})
}

func (c *Controller) enqueueWithIsUsedAsVolumeFunc(isUsedAsVolume func(volume corev1.Volume) bool) {
	pod, err := c.podLister.Pods(c.namespace).Get(c.podName)
	if err != nil {
		// We will enqueue on a Pod event eventually.
		utilruntime.HandleError(fmt.Errorf("can't get pod %q: %w", naming.ManualRef(c.namespace, c.podName), err))
		return
	}

	for _, v := range pod.Spec.Volumes {
		if !c.volumesToSync.Has(v.Name) {
			continue
		}

		if isUsedAsVolume(v) {
			c.Observer.Enqueue()
			return
		}
	}
}
