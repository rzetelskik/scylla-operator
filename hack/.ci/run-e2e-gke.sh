#!/usr/bin/env bash
#
# Copyright (C) 2023 ScyllaDB
#

set -euExo pipefail
shopt -s inherit_errexit

if [ -z ${SO_SUITE+x} ]; then
  echo "SO_SUITE can't be empty" > /dev/stderr
  exit 2
fi

if [ -z ${SO_IMAGE+x} ]; then
  echo "SO_IMAGE can't be empty" > /dev/stderr
  exit 2
fi

if [ -z ${SO_SCYLLACLUSTER_NODE_SERVICE_TYPE+x} ]; then
  echo "SO_SCYLLACLUSTER_NODE_SERVICE_TYPE can't be empty" > /dev/stderr
  exit 2
fi

if [ -z ${SO_SCYLLACLUSTER_NODES_BROADCAST_ADDRESS_TYPE+x} ]; then
  echo "SO_SCYLLACLUSTER_NODES_BROADCAST_ADDRESS_TYPE can't be empty" > /dev/stderr
  exit 2
fi

if [ -z ${SO_SCYLLACLUSTER_CLIENTS_BROADCAST_ADDRESS_TYPE+x} ]; then
  echo "SO_SCYLLACLUSTER_CLIENTS_BROADCAST_ADDRESS_TYPE can't be empty" > /dev/stderr
  exit 2
fi

if [ -z ${ARTIFACTS+x} ]; then
  echo "ARTIFACTS can't be empty" > /dev/stderr
  exit 2
fi

if [ -z ${KUBECONFIG_DIR+x} ]; then
  kubeconfigs=("${KUBECONFIG}")
else
  kubeconfigs=()
  for f in $( find "$(realpath "${KUBECONFIG_DIR}")" -maxdepth 1 -type f -name '*.kubeconfig' ); do
    kubeconfigs+=("${f}")
  done
fi

SO_NODECONFIG_PATH="${SO_NODECONFIG_PATH=./hack/.ci/manifests/cluster/nodeconfig.yaml}"
SO_DISABLE_NODECONFIG="${SO_DISABLE_NODECONFIG:-false}"

SO_BUCKET_NAME="${SO_BUCKET_NAME:-}"

field_manager=run-e2e-script

function kubectl_create {
    if [[ -z ${REENTRANT+x} ]]; then
        # In an actual CI run we have to enforce that no two objects have the same name.
        kubectl create --field-manager="${field_manager}" "$@"
    else
        # For development iterations we want to update the objects.
        kubectl apply --server-side=true --field-manager="${field_manager}" --force-conflicts "$@"
    fi
}

# $1 - namespace
# $2 - pod name
# $3 - container name
function wait-for-container-exit-with-logs {
  exit_code=""
  while [[ "${exit_code}" == "" ]]; do
    kubectl -n="${1}" logs -f pod/"${2}" -c="${3}" > /dev/stderr || echo "kubectl logs failed before pod has finished, retrying..." > /dev/stderr
    exit_code="$( kubectl -n="${1}" get pods/"${2}" --template='{{ range .status.containerStatuses }}{{ if and (eq .name "'"${3}"'") (ne .state.terminated.exitCode nil) }}{{ .state.terminated.exitCode }}{{ end }}{{ end }}' )"
  done
  echo -n "${exit_code}"
}

function gather-artifacts {
  (
    kubectl_create -n=e2e -f=- <<EOF
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: must-gather
  name: must-gather
spec:
  restartPolicy: Never
  containers:
  - name: wait-for-artifacts
    command:
    - /usr/bin/sleep
    - infinity
    image: "${SO_IMAGE}"
    imagePullPolicy: Always
    volumeMounts:
    - name: artifacts
      mountPath: /tmp/artifacts
  - name: must-gather
    args:
    - must-gather
    - --all-resources
    - --loglevel=2
    - --dest-dir=/tmp/artifacts
    image: "${SO_IMAGE}"
    imagePullPolicy: Always
    volumeMounts:
    - name: artifacts
      mountPath: /tmp/artifacts
  volumes:
  - name: artifacts
    emptyDir: {}
EOF
    kubectl -n=e2e wait --for=condition=Ready pod/must-gather

    exit_code="$( wait-for-container-exit-with-logs e2e must-gather must-gather )"

    kubectl -n=e2e cp --retries=42 must-gather:/tmp/artifacts "${ARTIFACTS}/must-gather"
    ls -l "${ARTIFACTS}/must-gather"

    kubectl -n=e2e delete pod/must-gather --wait=false

    if [[ "${exit_code}" != "0" ]]; then
      echo "Collecting artifacts using must-gather failed"
      exit "${exit_code}"
    fi
  )
}

