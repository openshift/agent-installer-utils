#!/bin/bash

# Define variables
IMAGE_NAME="localhost/appliance-test"
FILE_NAME="agent-ove.x86_64.iso"
OUTPUT_DIR="output-iso"

# Check if the output directory exists, create it if not
if [ ! -d "$OUTPUT_DIR" ]; then
    echo "Creating directory: $OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
fi

# Get the image ID from the image name
echo "Finding image ID for name: $IMAGE_NAME..."
IMAGE_ID=$(sudo buildah images --storage-driver vfs --all --json | jq -r '.[] | select(.names != null and (.names[] == "localhost/appliance-test:latest")) | .id')

if [ -z "$IMAGE_ID" ]; then
    echo "Error: Could not find image ID for name '$IMAGE_NAME'. Please check the image name."
    exit 1
fi

echo "Found image ID: $IMAGE_ID"

# 1. Create a working container from the image
echo "Creating container from image $IMAGE_ID..."
CONTAINER_ID=$(sudo buildah from --storage-driver vfs --root /var/lib/containers/storage "$IMAGE_ID")
if [ $? -ne 0 ]; then
    echo "Error: Failed to create container. Exiting."
    exit 1
fi

echo "Container created with ID: $CONTAINER_ID"

# 2. Mount the container's file system
echo "Mounting container's file system..."
MOUNT_POINT=$(sudo buildah mount --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID")
if [ $? -ne 0 ]; then
    echo "Error: Failed to mount container. Exiting."
    # Clean up the container before exiting
    sudo buildah rm --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"
    exit 1
fi

echo "Container mounted at: $MOUNT_POINT"

# 3. Find the file in the mounted filesystem and copy it
echo "Searching for $FILE_NAME in container filesystem..."
SOURCE_PATH=$(sudo find "$MOUNT_POINT" -name "$FILE_NAME" -print -quit)

if [ -z "$SOURCE_PATH" ]; then
    echo "Error: File not found in container: $FILE_NAME"
    # Clean up the container before exiting
    sudo buildah unmount --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"
    sudo buildah rm --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"
    exit 1
fi

DESTINATION_PATH="$OUTPUT_DIR/$FILE_NAME"
echo "File found at: $SOURCE_PATH"
echo "Copying file from container to host..."
sudo cp "$SOURCE_PATH" "$DESTINATION_PATH"
if [ $? -ne 0 ]; then
    echo "Error: Failed to copy file. Exiting."
    # Clean up the container before exiting
    sudo buildah unmount --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"
    sudo buildah rm --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"
    exit 1
fi

echo "File successfully copied to $DESTINATION_PATH"

# 4. Unmount and remove the container
echo "Cleaning up container..."
sudo buildah unmount --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"
sudo buildah rm --storage-driver vfs --root /var/lib/containers/storage "$CONTAINER_ID"

echo "Done."
