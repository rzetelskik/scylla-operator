// Copyright (C) 2025 ScyllaDB

package scylladbmanagerclusterregistration

import (
	"context"
	"fmt"

	"github.com/scylladb/scylla-manager/v3/pkg/managerclient"
	"github.com/scylladb/scylla-manager/v3/swagger/gen/scylla-manager/models"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/helpers"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	"github.com/scylladb/scylla-operator/pkg/naming"
	hashutil "github.com/scylladb/scylla-operator/pkg/util/hash"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (smcrc *Controller) syncRegistration(
	ctx context.Context,
	smcr *scyllav1alpha1.ScyllaDBManagerClusterRegistration,
	status *scyllav1alpha1.ScyllaDBManagerClusterRegistrationStatus,
) ([]metav1.Condition, error) {
	var progressingConditions []metav1.Condition

	managerClient, err := smcrc.getManagerClient(ctx, smcr)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't get manager client: %w", err)
	}

	requiredManagerCluster, managedHash, err := smcrc.makeRequiredScyllaDBManagerCluster(ctx, smcr)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't make required ScyllaDB Manager cluster: %w", err)
	}

	managerCluster, found, err := getScyllaDBManagerCluster(ctx, smcr, managerClient)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't ScyllaDB Manager cluster: %w", err)
	}

	//if found {
	//	managerClusterOwnerUIDLabelValue, managerClusterHasOwnerUIDLabel := managerCluster.Labels[naming.OwnerUIDLabel]
	//	if managerClusterHasOwnerUIDLabel {
	//		if managerClusterOwnerUIDLabelValue != string(smcr.UID) {
	//			// Ideally we wouldn't do anything here as this is error-prone and might hinder discovering bugs.
	//			// However, the cluster could have been created by the legacy component (manager-controller), so we update it to become a new owner.
	//			klog.Warningf("Cluster %q already exists in ScyllaDB Manager state and has an owner UID label (%q), but it has a different owner. Adopting it.", managerCluster.Name, managerClusterOwnerUIDLabelValue)
	//
	//			requiredManagerCluster.ID = managerCluster.ID
	//			// Updating here.
	//
	//			return progressingConditions, nil
	//		}
	//
	//		if managedHash == managerCluster.Labels[naming.ManagedHash] {
	//			// Cluster matches the desired state, nothing to do.
	//			return progressingConditions, nil
	//		}
	//
	//		requiredManagerCluster.ID = managerCluster.ID
	//		// Updating here.
	//
	//		return progressingConditions, nil
	//	} else {
	//		klog.Warningf("ScyllaDB Manager cluster %q is missing an owner UID label. Deleting it to avoid a name collision.", managerCluster.Name)
	//
	//		err = managerClient.DeleteCluster(ctx, managerCluster.ID)
	//		if err != nil {
	//			return progressingConditions, fmt.Errorf("can't delete ScyllaDB Manager cluster %q: %w ", managerCluster.Name, err)
	//		}
	//	}
	//}

	if !found {
		var managerClusterID string
		managerClusterID, err = managerClient.CreateCluster(ctx, requiredManagerCluster)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't create ScyllaDB Manager cluster %q: %w", requiredManagerCluster.Name, err)
		}

		status.ClusterID = &managerClusterID
		return progressingConditions, nil
	}

	managerClusterOwnerUIDLabelValue, managerClusterHasOwnerUIDLabel := managerCluster.Labels[naming.OwnerUIDLabel]
	if !managerClusterHasOwnerUIDLabel {
		klog.Warningf("ScyllaDB Manager cluster %q is missing an owner UID label. Deleting it to avoid a name collision.", managerCluster.Name)

		err = managerClient.DeleteCluster(ctx, managerCluster.ID)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't delete ScyllaDB Manager cluster %q: %w ", managerCluster.Name, err)
		}

		//progressingConditions = append(progressingConditions, metav1.Condition{
		//	Type:               registrationControllerProgressingCondition,
		//	Status:             metav1.ConditionTrue,
		//	ObservedGeneration: smcr.Generation,
		//	Reason:             "DeletedCollidingScyllaDBManagerCluster",
		//	Message:            "Requeuing after deletion of a colliding ScyllaDB Manager cluster.",
		//})
		//
		//// TODO: progressing condition
	}

	if managerClusterOwnerUIDLabelValue != string(smcr.UID) {
		// Ideally we wouldn't do anything here as this is error-prone and might hinder discovering bugs.
		// However, the cluster could have been created by the legacy component (manager-controller), so we update it to become a new owner.
		klog.Warningf("Cluster %q already exists in ScyllaDB Manager state and has an owner UID label (%q), but it has a different owner. Adopting it.", managerCluster.Name, managerClusterOwnerUIDLabelValue)

		requiredManagerCluster.ID = managerCluster.ID
		// TODO: update.
		return progressingConditions, nil
	}

	if managedHash == managerCluster.Labels[naming.ManagedHash] {
		// Cluster matches the desired state, nothing to do.
		return progressingConditions, nil
	}

	requiredManagerCluster.ID = managerCluster.ID
	// TODO: update.
	return progressingConditions, nil
}