function handle-exit {
  for i in "${!kubeconfigs[@]}"; do
    KUBECONFIG="${kubeconfigs[$i]}" gather-artifacts &
    gather_artifacts_bg_pids["${i}"]=$!
  done

  gather_artifacts_failed=0
  for pid in "${gather_artifacts_bg_pids[*]}"; do
    wait "${pid}" || gather_artifacts_failed=1
  done

  if [ ${gather_artifacts_failed} -eq 1 ]; then
    echo "Error gathering artifacts" > /dev/stderr
  fi
}

trap handle-exit EXIT

function setup {
  (
    # Allow admin to use ephemeralcontainers
    kubectl_create -f=- <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: scylladb-e2e:hotfixes
  labels:
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
- apiGroups:
  - ""
  resources:
  - pods/ephemeralcontainers
  verbs:
  - patch
EOF

    # FIXME: remove the workaround once https://github.com/scylladb/scylla-operator/issues/749 is done
    kubectl_create -n=default -f=- <<EOF
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: sysctl
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: sysctl
  template:
    metadata:
      labels:
        app.kubernetes.io/name: sysctl
    spec:
      containers:
      - name: sysctl
        securityContext:
          privileged: true
        image: "${SO_IMAGE}"
        imagePullPolicy: IfNotPresent
        command:
        - /usr/bin/bash
        - -euExo
        - pipefail
        - -O
        - inherit_errexit
        - -c
        args:
        - |
          sysctl fs.aio-max-nr=0xffffffff

          sleep infinity &
          wait
      nodeSelector:
        scylla.scylladb.com/node-type: scylla
EOF
    kubectl -n=default rollout status daemonset/sysctl

    kubectl apply --server-side -f=./pkg/api/scylla/v1alpha1/scylla.scylladb.com_nodeconfigs.yaml
    kubectl wait --for=condition=established crd/nodeconfigs.scylla.scylladb.com

    if [[ "${SO_DISABLE_NODECONFIG}" == "true"  ]] || [[ -z "${SO_NODECONFIG_PATH}" ]]; then
      echo "Skipping NodeConfig creation"
    else
      kubectl_create -f="${SO_NODECONFIG_PATH}"
      kubectl_create -n=local-csi-driver -f=./hack/.ci/manifests/namespaces/local-csi-driver/
    fi

    kubectl create namespace e2e --dry-run=client -o=yaml | kubectl_create -f=-
    kubectl create clusterrolebinding e2e --clusterrole=cluster-admin --serviceaccount=e2e:default --dry-run=client -o=yaml | kubectl_create -f=-

    REENTRANT=true timeout -v 10m ./hack/ci-deploy.sh "${SO_IMAGE}"
    # Raise loglevel in CI.
    # TODO: Replace it with ScyllaOperatorConfig field when available.
    kubectl -n=scylla-operator patch --field-manager="${field_manager}" deployment/scylla-operator --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--loglevel=4"}]'
    kubectl -n=scylla-operator rollout status deployment/scylla-operator

    kubectl -n=scylla-manager patch --field-manager="${field_manager}" deployment/scylla-manager-controller --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--loglevel=4"}]'
    kubectl -n=scylla-manager rollout status deployment/scylla-manager-controller
  )
}

SCYLLA_OPERATOR_FEATURE_GATES='AllAlpha=true,AllBeta=true'
export SCYLLA_OPERATOR_FEATURE_GATES

for i in "${!kubeconfigs[@]}"; do
  KUBECONFIG="${kubeconfigs[$i]}" setup &
  setup_bg_pids["${i}"]=$!
done

setup_failed=0
for pid in "${setup_bg_pids[*]}"; do
  wait "${pid}" || setup_failed=1
done

if [ ${setup_failed} -eq 1 ]; then
  echo "Error setting up clusters" > /dev/stderr
  exit 2
