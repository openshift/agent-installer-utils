#!/bin/bash

# Fail on unset variables and errors
set -euo pipefail

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
    echo "  --pull-secret <path>           Path to the pull secret file (e.g., ~/pull_secret.json)"
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
    echo "$0 --pull-secret ~/pull_secret.json --release-image-url registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-03-18-173638"
    echo "$0 --pull-secret ~/pull_secret.json --ocp-version 4.18.4"
    echo "$0 --pull-secret ~/pull_secret.json --ocp-version 4.18.4 --arch x86_64"
    echo "$0 --pull-secret ~/pull_secret.json --ocp-version 4.18.4 --rendezvousIP 192.168.122.2"
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
                if [[ -n "$RELEASE_VERSION" ]]; then
                    echo "Error: Cannot specify both --release-image-url and --ocp-version." >&2
                    exit 1
                fi
                RELEASE_IMAGE_URL="$2"; shift ;;
            --ocp-version) 
                if [[ -n "$RELEASE_IMAGE_URL" ]]; then
                    echo "Error: Cannot specify both --release-image-url and --ocp-version." >&2
                    exit 1
                fi
                RELEASE_VERSION="$2"; shift ;;
            --arch) ARCH="$2"; shift ;;
            --pull-secret) PULL_SECRET="$2"; shift ;;
            --rendezvousIP) RENDEZVOUS_IP="$2"; shift ;;
            *) 
                echo "Unknown parameter: $1" >&2
                exit 1 ;;
        esac
        shift
    done
}

function validate_inputs() {
    if [[ -z "${RELEASE_VERSION:-}" && -z "${RELEASE_IMAGE_URL:-}" ]]; then
        echo "Error: Either OpenShift version (--ocp-version) or release image URL (--release-image-url) must be provided." >&2
        exit 1
    fi

    if [[ -z "${PULL_SECRET:-}" ]]; then
        echo "Error: Pull secret is required." >&2
        exit 1
    fi

    if [ ! -f "$PULL_SECRET" ]; then
        echo "Error: File $PULL_SECRET does not exist." >&2
        exit 1
    fi

    # Use default architecture if not provided
    # To do: Validate if provided arch is a valid one [AMD64 (x86_64), s390x (IBM System Z), ppc64 little endian (Power PC) or arm (aarch64)]
    if [[ -z "${ARCH:-}" ]]; then
        ARCH="x86_64"
        echo "Warning: Architecture not specified. Using default architecture: $ARCH."
    fi

    # Ensure that the OCP version is in the format `x.y.z`
    if [[ -n "$RELEASE_VERSION" && ! "$RELEASE_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Error: RELEASE_VERSION must be in the format major.minor.patch (e.g., 4.18.4)." >&2
        exit 1
    fi
}

function create_appliance_config() {
    echo "Creating appliance config..."
    local version=$1
    local release_url=$2
    local arch=$3
    local pullSecret=$4

    local major_minor_patch_version=$(echo "\"$FULL_OCP_VERSION\"" | jq -r 'split("-")[0]')

    APPLIANCE_WORK_DIR="/tmp/iso_builder/appliance-assets-$FULL_OCP_VERSION"
    mkdir -p "${APPLIANCE_WORK_DIR}"

# ToDo: Add rendezvousIp: user_specified_rendezvous_ip_address
  cat >"${APPLIANCE_WORK_DIR}/appliance-config.yaml" <<EOF
apiVersion: v1beta1
kind: ApplianceConfig
ocpRelease:
  # Always include version, either directly from --ocp-version or extracted from --release-image-url
  $( [[ -n "$version" ]] && printf "version: \"%s\"\n" "$version" ) # Use value from --ocp-version, if provided
  $( [[ -z "$version" && -n "$release_url" ]] && printf "version: \"%s\"\n" "$major_minor_patch_version" )  # Extract value from --release-image-url, if --ocp-version not specified
  $( [[ -n "$version" ]] && printf "channel: candidate\n" ) # Add channel only if --ocp-version is provided
  $( [[ -n "$version" ]] && printf "cpuArchitecture: \"%s\"\n" "$arch" ) # Add cpuArchitecture only if --ocp-version is provided
  $( [[ -n "$release_url" ]] && printf "url: \"%s\"\n" "$release_url")  # Add URL if provided
diskSizeGB: 200
pullSecret: '$(cat "${pullSecret}")'
imageRegistry:
  uri: quay.io/libpod/registry:2.8
userCorePass: core
stopLocalRegistry: false
enableDefaultSources: false
enableInteractiveFlow: true
operators:
  - catalog: registry.redhat.io/redhat/redhat-operator-index:v4.19
    packages:
      - name: mtv-operator
EOF
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
    local PULL_SECRET=$1
    local OSARCH
    if [ "${ARCH}" == "x86_64" ]; then
        OSARCH="amd64"
    else
        OSARCH="${ARCH}"
    fi

    local ARTIFACTS_DIR="${WORK_DIR}"/agent-artifacts
    mkdir -p "${ARTIFACTS_DIR}"

    local IMAGE_PULL_SPEC=$(oc adm release info --registry-config="${PULL_SECRET}" --image-for=agent-installer-utils --filter-by-os=linux/"${OSARCH}" --insecure=true "${RELEASE_VERSION}")
    
    local FILES=("/usr/bin/agent-tui" "/usr/lib64/libnmstate.so.*")
    for FILE in "${FILES[@]}"; do
        echo "Extracting $FILE..."
        oc image extract --path="${FILE}:${ARTIFACTS_DIR}" --registry-config="${PULL_SECRET}" --filter-by-os=linux/"${OSARCH}" --insecure=true --confirm "${IMAGE_PULL_SPEC}"
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
    
    skopeo copy -q --authfile="${PULL_SECRET}" docker://"${PULL_SPEC}" oci-archive:"${IMAGE_DIR}"/"${IMAGE}".tar
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
    PULL_SECRET=""
    RELEASE_VERSION=""
    RELEASE_IMAGE_URL=""
    ARCH=""
    RENDEZVOUS_IP=""
    FULL_OCP_VERSION=""

    parse_inputs "$@"
    validate_inputs

    WORK_DIR="/tmp/iso_builder/ove-iso"
    mkdir -p "${WORK_DIR}"

    # Check if appropriate parameters are provided and invoke the function
    if [[ -n "${RELEASE_VERSION}" && -z "${RELEASE_IMAGE_URL}" ]]; then
        FULL_OCP_VERSION="${RELEASE_VERSION}"
        echo "Using OCP version ${RELEASE_VERSION}"
        create_appliance_config "${RELEASE_VERSION}" "" "${ARCH}" "${PULL_SECRET}"
    elif [[ -n "${RELEASE_IMAGE_URL}" && -z "${RELEASE_VERSION}" ]]; then
        echo "Using release image ${RELEASE_IMAGE_URL}"
        FULL_OCP_VERSION=$(echo "\"$RELEASE_IMAGE_URL\"" | jq -r 'split(":")[1]')
        create_appliance_config "" "${RELEASE_IMAGE_URL}" "${ARCH}" "${PULL_SECRET}"
    fi
    build_live_iso
    extract_live_iso
    setup_agent_artifacts "${PULL_SECRET}"
    create_ove_iso
    update_ignition
    cleanup

    echo "Generated agent based installer OVE ISO at: $AGENT_OVE_ISO"
}

[[ $# -lt 3 ]] && usage
main "$@"