#!/bin/bash

# Fail on unset variables and errors
set -euo pipefail

function parse_inputs() {
    ARCH="x86_64"
    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --release-image) RELEASE_VERSION="$2"; shift ;;
            --arch) ARCH="$2"; shift ;;
            --pull-secret) PULL_SECRET="$2"; shift ;;
            --rendezvousIP) RENDEZVOUS_IP="$2"; shift ;;
            *) echo "Unknown parameter: $1"; exit 1 ;;
        esac
        shift
    done
}

function validate_inputs() {
    if [[ -z "${RELEASE_VERSION:-}" || -z "${ARCH:-}" || -z "${PULL_SECRET:-}" ]]; then
        echo "Error: OpenShift version, architecture and pull secret are required."
        exit 1
    fi
    if [ ! -f "$PULL_SECRET" ]; then
        echo "File $PULL_SECRET does not exist." >&2
        exit 1
    fi
}

parse_inputs "$@"
validate_inputs

OCP_VERSION=$(echo $RELEASE_VERSION | awk -F ':' '{print $2}' | awk -F'-' '{print $1}')
TMP_dir=/tmp/iso_builder/${OCP_VERSION}
work_dir="${TMP_dir}/ove/work"

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOGDIR=${TMP_dir}/logs
source $SCRIPTDIR/logging.sh

echo "Using release version: $RELEASE_VERSION"
echo "Using arch: $ARCH"
function usage() {
    echo "----------------------------------------------------------------------------------------------------------------------"
    echo "ABI OVE Image Builder"
    echo
    echo "This script generates 'agent-ove-<ocp-version>-<arch>.iso' in the '/tmp/iso_builder/<ocp-version>/ove/output' directory."
    echo "The default architecture is x86_64."
    echo
    echo "Usage:"
    echo "$0 --release-image <openshift-release> --pull-secret <pull-secret> --arch [architecture] --rendezvousIP [rendezvousIP]"
    echo
    echo "Examples:"
    echo "$0 --release-image registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-02-26-035445 --pull-secret ~/pull_secret.json"
    echo "$0 --release-image registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-02-26-035445 --pull-secret ~/pull_secret.json --arch x86_64"
    echo "$0 --release-image registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-02-26-035445 --pull-secret ~/pull_secret.json --rendezvousIP 192.168.122.2"
    echo
    echo "Outputs:"
    echo "  - agent-ove-4.19.0-x86_64.iso: Bootable agent OVE ISO image."
    echo
    echo "Directory structure after running the script:"
    echo "/tmp/iso_builder"
    echo "└── 4.19.0"
    echo "  ├── agent-artifacts"
    echo "  ├── appliance"
    echo "  │   ├── mnt"
    echo "  │   └── work"
    echo "  ├── ignition"
    echo "  ├── logs"
    echo "  └── ove"
    echo "    ├── output"
    echo "    │   └── agent-ove-4.19.0-x86_64.iso"
    echo "    └── work"
    echo "----------------------------------------------------------------------------------------------------------------------"
    exit 1
}



function create_appliance_config() {
    local OCP_VERSION=$1
    local arch=$2
    local pullsecret=$3

    APPLIANCE_WORK_DIR="/tmp/iso_builder/$OCP_VERSION/appliance/work"
    mkdir -p "${APPLIANCE_WORK_DIR}"
    if [ ! -f "${APPLIANCE_WORK_DIR}"/appliance-config.yaml ]; then
        echo "Creating appliance config ${APPLIANCE_WORK_DIR}/appliance-config.yaml."
# ToDo: Add rendezvousIp: user_specified_rendezvous_ip_address
cat >"${APPLIANCE_WORK_DIR}/appliance-config.yaml" <<EOF
apiVersion: v1beta1
kind: ApplianceConfig
ocpRelease:
  version: "${OCP_VERSION}"
  channel: candidate
  cpuArchitecture: "${arch}"
diskSizeGB: 200
pullSecret: '$(cat "${pullsecret}")'
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
    else
        echo "Skip creating appliance config. Reusing ${APPLIANCE_WORK_DIR}/appliance-config.yaml"
    fi
}

