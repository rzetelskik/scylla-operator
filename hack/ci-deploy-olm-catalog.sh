#!/bin/bash
#
# Copyright (C) 2025 ScyllaDB
#

set -euxEo pipefail
shopt -s inherit_errexit

source "$( dirname "${BASH_SOURCE[0]}" )/lib/bash.sh"
source "$( dirname "${BASH_SOURCE[0]}" )/lib/kube.sh"

if [[ -z ${1+x} ]]; then
  # TODO: improve usage
  echo "Missing OLM catalog image ref.\nUsage: ${0} <olm_catalog_image_ref>" >&2 >/dev/null
  exit 1
fi

trap cleanup-bg-jobs-on-exit EXIT

ARTIFACTS=${ARTIFACTS:-$( mktemp -d )}
OLM_CATALOG_IMAGE_REF=${1}

if [ -z "${ARTIFACTS_DEPLOY_DIR+x}" ]; then
  ARTIFACTS_DEPLOY_DIR=${ARTIFACTS}/deploy
fi

mkdir -p "${ARTIFACTS_DEPLOY_DIR}"
cat <<EOF > "${ARTIFACTS_DEPLOY_DIR}/scylladb-operator.catalog-source.yaml"
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: scylladb-operator-catalog
  namespace: openshift-marketplace
spec:
  displayName: ScyllaDB Operator
  image: "${OLM_CATALOG_IMAGE_REF}"
  publisher: ScyllaDB
  sourceType: grpc
  updateStrategy:
    registryPoll:
      interval: 10m
EOF

kubectl_create -n=openshift-marketplace -f="${ARTIFACTS_DEPLOY_DIR}/scylladb-operator.catalog-source.yaml"

cat <<EOF > "${ARTIFACTS_DEPLOY_DIR}/scylladb-operator.subscription.yaml"
apiVersion: v1
kind: Namespace
metadata:
  name: scylla-operator
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: scylladb-operator-operator-group
  namespace: scylla-operator
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: scylladb-operator-subscription
  namespace: scylla-operator
spec:
  # Fix channel in catalog. It doesn't match whatever is in the bundle.
  channel: candidate-v1.0
  installPlanApproval: Automatic
  name: scylladb-operator
  source: scylladb-operator-catalog
  sourceNamespace: openshift-marketplace
EOF

kubectl_create -n=scylla-operator -f="${ARTIFACTS_DEPLOY_DIR}/scylladb-operator.subscription.yaml"
