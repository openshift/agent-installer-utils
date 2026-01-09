#!/bin/bash
#
# Skill: build-iso-builder
# Description: Build and push agent-iso-builder container image
# Usage: /build-iso-builder --tag <image:tag> [--authfile <path>] [--push]
#

set -euo pipefail

# Variables
IMAGE_TAG=""
AUTHFILE=""
AUTO_PUSH=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --tag)
      IMAGE_TAG="$2"
      shift 2
      ;;
    --authfile)
      AUTHFILE="$2"
      shift 2
      ;;
    --push)
      AUTO_PUSH=true
      shift
      ;;
    --help|-h)
      echo "Usage: /build-iso-builder --tag <image:tag> [--authfile <path>] [--push]"
      echo ""
      echo "Required Options:"
      echo "  --tag <image:tag>    Container image tag"
      echo ""
      echo "Optional Options:"
      echo "  --authfile <path>    Path to JSON authentication file for registry push"
      echo "  --push               Automatically push to registry without prompting"
      echo ""
      echo "Examples:"
      echo "  /build-iso-builder --tag quay.io/myuser/agent-iso-builder:latest"
      echo "  /build-iso-builder --tag quay.io/myuser/agent-iso-builder:latest --authfile ~/.docker/config.json"
      echo "  /build-iso-builder --tag quay.io/myuser/agent-iso-builder:latest --authfile ~/.docker/config.json --push"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Validate required parameters
if [[ -z "${IMAGE_TAG}" ]]; then
  echo "Error: --tag parameter is required"
  echo ""
  echo "Usage: /build-iso-builder --tag <image:tag> [--authfile <path>] [--push]"
  echo "Use --help for more information"
  exit 1
fi

# Validate authfile exists if provided
if [[ -n "${AUTHFILE}" && ! -f "${AUTHFILE}" ]]; then
  echo "Error: Authentication file not found: ${AUTHFILE}"
  exit 1
fi

# Get repository root (go up 3 levels from .claude/skills/build-iso-builder/build.sh)
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "${REPO_ROOT}"

echo "============================================"
echo "Building agent-iso-builder container"
echo "============================================"
echo "Repository: ${REPO_ROOT}"
echo "Image tag:  ${IMAGE_TAG}"
echo ""

# Create temporary build directory
BUILD_DIR=$(mktemp -d -t agent-iso-builder-build-XXXXXX)
trap "rm -rf ${BUILD_DIR}" EXIT

echo "Creating temporary build directory: ${BUILD_DIR}"

# Copy iso_builder tools to build context
echo "Copying tools/iso_builder to build context..."
cp -r tools/iso_builder "${BUILD_DIR}/src"

# Dynamically generate Dockerfile
echo "Generating Dockerfile..."
cat > "${BUILD_DIR}/Dockerfile" <<'EOF'
# Dynamically generated Dockerfile for agent-iso-builder
# This packages the iso_builder tools for use by dev-scripts
# Note: Uses scratch base since dev-scripts only extracts /src, doesn't run commands

FROM scratch
COPY src /src
EOF

echo ""
echo "Generated Dockerfile:"
echo "--------------------"
cat "${BUILD_DIR}/Dockerfile"
echo "--------------------"
echo ""

# Build container image
echo "Building container image: ${IMAGE_TAG}"
podman build -t "${IMAGE_TAG}" "${BUILD_DIR}"

echo ""
echo "Build successful!"
echo ""

# Determine if we should push
SHOULD_PUSH=false
if [[ "${AUTO_PUSH}" == "true" ]]; then
  SHOULD_PUSH=true
else
  # Ask before pushing
  read -p "Push image to registry? (y/N) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    SHOULD_PUSH=true
  fi
fi

# Push if requested
if [[ "${SHOULD_PUSH}" == "true" ]]; then
  echo "Pushing ${IMAGE_TAG} to registry..."

  # Push with optional authfile
  if [[ -n "${AUTHFILE}" ]]; then
    podman push --authfile "${AUTHFILE}" "${IMAGE_TAG}"
  else
    podman push "${IMAGE_TAG}"
  fi

  echo ""
  echo "============================================"
  echo "Successfully pushed ${IMAGE_TAG}"
  echo "============================================"
else
  echo "Skipping push. Image is available locally."
  echo ""
  echo "To push manually, run:"
  if [[ -n "${AUTHFILE}" ]]; then
    echo "  podman push --authfile \"${AUTHFILE}\" \"${IMAGE_TAG}\""
  else
    echo "  podman push \"${IMAGE_TAG}\""
  fi
fi

echo ""
echo "To use this image in dev-scripts, set:"
echo "  export AGENT_ISO_BUILDER_IMAGE=\"${IMAGE_TAG}\""
