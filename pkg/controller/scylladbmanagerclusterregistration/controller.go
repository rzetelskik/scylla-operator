// Copyright (C) 2025 ScyllaDB

package scylladbmanagerclusterregistration

import (
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	scyllav1alpha1client "github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned/typed/scylla/v1alpha1"
	scyllav1alpha1informers "github.com/scylladb/scylla-operator/pkg/client/scylla/informers/externalversions/scylla/v1alpha1"
	scyllav1alpha1listers "github.com/scylladb/scylla-operator/pkg/client/scylla/listers/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	ControllerName = "ScyllaDBManagerClusterRegistrationController"
)

var (
	keyFunc                                         = cache.DeletionHandlingMetaNamespaceKeyFunc
	scyllaDBManagerClusterRegistrationControllerGVK = scyllav1alpha1.GroupVersion.WithKind("ScyllaDBManagerClusterRegistration")
)

type Controller struct {
	kubeClient   kubernetes.Interface
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface

	scyllaDBManagerClusterRegistrationLister scyllav1alpha1listers.ScyllaDBManagerClusterRegistrationLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue    workqueue.RateLimitingInterface
	handlers *controllerhelpers.Handlers[*scyllav1alpha1.ScyllaDBManagerClusterRegistration]
}

func NewController(
	kubeClient kubernetes.Interface,
	scyllaClient scyllav1alpha1client.ScyllaV1alpha1Interface,
	scyllaDBManagerClusterRegistrationInformer scyllav1alpha1informers.ScyllaDBManagerClusterRegistrationInformer,
) (*Controller, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	smcrc := &Controller{
		kubeClient:   kubeClient,
		scyllaClient: scyllaClient,

		scyllaDBManagerClusterRegistrationLister: scyllaDBManagerClusterRegistrationInformer.Lister(),
	}

	return smcrc, nil
}
