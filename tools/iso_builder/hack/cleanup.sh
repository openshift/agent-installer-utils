#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage:"
  echo "  cleanup.sh clean-tmp-dir"
  echo "  cleanup.sh clean-appliance-dir"
}

cleanall() {
  echo "Cleaning directory: /tmp/iso_builder"
  rm -rf /tmp/iso_builder
  echo "Done."
}

clean-appliance-temp-dir() {
  read -p "Enter the appliance directory path to clean (e.g., /tmp/iso_builder/<OCP_VERSION>/appliance): " dir;

  if [[ ! -d "$dir" ]]; then
    echo "Error: Directory '$dir' does not exist."
    exit 1
  fi

  if [[ ! -d "$dir/temp" ]]; then
    echo "Error: The directory '$dir/' does not contain a 'temp' folder."
    echo "This may indicate that the appliance directory has already been cleaned, or the provided path is incorrect."
    exit 1
  fi

  echo "Cleaning directory: $dir"
  sudo podman run --rm -it --privileged --net=host -v "$dir:/assets:Z" quay.io/edge-infrastructure/openshift-appliance:latest clean
  echo "Done."
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

case "$1" in
  cleanall)
    cleanall
    ;;
  clean-appliance-temp-dir)
    clean-appliance-temp-dir
    ;;
  *)
    echo "Error: Unknown command '$1'"
    usage
    exit 1
    ;;
esac
