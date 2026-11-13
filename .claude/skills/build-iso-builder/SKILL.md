---
name: build-iso-builder
description: Build and push the agent-iso-builder container image for use with dev-scripts. Use when the user wants to build or update the iso-builder container image.
allowed-tools: Bash, Read
---

# build-iso-builder

Builds and optionally pushes the agent-iso-builder container image.

## Usage

```
/build-iso-builder --tag <image:tag> [--authfile <path>] [--push]
```

## Required Parameters

- `--tag <image:tag>` - Container image tag (e.g., `quay.io/myuser/agent-iso-builder:latest`)

## Optional Parameters

- `--authfile <path>` - Path to JSON authentication file for registry push (e.g., `~/.docker/config.json`)
- `--push` - Automatically push to registry without prompting

## What it does

1. Creates a temporary build directory
2. Copies `tools/iso_builder` to the build context
3. Generates a Dockerfile using a scratch base image
4. Builds the container image with podman
5. Optionally pushes the image to the registry (with user confirmation)

## Implementation

Execute the bundled `build.sh` script with the user's parameters:

```bash
bash .claude/skills/build-iso-builder/build.sh [user's arguments]
```

The script handles all aspects of the build process including validation, container building, and optional pushing.

## Examples

```
/build-iso-builder --tag quay.io/myuser/agent-iso-builder:latest
/build-iso-builder --tag quay.io/myuser/agent-iso-builder:latest --authfile ~/.docker/config.json
/build-iso-builder --tag quay.io/myuser/agent-iso-builder:latest --authfile ~/.docker/config.json --push
```

## Output

The script will:
- Build the container image locally
- Ask for confirmation before pushing to the registry
- Display instructions for using the image in dev-scripts:
  ```
  export AGENT_ISO_BUILDER_IMAGE="<your-tag>"
  ```