func (smcrc *Controller) makeRequiredScyllaDBManagerCluster(ctx context.Context, smcr *scyllav1alpha1.ScyllaDBManagerClusterRegistration) (*managerclient.Cluster, string, error) {
	managerClusterName := getScyllaDBManagerClusterName(smcr)

	authToken, err := smcrc.getAuthToken(ctx, smcr)
	if err != nil {
		return nil, "", fmt.Errorf("can't get auth token: %w", err)
	}

	host, err := smcrc.getHost(ctx, smcr)
	if err != nil {
		return nil, "", fmt.Errorf("can't get host: %w", err)
	}

	requiredManagerCluster := &managerclient.Cluster{
		Name:      managerClusterName,
		Host:      host,
		AuthToken: authToken,
		// TODO: enable CQL over TLS when https://github.com/scylladb/scylla-operator/issues/1673 is completed
		ForceNonSslSessionPort: true,
		ForceTLSDisabled:       true,
		Labels: map[string]string{
			// TODO: smcr UID or SDC/SC uid?
			naming.OwnerUIDLabel: string(smcr.UID),
		},
	}

	managedHash, err := hashutil.HashObjects(requiredManagerCluster)
	if err != nil {
		return nil, "", fmt.Errorf("can't calculate managed hash for cluster %q: %w", managerClusterName, err)
	}
	requiredManagerCluster.Labels[naming.ManagedHash] = managedHash

	return requiredManagerCluster, managedHash, nil
}

func getScyllaDBManagerClusterName(smcr *scyllav1alpha1.ScyllaDBManagerClusterRegistration) string {
	nameOverrideAnnotationValue, hasNameOverrideAnnotation := smcr.Annotations[naming.ScyllaDBManagerClusterRegistrationNameOverrideAnnotation]
	if hasNameOverrideAnnotation {
		return nameOverrideAnnotationValue
	}

	namespacePrefix := ""
	if smcr.Labels[naming.GlobalScyllaDBManagerLabel] == naming.LabelValueTrue {
		namespacePrefix = smcr.Namespace + "/"
	}

	return namespacePrefix + smcr.Spec.ScyllaDBClusterRef.Kind + "/" + smcr.Spec.ScyllaDBClusterRef.Name
}

func (smcrc *Controller) getHost(ctx context.Context, smcr *scyllav1alpha1.ScyllaDBManagerClusterRegistration) (string, error) {
	switch smcr.Spec.ScyllaDBClusterRef.Kind {
	case naming.ScyllaDBDatacenterKind:
		sdc, err := smcrc.scyllaDBDatacenterLister.ScyllaDBDatacenters(smcr.Namespace).Get(smcr.Spec.ScyllaDBClusterRef.Name)
		if err != nil {
			return "", fmt.Errorf("can't get ScyllaDBDatacenter %q: %w", naming.ManualRef(smcr.Namespace, smcr.Spec.ScyllaDBClusterRef.Name), err)
		}

		return naming.CrossNamespaceServiceName(sdc), nil

	default:
		return "", fmt.Errorf("unsupported scyllaDBClusterRef Kind: %q", smcr.Spec.ScyllaDBClusterRef.Kind)

	}
}

func (smcrc *Controller) getAuthToken(ctx context.Context, smcr *scyllav1alpha1.ScyllaDBManagerClusterRegistration) (string, error) {
	var authTokenSecretName string
	switch smcr.Spec.ScyllaDBClusterRef.Kind {
	case naming.ScyllaDBDatacenterKind:
		sdc, err := smcrc.scyllaDBDatacenterLister.ScyllaDBDatacenters(smcr.Namespace).Get(smcr.Spec.ScyllaDBClusterRef.Name)
		if err != nil {
			return "", fmt.Errorf("can't get ScyllaDBDatacenter %q: %w", naming.ManualRef(smcr.Namespace, smcr.Spec.ScyllaDBClusterRef.Name), err)
		}

		authTokenSecretName = naming.AgentAuthTokenSecretName(sdc)
		authTokenSecret, err := smcrc.secretLister.Secrets(smcr.Namespace).Get(authTokenSecretName)
		if err != nil {
			return "", fmt.Errorf("can't get secret %q: %w", naming.ManualRef(smcr.Namespace, authTokenSecretName), err)
		}

		authToken, err := helpers.GetAgentAuthTokenFromSecret(authTokenSecret)
		if err != nil {
			return "", fmt.Errorf("can't get agent auth token from secret %q: %w", naming.ObjRef(authTokenSecret), err)
		}

		return authToken, nil

	default:
		return "", fmt.Errorf("unsupported scyllaDBClusterRef Kind: %q", smcr.Spec.ScyllaDBClusterRef.Kind)

	}
}

func getScyllaDBManagerCluster(ctx context.Context, smcr *scyllav1alpha1.ScyllaDBManagerClusterRegistration, managerClient *managerclient.Client) (*managerclient.Cluster, bool, error) {
	// TODO: link to issue to add GetClusterByName method
	managerClusters, err := managerClient.ListClusters(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("can't list clusters registered with manager: %w", err)
	}

	// Cluster names in manager state are unique, so it suffices to only find one with a matching name.
	managerClusterName := getScyllaDBManagerClusterName(smcr)
	managerCluster, _, found := slices.Find(managerClusters, func(c *models.Cluster) bool {
		return c.Name == managerClusterName
	})

	return managerCluster, found, nil
}
