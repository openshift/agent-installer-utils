#!/bin/bash

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"

if [ "$IS_CONTAINER" != "" ]; then
	readarray -t packages < <(go list ./... | grep -v vendor)
	printf 'Go vetting %s\n' "${packages[@]}"
	go vet "${packages[@]}"
else
	if ! podman image exists agent-installer-utils-gotools; then
		source "${SCRIPT_DIR}/build-gotool.sh"
	fi
  podman run --rm \
    --env IS_CONTAINER=TRUE \
    --volume "${PWD}:/go/src/github.com/openshift/agent-installer-utils:z" \
    --workdir /go/src/github.com/openshift/agent-installer-utils \
    agent-installer-utils-gotools \
    ./hack/go-vet.sh "${@}"
fi;
