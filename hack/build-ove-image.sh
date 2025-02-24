#!/bin/bash

# Fail on unset variables and errors
set -euo pipefail

function usage() {
    echo "----------------------------------------------------------------------------------------------------------------------"
    echo "ABI OVE Image Builder"
    echo
    echo "This script generates 'agent-ove-<arch>.iso' in the 'ove-assets' directory."
    echo "If the 'ove-assets' directory doesn't exist, it will be created at the current location."
    echo
    echo "Usage:"
    echo "$0 --version <openshift-release> --arch <architecture> --pull-secret <pull-secret> --ip [rendezvousIP] --ssh-key [sshKey] --operators [operators]"
    echo
    echo "Examples:"
    echo "$0 --version registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-02-26-035445 --arch x86_64 --pull-secret ~/pull_secret.json"
    echo "$0 --version 4.18.0-rc.5 --arch x86_64 --pull-secret ~/pull_secret.json --ip 192.168.122.2"
    echo "$0 --version 4.18.0-rc.5 --arch x86_64 --pull-secret ~/pull_secret.json --ssh-key ~/.ssh/idrsa.pub"
    echo "$0 --version 4.18.0-rc.5 --arch x86_64 --pull-secret ~/pull_secret.json --operators operator1,operator2"
    echo
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
            --version) VERSION="$2"; shift ;;
            --arch) ARCH="$2"; shift ;;
            --pull-secret) PULL_SECRET="$2"; shift ;;
            --ip) RENDEZVOUS_IP="$2"; shift ;;
            --ssh-key) SSH_KEY="$2"; shift ;;
            --operators) OPERATORS="$2"; shift ;;
            *) echo "Unknown parameter: $1"; exit 1 ;;
        esac
        shift
    done
}

function validate_inputs() {
    if [[ -z "${VERSION:-}" || -z "${ARCH:-}" || -z "${PULL_SECRET:-}" ]]; then
        echo "Error: OpenShift version, architecture and pull secret are required."
        exit 1
    fi
    if [ ! -f "$PULL_SECRET" ]; then
        echo "File $PULL_SECRET does not exist." >&2
        exit 1
    fi
}

# ToDo: Once appliance PR https://github.com/openshift/appliance/pull/335 is merged, change this
# to checkout from the main branch
function clone_appliance_repo() {
    local repo_url="https://github.com/danielerez/appliance.git"
    local branch_name="live_iso"
    local clone_dir="./appliance"

    if [[ ! -d "$clone_dir" ]]; then 
        echo "Cloning repository from $repo_url (branch: $branch_name)..."
        git clone --branch "$branch_name" "$repo_url" "$clone_dir"
    else
        echo "Repository already cloned in $clone_dir. Pulling latest changes..."
        pushd "$clone_dir" || exit
        git fetch
        git checkout "$branch_name"
        git pull
        popd
    fi
}

function build_appliance() {
    pushd appliance
    make build
    popd
}

function keygen()
{
  echo "WARN" "No public SSH key provided, generating a new one..." "-n"

  TEMP_DIR="${TMP_APPLIANCE_DIR}/ove_$(date +%Y%m%d%H%M%S)_$(uuidgen)"
  mkdir -p "${TEMP_DIR}"

  ssh-keygen -q -t ed25519 -N '' -f ${TEMP_DIR}/agent_ed25519
  [ $? -ne 0 ] && echo "ERRO" "'ssh-keygen' failure. Aborting execution." && exit 2 || echo "SUCC"
  SSHKEY="${TEMP_DIR}/agent_ed25519.pub"
  echo "DEBG" "$(cat ${SSHKEY})"
}

function create_appliance_config() {
    local RELEASE_VERSION=$1
    local VERSION=$(echo $RELEASE_VERSION | awk -F ':' '{print $2}')
    local ARCH=$2
    local PULLSECRET=$3
    local SSH_KEY=$4
    
    if [ -z "$SSH_KEY" ]; then
        keygen
    else
        SSHKEY="$SSH_KEY"
        echo "DEBG" "Using provided SSH key: $(cat ${SSHKEY})"
    fi

  cat >"${TMP_APPLIANCE_DIR}/appliance-config.yaml" <<EOF
apiVersion: v1beta1
kind: ApplianceConfig
ocpRelease:
  version: "${VERSION}"
  channel: candidate
  cpuArchitecture: "${ARCH}"
diskSizeGB: 200
pullSecret: '$(cat "${PULLSECRET}")'
sshKey: $(cat "${SSHKEY}")
imageRegistry:
  uri: quay.io/libpod/registry:2.8
userCorePass: core
stopLocalRegistry: false
enableDefaultSources: false
EOF
}

function build_live_iso() {
    sudo podman run --rm -it --privileged --net=host -v "${TMP_APPLIANCE_DIR}"/:/assets:Z quay.io/edge-infrastructure/openshift-appliance:latest build live-iso
}

