#!/bin/bash

IMAGE_NAME="appliance-test:latest"
OUTPUT_DIR="./output"
ISO_FILE="agent-ove.x86_64.iso"
STORAGE_DRIVER="vfs"

# Use podman images to get the image ID for the specified tag
echo "Resolving image ID for: ${IMAGE_NAME}..."
IMAGE_ID=$(sudo podman images --storage-driver "${STORAGE_DRIVER}" --filter "reference=${IMAGE_NAME}" --format "{{.ID}}")

if [ -z "$IMAGE_ID" ]; then
    echo "Error: Failed to find image with tag ${IMAGE_NAME}. Exiting."
    exit 1
fi

echo "Found image ID: ${IMAGE_ID}"

mkdir -p "${OUTPUT_DIR}"

# Create a container from the image ID, and capture its ID
echo "Creating container from image ID: ${IMAGE_ID} with storage driver ${STORAGE_DRIVER}..."
CID=$(sudo podman create --storage-driver "${STORAGE_DRIVER}" "${IMAGE_ID}")

if [ -z "$CID" ]; then
    echo "Error: Failed to create container. Exiting."
    exit 1
fi

echo "Container created with ID: ${CID}"

# Copy the ISO file from the container to the host
echo "Copying ${ISO_FILE} from container to host's ${OUTPUT_DIR}..."
sudo podman cp --storage-driver "${STORAGE_DRIVER}" "${CID}:/release/${ISO_FILE}" "${OUTPUT_DIR}/${ISO_FILE}"
# Uncomment this if it useful to get all the files in /tmp
# sudo podman cp --storage-driver "${STORAGE_DRIVER}" "${CID}:/tmp/." "${OUTPUT_DIR}"

if [ $? -ne 0 ]; then
    echo "Error: Failed to copy the ISO file. The file may not exist in the container."
    sudo podman rm --storage-driver "${STORAGE_DRIVER}" "$CID" > /dev/null
    exit 1
fi

echo "Removing container..."
sudo podman rm --storage-driver "${STORAGE_DRIVER}" "$CID"

echo "Success! ISO file is available at ${OUTPUT_DIR}/${ISO_FILE}"
