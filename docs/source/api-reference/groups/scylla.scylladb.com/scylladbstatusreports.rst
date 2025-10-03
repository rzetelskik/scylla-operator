ScyllaDBStatusReport (scylla.scylladb.com/v1alpha1)
===================================================

| **APIVersion**: scylla.scylladb.com/v1alpha1
| **Kind**: ScyllaDBStatusReport
| **PluralName**: scylladbstatusreports
| **SingularName**: scylladbstatusreport
| **Scope**: Namespaced
| **ListKind**: ScyllaDBStatusReportList
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
   * - :ref:`datacenters<api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.datacenters[]>`
     - array (object)
     - Datacenters holds the list of datacenter reports.
   * - kind
     - string
     - Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
   * - :ref:`metadata<api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.metadata>`
     - object
     - 

.. _api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.datacenters[]:

.datacenters[]
^^^^^^^^^^^^^^

Description
"""""""""""
DatacenterStatusReport holds a report for a single datacenter.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - name
     - string
     - Name is the name of the datacenter.
   * - :ref:`nodes<api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.datacenters[].nodes[]>`
     - array (object)
     - Nodes holds the list of node reports from this datacenter.

.. _api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.datacenters[].nodes[]:

.datacenters[].nodes[]
^^^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""
NodeStatusReport holds a report for a single node.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - hostID
     - string
     - HostID is the ScyllaDB node's host ID.
   * - :ref:`observedNodes<api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.datacenters[].nodes[].observedNodes[]>`
     - array (object)
     - ObservedNodes holds the list of node statuses as observed by this node.

.. _api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.datacenters[].nodes[].observedNodes[]:

.datacenters[].nodes[].observedNodes[]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""


Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - hostID
     - string
     - HostID is the ScyllaDB node's host ID.
   * - status
     - string
     - Status is the status of the node.

.. _api-scylla.scylladb.com-scylladbstatusreports-v1alpha1-.metadata:

.metadata
^^^^^^^^^

Description
"""""""""""


Type
""""
object

