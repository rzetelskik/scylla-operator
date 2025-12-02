#!/bin/bash
#
# Copyright (C) 2025 ScyllaDB
#

set -euxEo pipefail
shopt -s inherit_errexit

function strip_digestfile() {
  cat "${1}" | tr '\n' ' '
}

tmp_dir="$( mktemp -d )"
function cleanup() {
    rm -rf "${tmp_dir}"
}
trap cleanup EXIT

bundle_image_ref="docker.io/rzetelskik/so-olm-bundle:latest"
catalog_image_ref="docker.io/rzetelskik/so-olm-catalog:latest"
platforms="linux/amd64"

mkdir -p "${tmp_dir}/bundle"
cp -r ./../bundle/manifests ./../bundle/metadata "${tmp_dir}/bundle/"
chmod -R a+rw "${tmp_dir}"

podman run -it --rm \
  --user root:root \
  --pull=IfNotPresent \
  --volume="${tmp_dir}":/workspace:rw,Z \
  --workdir=/workspace \
  quay.io/redhat-isv/operator-pipelines-images:released \
  bundle-dockerfile \
  --bundle-path=/workspace/bundle \
  --destination=/workspace/Dockerfile \
  --verbose

buildah manifest rm "${bundle_image_ref}" 2>/dev/null || true
buildah build --squash --format=docker \
--file "${tmp_dir}/Dockerfile" \
--platform "${platforms}" \
--manifest "${bundle_image_ref}" \
"${tmp_dir}/bundle"

digestfile="$( mktemp )"
buildah manifest push --all --digestfile="${digestfile}" "${bundle_image_ref}"

# TODO: verify digest is not empty

bundle_image_digest="docker.io/rzetelskik/so-olm-bundle@$( strip_digestfile "${digestfile}" )"

mkdir -p "${tmp_dir}/opm-storage/catalog"

cat << EOF >> "${tmp_dir}/opm-storage/catalog-template.yaml"
Schema: olm.semver
Candidate:
  Bundles:
  - Image: ${bundle_image_digest}
EOF

podman run -it --rm \
  --user root:root \
  --pull=IfNotPresent \
  --volume="${tmp_dir}/opm-storage":/workspace:rw,Z \
  --workdir=/workspace \
  quay.io/scylladb/scylla-operator-images:kube-tools \
  /bin/bash -euExo pipefail -O inherit_errexit -c '
opm generate dockerfile catalog
opm alpha render-template semver -o yaml < catalog-template.yaml > catalog/catalog.yaml
opm validate catalog
'

buildah manifest rm "${catalog_image_ref}" 2>/dev/null || true
buildah build --squash --format=docker \
--file "${tmp_dir}/opm-storage/catalog.Dockerfile" \
--platform "${platforms}" \
--manifest "${catalog_image_ref}" \
"${tmp_dir}/opm-storage"

digestfile="$( mktemp )"
buildah manifest push --all --digestfile="${digestfile}" "${catalog_image_ref}"

echo "docker.io/rzetelskik/so-olm-catalog@$( strip_digestfile "${digestfile}" )"