fi

KUBECONFIG="${kubeconfigs[0]}"
export KUBECONFIG

ingress_class_name='haproxy'
ingress_custom_annotations='haproxy.org/ssl-passthrough=true'
ingress_controller_address="$( kubectl -n=haproxy-ingress get svc haproxy-ingress --template='{{ .spec.clusterIP }}' ):9142"

kubectl create -n=e2e pdb my-pdb --selector='app=e2e' --min-available=1 --dry-run=client -o=yaml | kubectl_create -f=-

kubectl create -n=e2e configmap kubeconfigs "$( IFS=' '; echo "${kubeconfigs[@]/#/--from-file=}" )" --dry-run=client -o yaml | kubectl_create -f=-

gcs_sa_in_container_path=""
if [[ -n "${SO_GCS_SERVICE_ACCOUNT_CREDENTIALS_PATH+x}" ]]; then
  gcs_sa_in_container_path=/var/run/secrets/gcs-service-account-credentials/gcs-service-account.json
  kubectl create -n=e2e secret generic gcs-service-account-credentials --from-file="${SO_GCS_SERVICE_ACCOUNT_CREDENTIALS_PATH}" --dry-run=client -o=yaml | kubectl_create -f=-
else
  kubectl create -n=e2e secret generic gcs-service-account-credentials --dry-run=client -o=yaml | kubectl_create -f=-
fi

kubectl_create -n=e2e -f=- <<EOF
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: e2e
  name: e2e
spec:
  restartPolicy: Never
  containers:
  - name: wait-for-artifacts
    command:
    - /usr/bin/sleep
    - infinity
    image: "${SO_IMAGE}"
    imagePullPolicy: Always
    volumeMounts:
    - name: artifacts
      mountPath: /tmp/artifacts
  - name: e2e
    command:
    - scylla-operator-tests
    - run
    - "${SO_SUITE}"
    - "--kubeconfig=$( IFS=','; basenames=( ${kubeconfigs[@]##*/} ) && echo ${basenames[*]/#//var/run/configmaps/kubeconfigs/} )"
    - --loglevel=2
    - --color=false
    - --artifacts-dir=/tmp/artifacts
    - "--feature-gates=${SCYLLA_OPERATOR_FEATURE_GATES}"
    - "--ingress-controller-address=${ingress_controller_address}"
    - "--ingress-controller-ingress-class-name=${ingress_class_name}"
    - "--ingress-controller-custom-annotations=${ingress_custom_annotations}"
    - "--scyllacluster-node-service-type=${SO_SCYLLACLUSTER_NODE_SERVICE_TYPE}"
    - "--scyllacluster-nodes-broadcast-address-type=${SO_SCYLLACLUSTER_NODES_BROADCAST_ADDRESS_TYPE}"
    - "--scyllacluster-clients-broadcast-address-type=${SO_SCYLLACLUSTER_CLIENTS_BROADCAST_ADDRESS_TYPE}"
    - "--object-storage-bucket=${SO_BUCKET_NAME}"
    - "--gcs-service-account-key-path=${gcs_sa_in_container_path}"
    image: "${SO_IMAGE}"
    imagePullPolicy: Always
    volumeMounts:
    - name: artifacts
      mountPath: /tmp/artifacts
    - name: gcs-service-account-credentials
      mountPath: /var/run/secrets/gcs-service-account-credentials
    - name: kubeconfigs
      mountPath: /var/run/configmaps/kubeconfigs
  volumes:
  - name: artifacts
    emptyDir: {}
  - name: gcs-service-account-credentials
    secret:
      secretName: gcs-service-account-credentials
  - name: kubeconfigs
    configMap:
      name: kubeconfigs
EOF
kubectl -n=e2e wait --for=condition=Ready pod/e2e

exit_code="$( wait-for-container-exit-with-logs e2e e2e e2e )"

kubectl -n=e2e cp --retries=42 e2e:/tmp/artifacts -c=wait-for-artifacts "${ARTIFACTS}"
ls -l "${ARTIFACTS}"

kubectl -n=e2e delete pod/e2e --wait=false

if [[ "${exit_code}" != "0" ]]; then
  echo "E2E tests failed"
  exit "${exit_code}"
fi

echo "E2E tests finished successfully"
