#!/bin/bash
#
SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")" 
podman build -t agent-installer-utils-gotools -f "${SCRIPT_DIR}/../images/Containerfile.gotools-local" "${SCRIPT_DIR}/.."
