ScyllaDBManagerCluster (scylla.scylladb.com/v1alpha1)
=====================================================

| **APIVersion**: scylla.scylladb.com/v1alpha1
| **Kind**: ScyllaDBManagerCluster
| **PluralName**: scylladbmanagerclusters
| **SingularName**: scylladbmanagercluster
| **Scope**: Namespaced
| **ListKind**: ScyllaDBManagerClusterList
| **Served**: true
| **Storage**: true

Description
-----------


Specification
-------------

.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - apiVersion
     - string
     - APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
   * - kind
     - string
     - Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
   * - :ref:`metadata<api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.metadata>`
     - object
     - 
   * - :ref:`spec<api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.spec>`
     - object
     - spec defines the desired state of ScyllaDBManagerCluster.
   * - :ref:`status<api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.status>`
     - object
     - status reflects the observed state of ScyllaDBManagerCluster.

.. _api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.metadata:

.metadata
^^^^^^^^^

Description
"""""""""""


Type
""""
object


.. _api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.spec:

.spec
^^^^^

Description
"""""""""""
spec defines the desired state of ScyllaDBManagerCluster.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - :ref:`scyllaDBClusterRef<api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.spec.scyllaDBClusterRef>`
     - object
     - scyllaDBClusterRef specifies the typed reference to the local ScyllaDB cluster. Supported kinds are ScyllaDBCluster (scylla.scylladb.com/v1alpha1) and ScyllaDBDatacenter (scylla.scylladb.com/v1alpha1).
   * - scyllaDBManagerRef
     - string
     - scyllaDBManagerRef specifies the reference to the local ScyllaDBManager that the cluster should be registered with.

.. _api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.spec.scyllaDBClusterRef:

.spec.scyllaDBClusterRef
^^^^^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""
scyllaDBClusterRef specifies the typed reference to the local ScyllaDB cluster. Supported kinds are ScyllaDBCluster (scylla.scylladb.com/v1alpha1) and ScyllaDBDatacenter (scylla.scylladb.com/v1alpha1).

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - kind
     - string
     - kind specifies the type of the resource.
   * - name
     - string
     - name specifies the name of the resource.

.. _api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.status:

.status
^^^^^^^

Description
"""""""""""
status reflects the observed state of ScyllaDBManagerCluster.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - availableNodes
     - integer
     - availableNodes specify the total number of available nodes in datacenter.
   * - :ref:`conditions<api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.status.conditions[]>`
     - array (object)
     - conditions hold conditions describing ScyllaDBDatacenter state. To determine whether a cluster rollout is finished, look for Available=True,Progressing=False,Degraded=False.
   * - currentNodes
     - integer
     - currentNodes specify the total number of nodes created in datacenter.
   * - currentVersion
     - string
     - version specifies the current version of ScyllaDB in use.
   * - nodes
     - integer
     - nodes specify the total number of nodes requested in datacenter.
   * - observedGeneration
     - integer
     - observedGeneration is the most recent generation observed for this ScyllaDBDatacenter. It corresponds to the ScyllaDBDatacenter's generation, which is updated on mutation by the API Server.
   * - :ref:`racks<api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.status.racks[]>`
     - array (object)
     - racks reflect the status of datacenter racks.
   * - readyNodes
     - integer
     - readyNodes specify the total number of ready nodes in datacenter.
   * - updatedNodes
     - integer
     - updatedNodes specify the number of nodes matching the current spec in datacenter.
   * - updatedVersion
     - string
     - updatedVersion specifies the updated version of ScyllaDB.

.. _api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.status.conditions[]:

.status.conditions[]
^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""
Condition contains details for one aspect of the current state of this API Resource.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - lastTransitionTime
     - string
     - lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
   * - message
     - string
     - message is a human readable message indicating details about the transition. This may be an empty string.
   * - observedGeneration
     - integer
     - observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance.
   * - reason
     - string
     - reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty.
   * - status
     - string
     - status of the condition, one of True, False, Unknown.
   * - type
     - string
     - type of condition in CamelCase or in foo.example.com/CamelCase.

.. _api-scylla.scylladb.com-scylladbmanagerclusters-v1alpha1-.status.racks[]:

.status.racks[]
^^^^^^^^^^^^^^^

Description
"""""""""""
RackStatus is the status of a ScyllaDB Rack

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - availableNodes
     - integer
     - availableNodes specify the total number of available nodes in rack.
   * - currentNodes
     - integer
     - currentNodes specify the total number of nodes created in rack.
   * - currentVersion
     - string
     - version specifies the current version of ScyllaDB in use.
   * - name
     - string
     - name specifies the name of datacenter this status describes.
   * - nodes
     - integer
     - nodes specify the total number of nodes requested in rack.
   * - readyNodes
     - integer
     - readyNodes specify the total number of ready nodes in rack.
   * - stale
     - boolean
     - stale indicates if the current rack status is collected for a previous generation. stale should eventually become false when the appropriate controller writes a fresh status.
   * - updatedNodes
     - integer
     - updatedNodes specify the number of nodes matching the current spec in rack.
   * - updatedVersion
     - string
     - updatedVersion specifies the updated version of ScyllaDB.
