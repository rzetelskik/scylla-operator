#!/usr/bin/env bash
#
# Copyright (C) 2025 ScyllaDB
#

set -euExo pipefail
shopt -s inherit_errexit

trap 'kill $( jobs -p ); exit 0' EXIT

#if [ -z "${ARTIFACTS+x}" ]; then
#  echo "ARTIFACTS can't be empty" > /dev/stderr
#  exit 2
#fi

#source "$( dirname "${BASH_SOURCE[0]}" )/../lib/kube.sh"
#source "$( dirname "${BASH_SOURCE[0]}" )/lib/e2e.sh"
#source "$( dirname "${BASH_SOURCE[0]}" )/run-e2e-shared.env.sh"
parent_dir="$( dirname "${BASH_SOURCE[0]}" )"

#trap gather-artifacts-on-exit EXIT
#trap gracefully-shutdown-e2es INT
#
#SO_NODECONFIG_PATH="${SO_NODECONFIG_PATH=${parent_dir}/manifests/cluster/nodeconfig-openshift-aws.yaml}"
#export SO_NODECONFIG_PATH
#
#SO_CSI_DRIVER_PATH="${SO_CSI_DRIVER_PATH=${parent_dir}/manifests/namespaces/local-csi-driver/}"
#export SO_CSI_DRIVER_PATH
#
# # TODO: When https://github.com/scylladb/scylla-operator/issues/2490 is completed,
# # we should make sure we have all required CRDs in the OpenShift cluster.
#SO_DISABLE_PROMETHEUS_OPERATOR="${SO_DISABLE_PROMETHEUS_OPERATOR:-true}"
#export SO_DISABLE_PROMETHEUS_OPERATOR
#
#SO_ENABLE_OPENSHIFT_USER_WORKLOAD_MONITORING="${SO_ENABLE_OPENSHIFT_USER_WORKLOAD_MONITORING:-true}"
#export SO_ENABLE_OPENSHIFT_USER_WORKLOAD_MONITORING
#
#run-deploy-script-in-all-clusters "${parent_dir}/../ci-deploy.sh"
#
#apply-e2e-workarounds-in-all-clusters
#run-e2e

REENTRANT=true
export REENTRANT
ARTIFACTS=${ARTIFACTS:-$( mktemp -d )}
export ARTIFACTS
KUBECONFIG=${KUBECONFIG:-/home/rzetelskik/.kube/config}
export KUBECONFIG

#"${parent_dir}/../ci-deploy-olm-catalog.sh" docker.io/rzetelskik/so-olm-catalog@sha256:4d2ba309988217be6f21c3c85554e55fd9e08c01015d3031e30c0af1fb22c277

SO_SUITE=scylla-operator/conformance/parallel/openshift
export SO_SUITE
SO_IMAGE=docker.io/scylladb/scylla-operator:latest
export SO_IMAGE
SO_SCYLLACLUSTER_NODE_SERVICE_TYPE=Headless
export SO_SCYLLACLUSTER_NODE_SERVICE_TYPE
SO_SCYLLACLUSTER_NODES_BROADCAST_ADDRESS_TYPE=PodIP
export SO_SCYLLACLUSTER_NODES_BROADCAST_ADDRESS_TYPE
SO_SCYLLACLUSTER_CLIENTS_BROADCAST_ADDRESS_TYPE=PodIP
export SO_SCYLLACLUSTER_CLIENTS_BROADCAST_ADDRESS_TYPE
SO_E2E_PARALLELISM=8
export SO_E2E_PARALLELISM
SO_E2E_TIMEOUT=60m
export SO_E2E_TIMEOUT
timeout --verbose --signal INT --kill-after=100m 90m "${parent_dir}/run-e2e-openshift-aws.sh"