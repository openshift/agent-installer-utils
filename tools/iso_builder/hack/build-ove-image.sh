#!/bin/bash

# Fail on unset variables and errors
set -euo pipefail
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOGDIR=/tmp/iso_builder/logs
source $SCRIPTDIR/logging.sh

function usage() {
    echo "----------------------------------------------------------------------------------------------------------------------"
    echo "ABI OVE Image Builder"
    echo
    echo "This script generates 'agent-ove-<arch>.iso' in the 'ove-assets' directory."
    echo "If the 'ove-assets' directory doesn't exist, it will be created at the current location."
    echo "The default architecture is x86_64."
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
    echo "  --rendezvousIP <IP>            (Optional) Rendezvous IP for the cluster"
    echo ""
    echo "Examples:"
    echo "$0 --pull-secret-file ~/pull_secret.json --release-image-url registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-03-18-173638"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.18.4"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.18.4 --arch x86_64"
    echo "$0 --pull-secret-file ~/pull_secret.json --ocp-version 4.18.4 --rendezvousIP 192.168.122.2"
    echo "Outputs:"
    echo "  - agent-ove-x86_64.iso: Bootable agent OVE ISO image."
    echo
    echo "Directory structure after running the script:"
    echo "  ./ove-assets/"
    echo "  └── agent-ove-x86_64.iso"
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
            --rendezvousIP) RENDEZVOUS_IP="$2"; shift ;;
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
}

function create_appliance_config() {
    echo "Creating appliance config..."
    local full_ocp_version
    
    if [ -n "${RELEASE_IMAGE_VERSION}" ]; then
        echo "Using OCP version ${RELEASE_IMAGE_VERSION}"
        full_ocp_version="${RELEASE_IMAGE_VERSION}"
    fi
    if [ -n "${RELEASE_IMAGE_URL}" ]; then
        echo "Using release image ${RELEASE_IMAGE_URL}"
        full_ocp_version=$(echo "\"$RELEASE_IMAGE_URL\"" | jq -r 'split(":")[1]')
    fi
    local major_minor_patch_version=$(echo "\"$full_ocp_version\"" | jq -r 'split("-")[0]')

    APPLIANCE_WORK_DIR="/tmp/iso_builder/appliance-assets-$full_ocp_version"
    mkdir -p "${APPLIANCE_WORK_DIR}"

# ToDo: Add rendezvousIp: user_specified_rendezvous_ip_address
    cat << EOF >> ${APPLIANCE_WORK_DIR}/appliance-config.yaml
apiVersion: v1beta1
kind: ApplianceConfig
diskSizeGB: 200
pullSecret: '$(cat "${PULL_SECRET_FILE}")'
userCorePass: core
stopLocalRegistry: false
enableDefaultSources: false
enableInteractiveFlow: true
operators:
  - catalog: registry.redhat.io/redhat/redhat-operator-index:v4.19
    packages:
      - name: mtv-operator
      - name: kubernetes-nmstate-operator
EOF

    if [ -n "${RELEASE_IMAGE_VERSION}" ]; then
        cat << EOF >> ${APPLIANCE_WORK_DIR}/appliance-config.yaml
ocpRelease:
  version: $major_minor_patch_version
  channel: candidate
  cpuArchitecture: $ARCH
EOF
    fi
    if [ -n "${RELEASE_IMAGE_URL}" ]; then
        cat << EOF >> ${APPLIANCE_WORK_DIR}/appliance-config.yaml
ocpRelease:
  version: $major_minor_patch_version
  url: $RELEASE_IMAGE_URL
EOF
    fi
}

function build_live_iso() {
    echo "Building appliance ISO..."
    local PULL_SPEC=quay.io/edge-infrastructure/openshift-appliance:latest
    sudo podman run --rm -it --privileged --pull always --net=host -v "${APPLIANCE_WORK_DIR}"/:/assets:Z  "${PULL_SPEC}" build live-iso
}

function extract_live_iso() {
    echo "Extracting ISO contents..."

    local READ_DIR="/tmp/iso_builder/appliance"
    mkdir -p "${READ_DIR}"

    if [ ! -f "${APPLIANCE_WORK_DIR}"/appliance.iso ]; then
        echo "Error: The appliance.iso disk image file is missing."
        echo "${APPLIANCE_WORK_DIR}"
        ls -lh "${APPLIANCE_WORK_DIR}"
        exit 1
    fi
    # Mount the ISO
    sudo mount -o loop "${APPLIANCE_WORK_DIR}"/appliance.iso "${READ_DIR}"
    VOLUME_LABEL=$(isoinfo -d -i "${APPLIANCE_WORK_DIR}"/appliance.iso | grep "Volume id:" | cut -d' ' -f3-)

    echo "Copying appliance ISO contents to a writable directory..."
    sudo rsync -aH --info=progress2 "${READ_DIR}/" "${WORK_DIR}/"

    sudo chown -R $(whoami):$(whoami) "${WORK_DIR}/"

    # Cleanup
    sudo umount "${READ_DIR}"
    sudo rm -rf "${READ_DIR}"

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

    local IMAGE_PULL_SPEC=$(oc adm release info --registry-config="${PULL_SECRET_FILE}" --image-for=agent-installer-utils --filter-by-os=linux/"${osarch}" --insecure=true "${RELEASE_IMAGE_VERSION}")
    
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
    sudo rm -rf "${ARTIFACTS_DIR}"/*
    sudo mv "${WORK_DIR}"/agent-artifacts.squashfs "${ARTIFACTS_DIR}"

    # copy the custom script for systemd
    sudo cp data/ove/data/files/usr/local/bin/setup-agent-tui.sh "${ARTIFACTS_DIR}"/setup-agent-tui.sh

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

    local UPDATED_IGNITION="${WORK_DIR}"/updated_ignition.ign
    jq ".systemd.units += [$NEW_UNIT]" "${OG_IGNITION}" > "${UPDATED_IGNITION}"

    echo "Embedding updated ignition into ISO..."
    coreos-installer iso ignition embed --force -i "${UPDATED_IGNITION}" "${AGENT_OVE_ISO}"
}

function cleanup() {
    sudo rm -rf "${WORK_DIR}"
}

function main()
{
    PULL_SECRET_FILE=""
    RELEASE_IMAGE_VERSION=""
    RELEASE_IMAGE_URL=""
    ARCH=""
    RENDEZVOUS_IP=""

    parse_inputs "$@"
    validate_inputs

    WORK_DIR="/tmp/iso_builder/ove-iso"
    mkdir -p "${WORK_DIR}"

    create_appliance_config
    build_live_iso
    extract_live_iso
    setup_agent_artifacts
    create_ove_iso
    update_ignition
    cleanup

    echo "Generated agent based installer OVE ISO at: $AGENT_OVE_ISO"
}

[[ $# -lt 3 ]] && usage
main "$@"