function patch_openshift_install_release_version() {
    local version=$1
    local installer=${APPLIANCE_WORK_DIR}/openshift-install
    OPENSHIFT_INSTALL_PATH=/home/test/installer/bin/
    cp ${OPENSHIFT_INSTALL_PATH}/openshift-install ${installer}

    local res=$(grep -oba ._RELEASE_VERSION_LOCATION_.XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX ${installer})
    local location=${res%%:*}

    # If the release marker was found then it means that the version is missing
    if [[ ! -z ${location} ]]; then
        echo "Patching openshift-install with version ${version}"
        printf "${version}\0" | dd of=${installer} bs=1 seek=${location} conv=notrunc &> /dev/null 
        ${installer} version
    else
        echo "Version already patched"
    fi
} 

function build_live_iso() {
    if [ ! -f "${APPLIANCE_WORK_DIR}"/appliance.iso ]; then
        echo "Building appliance ISO."

        # To Do: For dev/debug purposes, build a local installer and patch
        patch_openshift_install_release_version "${OCP_VERSION}"

        local pull_spec="quay.io/edge-infrastructure/openshift-appliance:latest"
        # Remove the local podman image if it exists, then pull and run the latest image from the registry
        image_id=$(sudo podman images -f "reference=$pull_spec" -q)
        if [ -n "$image_id" ]; then
            echo "Removing the local podman image $pull_spec"
            sudo podman rmi $image_id
        fi
        sudo podman run --rm -it --privileged --pull always --net=host -v "${APPLIANCE_WORK_DIR}"/:/assets:Z --env OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=${RELEASE_VERSION} "${pull_spec}" \
        build live-iso --debug-base-ignition --log-level=debug
    else
        echo "Skip building appliance ISO. Reusing ${APPLIANCE_WORK_DIR}/appliance.iso."
    fi
}

function extract_live_iso() {
    local OCP_VERSION=$1
    local appliance_mnt_dir="/tmp/iso_builder/$OCP_VERSION/appliance/mnt"
    if [ -d "${appliance_mnt_dir}" ]; then
        echo "Skip extracting appliance ISO. Reusing ${appliance_mnt_dir}."
    else
        echo "Extracting appliance ISO."
        mkdir -p "${appliance_mnt_dir}"
        if [ ! -f "${APPLIANCE_WORK_DIR}"/appliance.iso ]; then
            echo "Error: The appliance.iso disk image file is missing."
            echo "${APPLIANCE_WORK_DIR}"
            ls -lh "${APPLIANCE_WORK_DIR}"
            exit 1
        fi
        # Mount the ISO
        sudo mount -o loop "${APPLIANCE_WORK_DIR}"/appliance.iso "${appliance_mnt_dir}"
    fi
    if [ -d "${work_dir}" ]; then
        echo "Skip copying extracted appliance ISO contents to a writable directory. Reusing ${work_dir}."
    else
        mkdir -p "${work_dir}"
        echo "Copying extracted appliance ISO contents to a writable directory."
        sudo rsync -aH --info=progress2 "${appliance_mnt_dir}/" "${work_dir}/"
        sudo chown -R $(whoami):$(whoami) "${work_dir}/"
    fi
    VOLUME_LABEL=$(isoinfo -d -i "${APPLIANCE_WORK_DIR}"/appliance.iso | grep "Volume id:" | cut -d' ' -f3-)
}

