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
       local appliance_image=registry.ci.openshift.org/ocp/4.20:agent-preinstall-image-builder
        echo "Building appliance ISO (image: ${appliance_image})"
        $SUDO podman run --authfile "${PULL_SECRET_FILE}" --rm -it --privileged --pull always --net=host -v "${appliance_work_dir}"/:/assets:Z  "${appliance_image}" build live-iso --log-level debug
    else
        echo "Skip building appliance ISO. Reusing ${appliance_work_dir}/appliance.iso."
    fi
}

function finalize()
{
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
  *)
    echo "Error: The STEP variable must be 'all' or 'configure'." >&2
    exit 1
    ;;
esac

}

# Build agent installer OVI ISO
build "$@"
