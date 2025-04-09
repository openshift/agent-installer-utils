#!/bin/bash

# Fail on unset variables and errors
set -euo pipefail

function parse_inputs() {
    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --release-image-url) 
                if [[ -n "$RELEASE_IMAGE_VERSION" ]]; then
                    echo "Error: Cannot specify both --release-image-url and --ocp-version." >&2
                    usage
                    exit 1
                fi
                RELEASE_IMAGE_URL="$2"; shift ;;
            --ocp-version) 
                if [[ -n "$RELEASE_IMAGE_URL" ]]; then
                    echo "Error: Cannot specify both --release-image-url and --ocp-version." >&2
                    usage
                    exit 1
                fi
                RELEASE_IMAGE_VERSION="$2"; shift ;;
            --arch) ARCH="$2"; shift ;;
            --pull-secret-file) PULL_SECRET_FILE="$2"; shift ;;
            --ssh-key-file) SSH_KEY_FILE="$2"; shift ;;
            --dir) DIR_PATH="$2"; shift ;;
            *) 
                echo "Unknown parameter: $1" >&2
                usage
                exit 1 ;;
        esac
        shift
    done
}

function validate_inputs() {
    if [[ -z "${RELEASE_IMAGE_VERSION:-}" && -z "${RELEASE_IMAGE_URL:-}" ]]; then
        echo "Error: Either OpenShift version (--ocp-version) or release image URL (--release-image-url) must be provided." >&2
        usage
        exit 1
    fi

    if [[ -z "${PULL_SECRET_FILE:-}" ]]; then
        echo "Error: Pull secret file is required." >&2
        usage
        exit 1
    fi

    if [[ -n "$PULL_SECRET_FILE" && ! -f "$PULL_SECRET_FILE" ]]; then
        echo "Error: File $PULL_SECRET_FILE does not exist." >&2
        exit 1
    fi

    # Use default architecture if not provided
    # To do: Validate if provided arch is a valid one [AMD64 (x86_64), s390x (IBM System Z), ppc64 little endian (Power PC) or arm (aarch64)]
    if [[ -z "${ARCH:-}" ]]; then
        ARCH="x86_64"
        echo "Warning: Architecture not specified. Using default architecture: $ARCH."
    fi

    # Ensure that the OCP version is in the format `x.y.z`
    if [[ -n "$RELEASE_IMAGE_VERSION" && ! "$RELEASE_IMAGE_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Error: OpenShift version (--ocp-version) must be in the format major.minor.patch (e.g., 4.18.4)." >&2
        usage
        exit 1
    fi
    if [[ -n "$SSH_KEY_FILE" && ! -f "$SSH_KEY_FILE" ]]; then
        echo "File $SSH_KEY_FILE does not exist." >&2
        exit 1
    fi

    if [[ -z "${DIR_PATH:-}" ]]; then
        DIR_PATH="/tmp/iso_builder"
        echo "Directory path not specified. Using default location: $DIR_PATH."
    else
        echo "ISOBuilder assets will be stored in: $DIR_PATH."
    fi
}

function setup_vars() {
    full_ocp_version=""
    image_ref=""
    if [ -n "${RELEASE_IMAGE_VERSION}" ]; then
        echo "Using OCP version ${RELEASE_IMAGE_VERSION}"
        full_ocp_version="${RELEASE_IMAGE_VERSION}"
        image_ref="${RELEASE_IMAGE_VERSION}"
    fi
    if [ -n "${RELEASE_IMAGE_URL}" ]; then
        echo "Using release image ${RELEASE_IMAGE_URL}"
        full_ocp_version=$(skopeo inspect --authfile $PULL_SECRET_FILE docker://$RELEASE_IMAGE_URL | jq -r '.Labels["io.openshift.release"]')
        image_ref="${RELEASE_IMAGE_URL}"
    fi

    major_minor_patch_version=$(echo "\"$full_ocp_version\"" | jq -r 'split("-")[0]')
    ove_dir="${DIR_PATH}/$full_ocp_version/ove"
    work_dir="${ove_dir}/work"
    output_dir="${ove_dir}/output"
    agent_ove_iso="${output_dir}"/agent-ove."${ARCH}".iso
    LOGDIR=${DIR_PATH}/$full_ocp_version/logs
    

    mkdir -p "${DIR_PATH}"
    mkdir -p "${output_dir}"
}

function usage() {
    echo "----------------------------------------------------------------------------------------------------------------------"
    echo "ABI OVE Image Builder"
    echo
    echo "This script generates 'agent-ove.<arch>.iso' in the '/tmp/iso_builder/<OCP_VERSION>/ove/output' directory."
    echo "The default architecture is x86_64."
    echo "The default directory path is /tmp/iso_builder.\n"
    echo
    echo "Usage:"
    echo "  ./hack/build-ove-image.sh [OPTIONS]"
    echo ""
    echo "Required Options:"
    echo "  --pull-secret-file <path>      Path to the pull secret file (e.g., ~/pull_secret.json)"
    echo ""
    echo "One of the following must be specified:"
    echo "  --release-image-url <url>      Specifies the OpenShift release image URL, supporting both image tags and SHA digests."
    echo "                                 (e.g., registry.ci.openshift.org/ocp/release:<tag> or registry.ci.openshift.org/ocp/release@sha256:<digest>)."
    echo "                                 Recommended for CI/Nightly builds used in development and testing."
    echo ""
    echo "  --ocp-version <version>        Specifies the OpenShift version in major.minor.patch format (e.g., 4.19.0)."
    echo "                                 Recommended for general availability (GA) OCP versions 4.19 and later."
    echo ""
    echo "Optional:"
    echo "  --arch <architecture>          Target CPU architecture (default: x86_64)"
    echo "  --ssh-key-file <path>          Path to the SSH key file (e.g., ~/.ssh/id_rsa)"
    echo "  --dir <path>                   Path for ISOBuilder assets (default: /tmp/iso_builder)"
    echo ""
    echo "Examples:"
    echo "$0 --pull-secret-file ~/pull_secret.json --release-image-url registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-04-01-173804"
    echo "$0 --pull-secret-file ~/pull_secret.json --release-image-url registry.ci.openshift.org/ocp/release@sha256:4242c71d4bb159d1be56785216086d0e8573dd3aeb9e7ea30874e985b6fc76d9"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.19.1"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.19.1 --arch x86_64 --ssh-key-file ~/.ssh/id_rsa --dir ~/iso_builder"
    echo "Outputs:"
    echo "  - agent-ove.x86_64.iso: Bootable agent OVE ISO image."
    echo
    echo "Directory structure after running the script:"
    echo "  /tmp/iso_builder/"
    echo "  └── 4.19.0-0.ci-2025-04-01-173804"
    echo "      ├── agent-artifacts"
    echo "      ├── appliance"
    echo "      ├── ignition"
    echo "      ├── logs"
    echo "      └── ove"
    echo "          ├── output"
    echo "          │   └── agent-ove.x86_64.iso"
    echo "          └── work"
    echo "----------------------------------------------------------------------------------------------------------------------"
    exit 1
}