function prepare_agent_tui_artifacts() {
    if [ "${ARCH}" == "x86_64" ]; then
        OSARCH="amd64"
    else
        OSARCH="${ARCH}"
    fi

    ARTIFACTS_DIR="${TMP_APPLIANCE_DIR}/agent-artifacts"
    mkdir -p "${ARTIFACTS_DIR}"

    IMAGE_PULL_SPEC=$(oc adm release info --image-for=agent-installer-utils --filter-by-os=linux/"${OSARCH}" --insecure=true "${VERSION}")
    
    FILES=("/usr/bin/agent-tui" "/usr/lib64/libnmstate.so.*")
    for FILE in "${FILES[@]}"; do
        echo "Extracting $FILE"
        oc image extract --path="$FILE:${ARTIFACTS_DIR}" --filter-by-os=linux/$OSARCH --insecure=true --confirm "${IMAGE_PULL_SPEC}"
    done

    # Make sure files could be executed
    chmod -R 555 "${ARTIFACTS_DIR}"

    # Squash the directory to save space
    mksquashfs "${ARTIFACTS_DIR}" "${SQUASH_FILE}" -comp xz -b 1M -Xdict-size 512K
}

function extract_iso_and_copy_agent_tui_artifacts() {
    echo "Extracting ISO contents..."
    mkdir -p "${READ_DIR}" "${WORK_DIR}"
    mount -o loop "${APPLIANCE_ISO_PATH}" "${READ_DIR}"
    # Copy ISO contents to a writable directory
    rsync -av "${READ_DIR}"/ "${WORK_DIR}"/
    umount "${READ_DIR}"

    # Copy the squashed agent artifacts to the ISO
    AGENT_ARTIFACTS_DIR="${WORK_DIR}"/agent-artifacts
    mkdir -p "${AGENT_ARTIFACTS_DIR}"
    cp "${SQUASH_FILE}" "${AGENT_ARTIFACTS_DIR}"/

    # copy the custom script for systemd
    AGENT_SCRIPTS_DIR="${WORK_DIR}"/usr/local/bin
    mkdir -p "${AGENT_SCRIPTS_DIR}"
    cp data/ove/data/files/usr/local/bin/setup-agent-tui.sh "${AGENT_ARTIFACTS_DIR}"/setup-agent-tui.sh
}

function rebuild_iso() {
    echo "Rebuilding ISO..."
    volume_label=$(isoinfo -d -i "${APPLIANCE_ISO_PATH}" | grep "Volume id:" | cut -d' ' -f3-)

    xorriso -as mkisofs \
        -o "${AGENT_OVE_ISO}" \
        -J -R -V "${volume_label}" \
        -b isolinux/isolinux.bin \
        -c isolinux/boot.cat \
        -no-emul-boot -boot-load-size 4 -boot-info-table \
        -eltorito-alt-boot \
        -e images/efiboot.img \
        -no-emul-boot -boot-load-size 2489 \
        "${WORK_DIR}"
}

function  extract_original_igntion() {
    echo "Extracing ignition..."
    coreos-installer iso ignition show "${AGENT_OVE_ISO}" | jq . >> "${OG_IGNITION}"
}

function update_ignition() {
    echo "Updating ignition..."

    NEW_UNIT=$(cat <<EOF
{
    "contents": $(cat data/ove/data/systemd/agent-setup-tui.service | sed -z 's/\n$//' | jq -Rs .),
    "name": "agent-setup-tui.service",
    "enabled": true
}
EOF
)

    jq ".systemd.units += [$NEW_UNIT]" "${OG_IGNITION}" > "${UPDATED_IGNITION}"

    echo "Embedding updated ignition into ISO..."
    coreos-installer iso ignition embed --force -i "${UPDATED_IGNITION}" "${AGENT_OVE_ISO}"
}

function cleanup() {
    rm -rf "${READ_DIR}"
    rm -rf "${WORK_DIR}"
    rm -rf "${TMP_APPLIANCE_DIR}"
}

function main()
{
    RENDEZVOUS_IP=""
    SSH_KEY=""
    OPERATORS=""

    parse_inputs "$@"
    validate_inputs

    OVE_ASSETS_DIR="$(pwd)/ove-assets"
    mkdir -p "${OVE_ASSETS_DIR}"
    AGENT_OVE_ISO="${OVE_ASSETS_DIR}"/agent-ove-"${ARCH}".iso

    TMP_APPLIANCE_DIR="/tmp/appliance"
    mkdir -p "${TMP_APPLIANCE_DIR}"
    APPLIANCE_ISO_PATH="${TMP_APPLIANCE_DIR}"/appliance.iso
    SQUASH_FILE="${TMP_APPLIANCE_DIR}"/agent-artifacts.squashfs
    OG_IGNITION="${TMP_APPLIANCE_DIR}"/og_ignition.ign
    UPDATED_IGNITION="${TMP_APPLIANCE_DIR}"/updated_ignition.ign

    READ_DIR="/mnt/appliance/iso"              
    WORK_DIR="/mnt/ove/iso"                                

    clone_appliance_repo
    build_appliance
    create_appliance_config "$VERSION" "$ARCH" "$PULL_SECRET" "$SSH_KEY"
    build_live_iso

    prepare_agent_tui_artifacts
    extract_iso_and_copy_agent_tui_artifacts

    rebuild_iso

    extract_original_igntion
    update_ignition

    cleanup

    echo "Generated agent based installer OVE ISO at: $AGENT_OVE_ISO"
}

[[ $# -lt 1 ]] && usage
main "$@"