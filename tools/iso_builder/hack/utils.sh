#!/bin/bash

set -euo pipefail

export CUSTOM_OPENSHIFT_INSTALLER_PATH="${CUSTOM_OPENSHIFT_INSTALLER_PATH:-}"

function patch_openshift_install_release_version() {
    local version=$1
    local installer=${APPLIANCE_WORK_DIR}/openshift-install
    cp ${CUSTOM_OPENSHIFT_INSTALLER_PATH}/bin/openshift-install ${installer}

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