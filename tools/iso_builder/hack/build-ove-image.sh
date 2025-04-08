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
    appliance_work_dir="${DIR_PATH}/$full_ocp_version/appliance"
    mkdir -p "${appliance_work_dir}"
    if [ ! -f "${appliance_work_dir}"/appliance-config.yaml ]; then
        echo "Creating appliance config..."
        local major_minor_patch_version=$(echo "\"$full_ocp_version\"" | jq -r 'split("-")[0]')
        cfg=${appliance_work_dir}/appliance-config.yaml
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
    else
        echo "Skip creating appliance config. Reusing ${appliance_work_dir}/appliance-config.yaml"
    fi
}

function build_live_iso() {
    if [ ! -f "${appliance_work_dir}"/appliance.iso ]; then
        echo "Building appliance ISO..."
        local pull_spec=quay.io/edge-infrastructure/openshift-appliance:latest
        $SUDO podman run --rm -it --privileged --pull always --net=host -v "${appliance_work_dir}"/:/assets:Z  "${pull_spec}" build live-iso
    else
        echo "Skip building appliance ISO. Reusing ${appliance_work_dir}/appliance.iso."
    fi
}

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
        sudo rsync -aH --info=progress2 "${appliance_mnt_dir}/" "${work_dir}/"
        sudo chown -R $(whoami):$(whoami) "${work_dir}/"
    fi
    volume_label=$(isoinfo -d -i "${appliance_work_dir}"/appliance.iso | grep "Volume id:" | cut -d' ' -f3-)
}

function setup_agent_artifacts() {
    local artifacts_dir="${DIR_PATH}"/$full_ocp_version/agent-artifacts
    local image=assisted-install-ui
    local pull_spec=registry.ci.openshift.org/ocp/4.19:"${image}"
    local image_dir="${work_dir}"/images/"${image}"
    if [ -d "${artifacts_dir}" ] && [ -f "${image_dir}/${image}.tar" ]; then
        echo "Skip preparing agent TUI artifacts. Reusing ${artifacts_dir}."
    else
        echo "Preparing agent TUI artifacts..."
        mkdir -p "${artifacts_dir}"
        local osarch
        if [ "${ARCH}" == "x86_64" ]; then
            osarch="amd64"
        else
            osarch="${ARCH}"
        fi

        local image_pull_spec=$(oc adm release info --registry-config="${PULL_SECRET_FILE}" --image-for=agent-installer-utils --filter-by-os=linux/"${osarch}" --insecure=true "${image_ref}")
        local files=("/usr/bin/agent-tui" "/usr/lib64/libnmstate.so.*")
        for f in "${files[@]}"; do
            echo "Extracting $f to ${artifacts_dir}"
            oc image extract --path="${f}:${artifacts_dir}" --registry-config="${PULL_SECRET_FILE}" --filter-by-os=linux/"${osarch}" --insecure=true --confirm "${image_pull_spec}"
        done

        # Make sure files could be executed
        chmod -R 555 "${artifacts_dir}"

        artfcts="${work_dir}"/agent-artifacts
        mkdir -p "${artfcts}"

        # Squash the directory to save space
        mksquashfs "${artifacts_dir}" "${artfcts}"/agent-artifacts.squashfs -comp xz -b 1M -Xdict-size 512K

        if [ ! -f "${image_dir}"/"${image}".tar ]; then
            # Copy assisted-installer-ui image to /images dir
            mkdir -p "${image_dir}"
            skopeo copy -q --authfile="${PULL_SECRET_FILE}" docker://"${pull_spec}" oci-archive:"${image_dir}"/"${image}".tar
        else
            echo "Skip pulling assisted-installer-ui image. Reusing ${image_dir}/${image}.tar."
        fi
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

function update_ignition() {
    local ignition_dir="${DIR_PATH}"/"${full_ocp_version}"/ignition
    mkdir -p "${ignition_dir}"
    local og_ignition="${ignition_dir}"/og_ignition.ign
    local updated_ignition="${ignition_dir}"/updated_ignition.ign

    if [ ! -f "${og_ignition}" ] || [ ! -f "${agent_ove_iso}" ]; then
        echo "Extracing ignition."
        coreos-installer iso ignition show "${agent_ove_iso}" | jq . >> "${og_ignition}"
    else
        echo "Skipping extracting ignition. Reusing ${og_ignition}."
    fi

    if [ ! -f "${updated_ignition}" ] || [ ! -f "${agent_ove_iso}" ]; then
        echo "Updating ignition..."
        local new_unit=$(cat <<EOF
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
        jq ".systemd.units += [$new_unit] | .storage.files += [$new_file]" "${og_ignition}" > "${updated_ignition}"
        echo "Embedding updated ignition into ISO..."
        coreos-installer iso ignition embed --force -i "${updated_ignition}" "${agent_ove_iso}"
    else
        echo "Skipping updating ignition. Reusing ${updated_ignition}."
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

    create_appliance_config
    build_live_iso
    extract_live_iso
    setup_agent_artifacts
    create_ove_iso
    update_ignition

    echo "Generated agent based installer OVE ISO at: $agent_ove_iso"
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
