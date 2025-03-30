ScyllaDBManagerClusterRegistration (scylla.scylladb.com/v1alpha1)
=================================================================

| **APIVersion**: scylla.scylladb.com/v1alpha1
| **Kind**: ScyllaDBManagerClusterRegistration
| **PluralName**: scylladbmanagerclusterregistrations
| **SingularName**: scylladbmanagerclusterregistration
| **Scope**: Namespaced
| **ListKind**: ScyllaDBManagerClusterRegistrationList
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
   * - :ref:`metadata<api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.metadata>`
     - object
     - 
   * - :ref:`spec<api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.spec>`
     - object
     - spec defines the desired state of ScyllaDBManagerClusterRegistration.
   * - :ref:`status<api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.status>`
     - object
     - status reflects the observed state of ScyllaDBManagerClusterRegistration.

.. _api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.metadata:

.metadata
^^^^^^^^^

Description
"""""""""""


Type
""""
object


.. _api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.spec:

.spec
^^^^^

Description
"""""""""""
spec defines the desired state of ScyllaDBManagerClusterRegistration.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - :ref:`scyllaDBClusterRef<api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.spec.scyllaDBClusterRef>`
     - object
     - scyllaDBClusterRef specifies the typed reference to the local ScyllaDB cluster. Supported kinds are ScyllaDBCluster (scylla.scylladb.com/v1alpha1) and ScyllaDBDatacenter (scylla.scylladb.com/v1alpha1).
   * - scyllaDBManagerRef
     - string
     - scyllaDBManagerRef specifies the reference to the local ScyllaDBManager that the cluster should be registered with.

.. _api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.spec.scyllaDBClusterRef:

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

.. _api-scylla.scylladb.com-scylladbmanagerclusterregistrations-v1alpha1-.status:

.status
^^^^^^^

Description
"""""""""""
status reflects the observed state of ScyllaDBManagerClusterRegistration.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - clusterID
     - string
     - clusterID reflects the internal identification number of the cluster in ScyllaDB Manager state.
