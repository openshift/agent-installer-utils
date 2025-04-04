#!/bin/bash

# Fail on unset variables and errors
set -euo pipefail
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOGDIR=/tmp/iso_builder/logs
source $SCRIPTDIR/logging.sh

export SSH_KEY_FILE=""
export PULL_SECRET_FILE=""
export RELEASE_IMAGE_VERSION=""
export RELEASE_IMAGE_URL=""
export ARCH=""
export DIR_PATH=""

function usage() {
    echo "----------------------------------------------------------------------------------------------------------------------"
    echo "ABI OVE Image Builder"
    echo
    echo "This script generates 'agent-ove-<arch>.iso' in the 'ove-assets' directory."
    echo "If the 'ove-assets' directory doesn't exist, it will be created at the current location."
    echo "The default architecture is x86_64."
    echo "The default directory path is /tmp/iso_builder.\n"
    echo
    echo "Usage:"
    echo "  ./hack/build-ove-image.sh [OPTIONS]"
    echo ""
    echo "Required Options:"
    echo "  --pull-secret-file <path>           Path to the pull secret file (e.g., ~/pull_secret.json)"
    echo ""
    echo "One of the following must be specified:"
    echo "  --release-image-url <url>      OpenShift release image URL (e.g., registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-03-18-173638)"
    echo "  --ocp-version <version>        OpenShift version in major.minor.patch format (e.g., 4.18.4)"
    echo ""
    echo "Optional:"
    echo "  --arch <architecture>          Target CPU architecture (default: x86_64)"
    echo "  --dir <path>                   Path for ISOBuilder assets (default: /tmp/iso_builder)"
    echo ""
    echo "Examples:"
    echo "$0 --pull-secret-file ~/pull_secret.json --release-image-url registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-03-18-173638"
    echo "$0 --pull-secret-file ~/pull_secret.json --release-image-url registry.ci.openshift.org/ocp/release@sha256:1a991852031c0a2825c6ae2280bfd2c2b9b4564b59aef14e68b3ece3e47c8448"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.18.4"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.18.4 --arch x86_64 --dir ~/iso_builder"
    echo "Outputs:"
    echo "  - agent-ove-x86_64.iso: Bootable agent OVE ISO image."
    echo
    echo "Directory Structure After Running the Script:"
    echo " When no directory location is specified:"
    echo "  ./ove-assets/"
    echo "  └── agent-ove-x86_64.iso"
    echo " When a directory location is specified (e.g., --dir ~/iso_builder)"
    echo "~/iso_builder/"
    echo "    └── agent-ove-x86_64.iso"
    echo "----------------------------------------------------------------------------------------------------------------------"
    exit 1
}

function parse_inputs() {
    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --release-image-url) 
                if [[ -n "$RELEASE_IMAGE_VERSION" ]]; then
                    echo "Error: Cannot specify both --release-image-url and --ocp-version." >&2
                    exit 1
                fi
                RELEASE_IMAGE_URL="$2"; shift ;;
            --ocp-version) 
                if [[ -n "$RELEASE_IMAGE_URL" ]]; then
                    echo "Error: Cannot specify both --release-image-url and --ocp-version." >&2
                    exit 1
                fi
                RELEASE_IMAGE_VERSION="$2"; shift ;;
            --arch) ARCH="$2"; shift ;;
            --pull-secret-file) PULL_SECRET_FILE="$2"; shift ;;
            --ssh-key-file) SSH_KEY_FILE="$2"; shift ;;
            --dir) DIR_PATH="$2"; shift ;;
            *) 
                echo "Unknown parameter: $1" >&2
                exit 1 ;;
        esac
        shift
    done
}

