#!/usr/bin/env bash
#
# Copyright (C) 2026 ScyllaDB
#
# This script validates a version to release from the given ref and decides what the release publishes. For a version
# that already has its tag, it reads the tested image from the tag instead. It prints the results as key=value lines.

set -euExo pipefail
shopt -s inherit_errexit

if [[ "$#" -ne 4 ]]; then
  echo -e "Invalid arguments.\nUsage: ${0} <repository> <version> <ref> <dev_image_repo_ref>" >&2
  exit 1
fi

repository="${1}"
version="${2}"
ref="${3}"
dev_image_repo_ref="${4}"

readonly semver_pattern='^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$'
if [[ ! "${version}" =~ ${semver_pattern} ]]; then
  echo "Error: Version '${version}' is not MAJOR.MINOR.PATCH or MAJOR.MINOR.PATCH-PRERELEASE." >&2
  exit 1
fi
prerelease="$( [[ -n "${BASH_REMATCH[1]}" ]] && echo true || echo false )"
minor="$( cut -d . -f 1,2 <<< "${version}" )"

tag="v${version}"
released_tags="$( git ls-remote --tags --refs "https://github.com/${repository}.git" 'v*' | sed -e 's#.*refs/tags/##' )"
tag_exists=false
image_digest=''
image_ref=''
if grep -qxF "${tag}" <<< "${released_tags}"; then
  tag_exists=true
  # The release writes the digest of the tested image into the tag's annotation as an `Image-Digest:` trailer. A tag
  # without one wasn't made by the release, so nothing records that it passed the gate.
  tag_object="$( gh api "repos/${repository}/git/ref/tags/${tag}" --jq '.object | select(.type == "tag") | .sha' )"
  tag_commit=''
  if [[ -n "${tag_object}" ]]; then
    tag_json="$( gh api "repos/${repository}/git/tags/${tag_object}" )"
    tag_commit="$( jq -r .object.sha <<< "${tag_json}" )"
    image_digest="$( jq -r .message <<< "${tag_json}" | sed -n -e 's/^Image-Digest: //p' )"
  fi
  if [[ ! "${image_digest}" =~ ^sha256:[0-9a-f]{64}$ ]]; then
    echo "Error: Tag '${tag}' records no tested image, so it wasn't created by the release." >&2
    exit 1
  fi

  image_ref="${dev_image_repo_ref}@${image_digest}"

  # The image records the revision it was built from, and the digest covers that record.
  revision="$( skopeo inspect --raw "docker://${image_ref}" | jq -r '.annotations["org.opencontainers.image.revision"] // ""' )"
  if [[ "${revision}" != "${tag_commit}" ]]; then
    echo "Error: Tag '${tag}' points at ${tag_commit}, but its image was built from '${revision}'." >&2
    exit 1
  fi
# A new version is tagged on the commit it's released from, so that has to be the release branch of its minor.
elif [[ "${ref}" != "refs/heads/v${minor}" ]]; then
  echo "Error: Version '${version}' has to be released from branch 'v${minor}', not '${ref}'." >&2
  exit 1
fi

# `X.Y` follows the highest GA patch of its minor and `latest` the highest GA overall, so a patch of an older minor
# can't move `latest` backwards. Prereleases get neither.
rolling_tags=()
latest=false
if [[ "${prerelease}" == "false" ]]; then
  ga_tags="$( { grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' <<< "${released_tags}" || true; echo "${tag}"; } | sort --version-sort --unique )"
  if [[ "$( grep -F "v${minor}." <<< "${ga_tags}" | tail -n 1 )" == "${tag}" ]]; then
    rolling_tags+=( "${minor}" )
  fi
  if [[ "$( tail -n 1 <<< "${ga_tags}" )" == "${tag}" ]]; then
    rolling_tags+=( "latest" )
    latest=true
  fi
fi

echo "tag=${tag}"
echo "tag-exists=${tag_exists}"
echo "image-ref=${image_ref}"
echo "prerelease=${prerelease}"
echo "latest=${latest}"
echo "rolling-tags=${rolling_tags[*]}"
