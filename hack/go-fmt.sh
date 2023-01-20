#!/bin/bash

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"

if [ "$IS_CONTAINER" != "" ]; then
  gci_repo=github.com/daixiang0/gci
  for TARGET in "${@}"; do
    find "${TARGET}" -name '*.go' ! -path '*/vendor/*' ! -path '*/.build/*' -exec gofmt -s -w {} \+
    find "${TARGET}" -name '*.go' ! -path '*/vendor/*' ! -path '*/.build/*' -exec go run "$gci_repo" write -s standard -s default -s "prefix(github.com/openshift)" -s blank --skip-generated {} \+
  done
  git diff --exit-code
else
	if ! podman image exists agent-installer-utils-gotools; then
		source "${SCRIPT_DIR}/build-gotool.sh"
	fi
  podman run --rm \
    --env IS_CONTAINER=TRUE \
    --volume "${PWD}:/go/src/github.com/openshift/agent-installer-utils:z" \
    --workdir /go/src/github.com/openshift/agent-installer-utils \
    agent-installer-utils-gotools \
    ./hack/go-fmt.sh "${@}"
fi