function validate_inputs() {
    if [[ -z "${RELEASE_IMAGE_VERSION:-}" && -z "${RELEASE_IMAGE_URL:-}" ]]; then
        echo "Error: Either OpenShift version (--ocp-version) or release image URL (--release-image-url) must be provided." >&2
        exit 1
    fi

    if [[ -z "${PULL_SECRET_FILE:-}" ]]; then
        echo "Error: Pull secret file is required." >&2
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
    FULL_OCP_VERSION=""
    IMAGE_REF=""
    if [ -n "${RELEASE_IMAGE_VERSION}" ]; then
        echo "Using OCP version ${RELEASE_IMAGE_VERSION}"
        FULL_OCP_VERSION="${RELEASE_IMAGE_VERSION}"
        IMAGE_REF="${RELEASE_IMAGE_VERSION}"
    fi
    if [ -n "${RELEASE_IMAGE_URL}" ]; then
        echo "Using release image ${RELEASE_IMAGE_URL}"
        FULL_OCP_VERSION=$(skopeo inspect --authfile $PULL_SECRET_FILE docker://$RELEASE_IMAGE_URL | jq -r '.Labels["io.openshift.release"]')
        IMAGE_REF="${RELEASE_IMAGE_URL}"
    fi
    major_minor_patch_version=$(echo "\"$FULL_OCP_VERSION\"" | jq -r 'split("-")[0]')
    APPLIANCE_WORK_DIR="$DIR_PATH/appliance-assets-$FULL_OCP_VERSION"
}

function create_appliance_config() {
    echo "Creating appliance config..."
    mkdir -p "${APPLIANCE_WORK_DIR}"

    cfg=${APPLIANCE_WORK_DIR}/appliance-config.yaml

    cat << EOF >> ${cfg}  
apiVersion: v1beta1
kind: ApplianceConfig
diskSizeGB: 200
pullSecret: '$(cat "${PULL_SECRET_FILE}")'
userCorePass: core
stopLocalRegistry: false
enableDefaultSources: false
enableInteractiveFlow: true
EOF

    if [[ -n "$SSH_KEY_FILE" ]]; then
        cat << EOF >> ${cfg} 
sshKey: '$(cat "${SSH_KEY_FILE}")'
EOF
    fi

    cat << EOF >> ${cfg} 
operators:
  - catalog: registry.redhat.io/redhat/redhat-operator-index:v4.19
    packages:
      - name: mtv-operator
      - name: kubernetes-nmstate-operator
      - name: node-healthcheck-operator
      - name: node-maintenance-operator
      - name: fence-agents-remediation
      - name: self-node-remediation
      - name: cluster-kube-descheduler-operator
EOF

    if [ -n "${RELEASE_IMAGE_VERSION}" ]; then
        cat << EOF >> ${cfg}
ocpRelease:
  version: $major_minor_patch_version
  channel: candidate
  cpuArchitecture: $ARCH
EOF
    fi
    if [ -n "${RELEASE_IMAGE_URL}" ]; then
        cat << EOF >> ${cfg}
ocpRelease:
  version: $major_minor_patch_version
  url: $RELEASE_IMAGE_URL
EOF
    fi
}

function build_live_iso() {
    echo "Building appliance ISO..."
    local PULL_SPEC=quay.io/edge-infrastructure/openshift-appliance:latest
    $SUDO podman run --rm -it --privileged --pull always --net=host -v "${APPLIANCE_WORK_DIR}"/:/assets:Z  "${PULL_SPEC}" build live-iso
}

function extract_live_iso() {
    echo "Extracting ISO contents..."

    local READ_DIR="$DIR_PATH/appliance"
    mkdir -p "${READ_DIR}"

    if [ ! -f "${APPLIANCE_WORK_DIR}"/appliance.iso ]; then
        echo "Error: The appliance.iso disk image file is missing."
        echo "${APPLIANCE_WORK_DIR}"
        ls -lh "${APPLIANCE_WORK_DIR}"
        exit 1
    fi
    # Mount the ISO
    $SUDO mount -o loop "${APPLIANCE_WORK_DIR}"/appliance.iso "${READ_DIR}"
    VOLUME_LABEL=$(isoinfo -d -i "${APPLIANCE_WORK_DIR}"/appliance.iso | grep "Volume id:" | cut -d' ' -f3-)

    echo "Copying appliance ISO contents to a writable directory..."
    $SUDO rsync -aH --info=progress2 "${READ_DIR}/" "${WORK_DIR}/"

    $SUDO chown -R $(whoami):$(whoami) "${WORK_DIR}/"

    # Cleanup
    $SUDO umount "${READ_DIR}"
    $SUDO rm -rf "${READ_DIR}"

}

function setup_agent_artifacts() {
    echo "Preparing agent TUI artifacts..."
    local osarch
    if [ "${ARCH}" == "x86_64" ]; then
        osarch="amd64"
    else
        osarch="${ARCH}"
    fi

    local ARTIFACTS_DIR="${WORK_DIR}"/agent-artifacts
    mkdir -p "${ARTIFACTS_DIR}"

    local IMAGE_PULL_SPEC=$(oc adm release info --registry-config="${PULL_SECRET_FILE}" --image-for=agent-installer-utils --filter-by-os=linux/"${osarch}" --insecure=true "${IMAGE_REF}")
    
    local FILES=("/usr/bin/agent-tui" "/usr/lib64/libnmstate.so.*")
    for FILE in "${FILES[@]}"; do
        echo "Extracting $FILE..."
        oc image extract --path="${FILE}:${ARTIFACTS_DIR}" --registry-config="${PULL_SECRET_FILE}" --filter-by-os=linux/"${osarch}" --insecure=true --confirm "${IMAGE_PULL_SPEC}"
    done

    # Make sure files could be executed
    chmod -R 555 "${ARTIFACTS_DIR}"

    # Squash the directory to save space
    mksquashfs "${ARTIFACTS_DIR}" "${WORK_DIR}"/agent-artifacts.squashfs -comp xz -b 1M -Xdict-size 512K

    # Cleanup directory and save only one archieved file
    $SUDO rm -rf "${ARTIFACTS_DIR}"/*
    $SUDO mv "${WORK_DIR}"/agent-artifacts.squashfs "${ARTIFACTS_DIR}"


    # Copy assisted-installer-ui image to /images dir
    local IMAGE=assisted-install-ui
    local PULL_SPEC=registry.ci.openshift.org/ocp/4.19:"${IMAGE}"
    local IMAGE_DIR="${WORK_DIR}"/images/"${IMAGE}"
    mkdir -p "${IMAGE_DIR}"
    
    skopeo copy -q --authfile="${PULL_SECRET_FILE}" docker://"${PULL_SPEC}" oci-archive:"${IMAGE_DIR}"/"${IMAGE}".tar
}

function create_ove_iso() {
    local OUTPUT_DIR="$(pwd)/ove-assets"
    mkdir -p "${OUTPUT_DIR}"
    AGENT_OVE_ISO="${OUTPUT_DIR}"/agent-ove-"${ARCH}".iso

    echo "Creating ${AGENT_OVE_ISO}..."
    local BOOT_IMAGE="${WORK_DIR}/images/efiboot.img"
    local SIZE=$(stat --format="%s" "${BOOT_IMAGE}")
    # Calculate the number of 2048-byte sectors needed for the file
    # Add 2047 to round up any remaining bytes to a full sector
    local BOOT_LOAD_SIZE=$(( ("${SIZE}" + 2047) / 2048 ))

    xorriso -as mkisofs \
        -o "${AGENT_OVE_ISO}" \
        -J -R -V "${VOLUME_LABEL}" \
        -b isolinux/isolinux.bin \
        -c isolinux/boot.cat \
        -no-emul-boot -boot-load-size 4 -boot-info-table \
        -eltorito-alt-boot \
        -e images/efiboot.img \
        -no-emul-boot -boot-load-size "${BOOT_LOAD_SIZE}" \
        "${WORK_DIR}"
}

function update_ignition() {
    echo "Extracing ignition..."
    local OG_IGNITION="${WORK_DIR}"/og_ignition.ign
    
    coreos-installer iso ignition show "${AGENT_OVE_ISO}" | jq . >> "${OG_IGNITION}"

    echo "Updating ignition..."

    local NEW_UNIT=$(cat <<EOF
{
    "contents": $(cat data/ove/data/systemd/agent-setup-tui.service | sed -z 's/\n$//' | jq -Rs .),
    "name": "agent-setup-tui.service",
    "enabled": true
}
EOF
)

    local encoded_content=$(base64 -w 0 data/ove/data/files/usr/local/bin/setup-agent-tui.sh)
    local new_file=$(cat <<EOF
{
        "group": {},
        "overwrite": true,
        "path": "/usr/local/bin/setup-agent-tui.sh",
        "user": {
          "name": "root"
        },
        "contents": {
          "source": "data:text/plain;charset=utf-8;base64,$encoded_content",
          "verification": {}
        },
        "mode": 365
    }
EOF
)
    local UPDATED_IGNITION="${WORK_DIR}"/updated_ignition.ign
    jq ".systemd.units += [$NEW_UNIT] | .storage.files += [$new_file]" "${OG_IGNITION}" > "${UPDATED_IGNITION}"

    echo "Embedding updated ignition into ISO..."
    coreos-installer iso ignition embed --force -i "${UPDATED_IGNITION}" "${AGENT_OVE_ISO}"
}

function cleanup() {
    $SUDO rm -rf "${WORK_DIR}"
}

function main()
{
    start_time=$(date +%s)
    PULL_SECRET_FILE=""
    RELEASE_IMAGE_VERSION=""
    RELEASE_IMAGE_URL=""
    ARCH=""

    parse_inputs "$@"
    validate_inputs
    setup_vars

    WORK_DIR="$DIR_PATH/ove-iso"
    mkdir -p "${WORK_DIR}"

    if [ "$(id -u)" -eq 0 ]; then
        SUDO=""
    else
        SUDO="sudo"
    fi

    create_appliance_config
    build_live_iso
    extract_live_iso
    setup_agent_artifacts
    create_ove_iso
    update_ignition
    cleanup

    echo "Generated agent based installer OVE ISO at: $AGENT_OVE_ISO"
    end_time=$(date +%s)
    elapsed_time=$((end_time - start_time))
    minutes=$((elapsed_time / 60))
    seconds=$((elapsed_time % 60))

    echo "ISOBuilder execution time: ${minutes}m ${seconds}s" 
}

[[ $# -lt 2 ]] && usage
main "$@"