#!/bin/bash

# This script will create a docker image, tag it based on the contents
# of the VERSION file and push it to the docker repository.

set -euo pipefail

DRY_RUN=1
readonly VERSION="v$(cat VERSION)"

if [ "${VERSION}" != $(make print-version) ]; then
  echo "Version in 'VERSION' file ${VERSION} & git tag $(make print-version) don't match! Follow the release steps mentioned in https://github.com/biome-sh/biome-operator/blob/master/doc/release-process.md"
  exit 1
fi

run() {
  if [[ "${DRY_RUN}" -eq 1 ]]; then
    printf '%s\n' "$*"
  else
    "$@"
  fi
}

main() {
  if [[ $# -gt 0 ]]; then
    # Let's err on the side of caution.
    if [[ "${1}" == '-r' ]] || [[ "${1}" == '--run' ]]; then
      DRY_RUN=0
    fi
  fi

  if [[ "${DRY_RUN}" -eq 1 ]]; then
    echo "Script is running in dry run mode. Following commands will be executed if you pass the -r or --run flag:"
  fi

  run make image
  run docker push biomesh/biome-operator:"${VERSION}"
  run docker tag biomesh/biome-operator:"${VERSION}" biomesh/biome-operator:latest
  run docker push biomesh/biome-operator:latest
}

main "$@"
