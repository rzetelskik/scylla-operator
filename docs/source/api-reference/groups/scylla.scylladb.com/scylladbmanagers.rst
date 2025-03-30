ScyllaDBManager (scylla.scylladb.com/v1alpha1)
==============================================

| **APIVersion**: scylla.scylladb.com/v1alpha1
| **Kind**: ScyllaDBManager
| **PluralName**: scylladbmanagers
| **SingularName**: scylladbmanager
| **Scope**: Namespaced
| **ListKind**: ScyllaDBManagerList
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
   * - :ref:`metadata<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.metadata>`
     - object
     - 
   * - :ref:`spec<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec>`
     - object
     - spec defines the desired state of ScyllaDBManager.
   * - :ref:`status<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.status>`
     - object
     - status reflects the observed state of ScyllaDBManager.

.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.metadata:

.metadata
^^^^^^^^^

Description
"""""""""""


Type
""""
object


.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec:

.spec
^^^^^

Description
"""""""""""
spec defines the desired state of ScyllaDBManager.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - :ref:`selectors<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[]>`
     - array (object)
     - selectors specify which clusters should be registered with ScyllaDBManager. Supported kinds are ScyllaDBCluster (scylla.scylladb.com/v1alpha1) and ScyllaDBDatacenter (scylla.scylladb.com/v1alpha1).

.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[]:

.spec.selectors[]
^^^^^^^^^^^^^^^^^

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
   * - kind
     - string
     - kind specifies the type of resource.
   * - :ref:`labelSelector<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[].labelSelector>`
     - object
     - labelSelector specifies the label selector for resources of the specified kind.

.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[].labelSelector:

.spec.selectors[].labelSelector
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""
labelSelector specifies the label selector for resources of the specified kind.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - :ref:`matchExpressions<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[].labelSelector.matchExpressions[]>`
     - array (object)
     - matchExpressions is a list of label selector requirements. The requirements are ANDed.
   * - :ref:`matchLabels<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[].labelSelector.matchLabels>`
     - object
     - matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[].labelSelector.matchExpressions[]:

.spec.selectors[].labelSelector.matchExpressions[]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""
A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - key
     - string
     - key is the label key that the selector applies to.
   * - operator
     - string
     - operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
   * - values
     - array (string)
     - values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.

.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.spec.selectors[].labelSelector.matchLabels:

.spec.selectors[].labelSelector.matchLabels
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Description
"""""""""""
matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

Type
""""
object


.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.status:

.status
^^^^^^^

Description
"""""""""""
status reflects the observed state of ScyllaDBManager.

Type
""""
object


.. list-table::
   :widths: 25 10 150
   :header-rows: 1

   * - Property
     - Type
     - Description
   * - :ref:`conditions<api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.status.conditions[]>`
     - array (object)
     - conditions hold conditions describing ScyllaDBManager state.
   * - observedGeneration
     - integer
     - observedGeneration is the most recent generation observed for this ScyllaDBManager. It corresponds to the ScyllaDBManager's generation, which is updated on mutation by the API Server.

.. _api-scylla.scylladb.com-scylladbmanagers-v1alpha1-.status.conditions[]:

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
