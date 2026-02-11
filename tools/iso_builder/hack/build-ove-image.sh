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
export MIRROR_PATH=""
export REGISTRY_CERT=""

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

        if [[ -n "$MIRROR_PATH" ]]; then
            cat << EOF >> ${cfg}
mirrorPath: /mirror
EOF
        fi

    else
        echo "Skip creating appliance config. Reusing ${appliance_work_dir}/appliance-config.yaml"
    fi
}

function build_live_iso() {
    if [ ! -f "${appliance_work_dir}"/appliance.iso ]; then
        local appliance_image=registry.ci.openshift.org/ocp/${major_minor_version}:agent-preinstall-image-builder
        echo "Building appliance ISO (image: ${appliance_image})"

        # Build the podman run command with optional mirror path
        local podman_cmd="$SUDO podman run --authfile \"${PULL_SECRET_FILE}\" --rm -it --privileged --pull always --net=host -v \"${appliance_work_dir}\"/:/assets:Z"
        local appliance_cmd="build live-iso --log-level debug"

        # Add mirror path mount if provided
        if [[ -n "${MIRROR_PATH}" ]]; then
            echo "Using pre-mirrored images from: ${MIRROR_PATH}"
            podman_cmd="${podman_cmd} -v \"${MIRROR_PATH}\":/mirror:Z"
        fi

        # Add registry certificate mount if provided (for custom registries with self-signed certs)
        if [[ -n "${REGISTRY_CERT}" ]]; then
            echo "Mounting registry certificate for TLS verification: ${REGISTRY_CERT}"
            podman_cmd="${podman_cmd} -v \"${REGISTRY_CERT}\":/etc/pki/ca-trust/source/anchors/registry.crt:Z,ro"
            # Override entrypoint to run update-ca-trust before openshift-appliance
            # Must include --dir /assets to match the volume mount
            # Workaround: Move appliance.iso to /assets if it was created in wrong location
            podman_cmd="${podman_cmd} --entrypoint sh"
            appliance_cmd="-c 'update-ca-trust && /openshift-appliance --dir /assets ${appliance_cmd} ; if [ -f /appliance.iso ] && [ ! -f /assets/appliance.iso ]; then mv /appliance.iso /assets/appliance.iso; fi'"
        fi

        set -x
        eval "${podman_cmd} \"${appliance_image}\" ${appliance_cmd}"
        set +x
    else
        echo "Skip building appliance ISO. Reusing ${appliance_work_dir}/appliance.iso."
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
