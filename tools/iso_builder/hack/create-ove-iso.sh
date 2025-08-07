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

appliance_work_dir="${DIR_PATH}/$full_ocp_version/appliance"

function extract_live_iso() {
    local appliance_mnt_dir="$appliance_work_dir/mnt"
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
        # Mount the ISO
        $SUDO mount -o loop "${appliance_work_dir}"/appliance.iso "${appliance_mnt_dir}"
    fi
    if [ -d "${work_dir}" ]; then
        echo "Skip copying extracted appliance ISO contents to a writable directory. Reusing ${work_dir}."
    else
        mkdir -p "${work_dir}"
        echo "Copying extracted appliance ISO contents to a writable directory."
        $SUDO rsync -aH --info=progress2 "${appliance_mnt_dir}/" "${work_dir}/"
        $SUDO chown -R $(whoami):$(whoami) "${work_dir}/"
        $SUDO umount ${appliance_mnt_dir}
    fi
    volume_label=$(xorriso -indev "${appliance_work_dir}"/appliance.iso -toc 2>/dev/null | awk -F',' '/ISO session/ {print $4}' | xargs)
}

function setup_agent_artifacts() {
    local image=assisted-install-ui
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

function build()
{
    start_time=$(date +%s)

    if [ "$(id -u)" -eq 0 ]; then
        SUDO=""
    else
        SUDO="sudo"
    fi

    extract_live_iso
    setup_agent_artifacts
    create_ove_iso

    if [ "${ARCH}" == "x86_64" ]; then
        # The release ISO is large, so users will often prefer copying it to a USB stick 
        # rather than mounting it via virtual media on the BMC.
        #
        # However, booting from USB mass storage requires a Master Boot Record (MBR), 
        # whereas optical drives rely on the El Torito ISO9660 boot extension.
        #
        # To support both boot methods—USB and virtual media—we augment the El Torito ISO 
        # with an MBR, making it hybrid-bootable:
        /usr/bin/isohybrid --uefi $agent_ove_iso
    fi

    cp -v $agent_ove_iso $appliance_work_dir

    echo "Generated agent based installer OVE ISO at: $agent_ove_iso and copied to $appliance_work_dir"
    end_time=$(date +%s)
    elapsed_time=$(($end_time - $start_time))
    minutes=$((elapsed_time / 60))
    seconds=$((elapsed_time % 60))

    if [[ $minutes -gt 0 && $seconds -gt 0 ]]; then
        echo "ISOBuilder execution time: ${minutes}m ${seconds}s"
    fi
}

# Build agent installer OVI ISO
build "$@"
