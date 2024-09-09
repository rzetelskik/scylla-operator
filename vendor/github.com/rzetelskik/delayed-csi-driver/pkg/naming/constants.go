// Copyright (c) 2024 ScyllaDB.

package naming

const (
	DriverName = "delayed.csi.scylladb.com"
)

const (
	DelayedStorageBackendPersistentVolumeClaimRefAnnotation    = "delayed.csi.scylladb.com/backend-pvc-ref"
	DelayedStorageProxyPersistentVolumeClaimRefAnnotation      = "delayed.csi.scylladb.com/proxy-pvc-ref"
	DelayedStoragePersistentVolumeClaimBindCompletedAnnotation = "delayed.csi.scylladb.com/delayed-pvc-bind-completed"

	DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer = "delayed.csi.scylladb.com/backend-pvc-protection"

	DelayedStorageProxyVolumeAttachmentRefAnnotation = "delayed.csi.scylladb.com/proxy-va-ref"

	DelayedStorageMountedAnnotationFormat = "delayed.csi.scylladb.com/volume-mounted-%s"
	DelayedStorageMountedAnnotationTrue   = "true"
)
