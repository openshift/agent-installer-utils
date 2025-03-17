#!/bin/bash
set -e

TAR_FILE="/run/media/iso/images/assisted-install-ui/assisted-install-ui.tar"
IMAGE_TAG="localhost/agent-installer-ui:latest"

echo "Loading image from $TAR_FILE..."
podman load -q -i "$TAR_FILE"

if [ $? -eq 0 ]; then
    echo "Image loaded successfully!"
    IMAGE_ID=$(podman images --format '{{.ID}}' | head -n 1)
    if [ -n "$IMAGE_ID" ]; then
        podman tag "$IMAGE_ID" "$IMAGE_TAG"
        if [ $? -eq 0 ]; then
            echo "Image tagged successfully as $IMAGE_TAG"
        else
            echo "Failed to tag the image."
            exit 1
        fi
    else
        echo "Failed to retrieve image ID after loading."
        exit 1
    fi
else
    echo "Failed to load the image from $TAR_FILE."
    exit 1
fi