function setup_agent_artifacts() {
    local tmpdir=$2
    local artifacts_dir="${tmpdir}"/agent-artifacts
    if [ -d "${artifacts_dir}" ]; then
        echo "Skip preparing agent TUI artifacts. Reusing ${artifacts_dir}."
    else
        echo "Preparing agent TUI artifacts."
        mkdir -p "${artifacts_dir}"
        local osarch
        if [ "${ARCH}" == "x86_64" ]; then
            osarch="amd64"
        else
            osarch="${ARCH}"
        fi
        local pull_secret=$1
        local image_pull_spec=$(oc adm release info --registry-config="${pull_secret}" --image-for=agent-installer-utils --filter-by-os=linux/"${osarch}" --insecure=true "${RELEASE_VERSION}")
    
        local files=("/usr/bin/agent-tui" "/usr/lib64/libnmstate.so.*")
        for f in "${files[@]}"; do
            echo "Extracting $f to ${artifacts_dir}"
            oc image extract --path="${f}:${artifacts_dir}" --registry-config="${pull_secret}" --filter-by-os=linux/"${osarch}" --insecure=true --confirm "${image_pull_spec}"
        done

        # Make sure files could be executed
        chmod -R 555 "${artifacts_dir}"

        artfcts="${work_dir}"/agent-artifacts
        mkdir -p "${artfcts}"

        # Squash the directory to save space
        sudo mksquashfs "${artifacts_dir}" "${artfcts}"/agent-artifacts.squashfs -comp xz -b 1M -Xdict-size 512K

        # copy the custom scripts for systemd
        sudo cp data/ove/data/files/usr/local/bin/*.sh "${artfcts}"/

        # Copy assisted-installer-ui image to /images dir
        local image=assisted-install-ui
        local pull_spec=registry.ci.openshift.org/ocp/4.19:"${image}"
        local image_dir="${work_dir}"/images/"${image}"
        mkdir -p "${image_dir}"
        skopeo copy -q --authfile="${pull_secret}" docker://"${pull_spec}" oci-archive:"${image_dir}"/"${image}".tar
    fi
}

function create_ove_iso() {
    if [ ! -f "${AGENT_OVE_ISO}" ]; then
        local boot_image=$work_dir/images/efiboot.img
        if [ -f "${boot_image}" ]; then
            local size=$(stat --format="%s" "${boot_image}")
            # Calculate the number of 2048-byte sectors needed for the file
            # Add 2047 to round up any remaining bytes to a full sector
            local boot_load_size=$(( (size + 2047) / 2048 ))
        else
            echo "Error: Run 'make cleanall OCP_RELEASE_IMAGE=<OCP_RELEASE_IMAGE>'"
            exit 1
        fi

        echo "Creating ${AGENT_OVE_ISO}."
        xorriso -as mkisofs \
            -o "${AGENT_OVE_ISO}" \
            -J -R -V "${VOLUME_LABEL}" \
            -b isolinux/isolinux.bin \
            -c isolinux/boot.cat \
            -no-emul-boot -boot-load-size 4 -boot-info-table \
            -eltorito-alt-boot \
            -e images/efiboot.img \
            -no-emul-boot -boot-load-size "${boot_load_size}" \
            "${work_dir}"
    fi
}

function update_ignition() {
    local ignition_dir="${TMP_dir}"/ignition
    mkdir -p "${ignition_dir}"
    local og_ignition="${ignition_dir}"/og_ignition.ign
    local updated_ignition="${ignition_dir}"/updated_ignition.ign

    if [ ! -f "${og_ignition}" ] || [ ! -f "${AGENT_OVE_ISO}" ]; then
        echo "Extracing ignition."
        coreos-installer iso ignition show "${AGENT_OVE_ISO}" | jq . >> "${og_ignition}"
    else
        echo "Skipping extracting ignition. Reusing ${og_ignition}."
    fi

    if [ ! -f "${updated_ignition}" ] || [ ! -f "${AGENT_OVE_ISO}" ]; then
        echo "Updating ignition."
        local new_unit=$(cat <<EOF
{
    "contents": $(cat data/ove/data/systemd/agent-setup-tui.service | sed -z 's/\n$//' | jq -Rs .),
    "name": "agent-setup-tui.service",
    "enabled": true
},
{
    "contents": $(cat data/ove/data/systemd/agent-load-assisted-webui.service | sed -z 's/\n$//' | jq -Rs .),
    "name": "agent-load-assisted-webui.service",
    "enabled": true
}
EOF
)
    jq ".systemd.units += [$new_unit]" "${og_ignition}" > "${updated_ignition}"

    echo "Embedding updated ignition into ISO."
    coreos-installer iso ignition embed --force -i "${updated_ignition}" "${AGENT_OVE_ISO}"
    else
        echo "Skipping updating ignition. Reusing ${updated_ignition}."
    fi
}

function build()
{
    RENDEZVOUS_IP=""

    
    output_dir="${TMP_dir}/ove/output"

    mkdir -p "${TMP_dir}"
    mkdir -p "${output_dir}"

    AGENT_OVE_ISO="${output_dir}"/agent-ove-"${OCP_VERSION}"-"${ARCH}".iso

    create_appliance_config "${OCP_VERSION}" "${ARCH}" "${PULL_SECRET}"
    build_live_iso
    extract_live_iso "${OCP_VERSION}"
    setup_agent_artifacts "${PULL_SECRET}" "${TMP_dir}"
    create_ove_iso
    update_ignition
    echo "Agent based installer OVE ISO is avaialble at: $AGENT_OVE_ISO"
}

[[ $# -lt 2 ]] && usage
build "$@"