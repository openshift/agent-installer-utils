#!/bin/bash

# Fail on unset variables and errors 
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $SCRIPTDIR/helper.sh

export SSH_KEY_FILE=""
export PULL_SECRET_FILE=""
export RELEASE_IMAGE_VERSION=""
export RELEASE_IMAGE_URL=""
export ARCH=""
export DIR_PATH=""

# Check user provided params
[[ $# -lt 2 ]] && usage
parse_inputs "$@"
validate_inputs
setup_vars

source $SCRIPTDIR/logging.sh

function create_appliance_config() {
    # The appliance-config.yaml has been copied to the correct directory
    if [ -f "${appliance_work_dir}"/appliance-config.yaml ]; then
        echo "Creating appliance config..."
        cfg=${appliance_work_dir}/appliance-config.yaml

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

        if [[ -n "$PULL_SECRET_FILE" ]]; then
            cat << EOF >> ${cfg}
pullSecret: '$(cat "${PULL_SECRET_FILE}")'
EOF
        fi

        if [[ -n "$SSH_KEY_FILE" ]]; then
            cat << EOF >> ${cfg}
sshKey: '$(cat "${SSH_KEY_FILE}")'
EOF
        fi

    else
        echo "Skip creating appliance config. Reusing ${appliance_work_dir}/appliance-config.yaml"
    fi
}

function build_live_iso() {
    if [ ! -f "${appliance_work_dir}"/appliance.iso ]; then
       local appliance_image=quay.io/edge-infrastructure/openshift-appliance@sha256:82836d7a7e257e83c5065fcdb731a09b8bec7228be3ee8c9122b4b5c45463b73
        echo "Building appliance ISO (image: ${appliance_image})"
        patch_openshift_install_release_version "${full_ocp_version}"
        $SUDO podman run --authfile "${PULL_SECRET_FILE}" --rm -it --privileged --net=host -v "${appliance_work_dir}"/:/assets:Z  --env OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE="${RELEASE_IMAGE_URL}" "${appliance_image}" build live-iso --log-level debug --debug-base-ignition
    else
        echo "Skip building appliance ISO. Reusing ${appliance_work_dir}/appliance.iso."
    fi
}

function patch_openshift_install_release_version() {
    local version=$1
    local installer="${DIR_PATH}/$full_ocp_version/appliance/openshift-install"
    CUSTOM_OPENSHIFT_INSTALLER_PATH=/home/test/go/src/github.com/openshift/installer
    echo "Using custom openshift installer from ${CUSTOM_OPENSHIFT_INSTALLER_PATH}"
    cp "${CUSTOM_OPENSHIFT_INSTALLER_PATH}"/bin/openshift-install "${installer}"

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

function extract_live_iso() {
    local appliance_mnt_dir="${appliance_work_dir}/isomnt"
    if [ -d "${appliance_mnt_dir}" ]; then
        echo "Skip extracting appliance ISO. Reusing ${appliance_mnt_dir}."
    else
        echo "Extracting ISO contents..."
        mkdir -p "${appliance_mnt_dir}"

        if [ ! -f "${appliance_work_dir}"/appliance.iso ]; then
            echo "Error: Issue with appliance. The appliance.iso disk image file is missing."
            echo "${appliance_work_dir}"
            ls -lh "${appliance_work_dir}"
            exit 1
        fi
        # Mount the ISO when not in a container
        if [ ${appliance_work_dir} != '/' ]; then
            $SUDO mount -o loop "${appliance_work_dir}"/appliance.iso "${appliance_mnt_dir}"
        fi
    fi
    if [ -d "${work_dir}" ]; then
        echo "Skip copying extracted appliance ISO contents to a writable directory. Reusing ${work_dir}."
    else
        mkdir -p "${work_dir}"
        if [ ${appliance_work_dir} == '/' ]; then
            # Use osirrox to extract the ISO without mounting it
            $SUDO osirrox -indev "${appliance_work_dir}"/appliance.iso -extract / "${appliance_mnt_dir}"
        fi
        echo "Copying extracted appliance ISO contents to a writable directory."
        $SUDO rsync -aH --info=progress2 "${appliance_mnt_dir}/" "${work_dir}/"
        $SUDO chown -R $(whoami):$(whoami) "${work_dir}/"
        if mountpoint -q ${appliance_mnt_dir}; then
            $SUDO umount ${appliance_mnt_dir}
        fi
    fi
    volume_label=$(xorriso -indev "${appliance_work_dir}"/appliance.iso -toc 2>/dev/null | awk -F',' '/ISO session/ {print $4}' | xargs)
}

function setup_agent_artifacts() {
    local image=agent-installer-ui
    local pull_spec=registry.ci.openshift.org/ocp/4.20:"${image}"
    local image_dir="${work_dir}"/images/"${image}"

    if [ ! -f "${image_dir}"/"${image}".tar ]; then
        # Copy assisted-installer-ui image to /images dir
        echo "skopeo copy UI image to oci-archive:${image_dir}/${image}.tar"
        mkdir -p "${image_dir}"
        skopeo copy -q --authfile="${PULL_SECRET_FILE}" docker://"${pull_spec}" oci-archive:"${image_dir}"/"${image}".tar
    else
        echo "Skip pulling assisted-installer-ui image. Reusing ${image_dir}/${image}.tar."
    fi
     #copy custom assisted service image in OVE ISO
    local image=assisted-service-late-binding
    local pull_spec=quay.io/ppinjark/assisted-service:late-binding
    local image_dir="${work_dir}"/images/"${image}"
    mkdir -p "${image_dir}"
    quay_authfile=/home/test/go/src/github.com/openshift/agent-installer-utils/tools/iso_builder/hack/quay-login.json
    skopeo copy -q --authfile="${quay_authfile}" docker://"${pull_spec}" oci-archive:"${image_dir}"/"${image}".tar
}

function create_ove_iso() {
    if [ ! -f "${agent_ove_iso}" ]; then
        local boot_image=$work_dir/images/efiboot.img
        if [ -f "${boot_image}" ]; then
            local size=$(stat --format="%s" "${boot_image}")
            # Calculate the number of 2048-byte sectors needed for the file
            # Add 2047 to round up any remaining bytes to a full sector
            local boot_load_size=$(( (size + 2047) / 2048 ))
        else
            echo "Error: Clean /tmp/iso_builder directory."
            exit 1
        fi

        echo "Creating ${agent_ove_iso}."
        xorriso -as mkisofs \
        -o "${agent_ove_iso}" \
        -J -R -V "${volume_label}" \
        -b isolinux/isolinux.bin \
        -c isolinux/boot.cat \
        -no-emul-boot -boot-load-size 4 -boot-info-table \
        -eltorito-alt-boot \
        -e images/efiboot.img \
        -no-emul-boot -boot-load-size "${boot_load_size}" \
        "${work_dir}"
    fi
}

function finalize()
{
    cp "${appliance_work_dir}"/appliance.iso "${agent_ove_iso}"
    echo "Generated agent based installer OVE ISO at: $agent_ove_iso"
    end_time=$(date +%s)
    elapsed_time=$(($end_time - $start_time))
    minutes=$((elapsed_time / 60))
    seconds=$((elapsed_time % 60))

    if [[ $minutes -gt 0 && $seconds -gt 0 ]]; then
        echo "ISOBuilder execution time: ${minutes}m ${seconds}s"
    fi
}

function build()
{
    start_time=$(date +%s)

    if [ "$(id -u)" -eq 0 ]; then
        SUDO=""
    else
        SUDO="sudo"
    fi

    case "$STEP" in
    "all")
      # Copy the configuration files for the version
      cp -rf $(pwd)/config/${major_minor_version}/* ${appliance_work_dir}

      create_appliance_config
      build_live_iso
      finalize
      ;;
    "configure")
      create_appliance_config
      ;;
    "create-iso")
      finalize

      # Remove directory to limit size of container
      rm -r ${work_dir}

      # Move to top-level dir for easier retrieval
      mv -v ${agent_ove_iso} ${appliance_work_dir}
      ;;
  *)
    echo "Error: The STEP variable must be 'all', 'configure', or 'create-iso'." >&2
    exit 1
    ;;
esac

}

# Build agent installer OVI ISO
build "$@"
