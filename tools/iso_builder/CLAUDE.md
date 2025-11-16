# CLAUDE.md - ISO Builder

## Overview

ISO Builder creates bootable ISO images for agent-based OpenShift installations in disconnected environments. The generated ISO includes a given OCP release images with a subset of OLM operators, and some additional support images.

## Prerequisites

Install required tools (requires unsandboxed Bash):
```bash
sudo dnf install -y xorriso syslinux skopeo

# Also need: oc, podman
```

## Build Commands

```bash
cd tools/iso_builder

# Using OCP version (recommended for GA versions 4.19+)
make build-ove-iso PULL_SECRET_FILE=~/pull_secret.json RELEASE_IMAGE_VERSION=4.19.1

# Using release image URL (recommended for CI/nightly builds)
make build-ove-iso PULL_SECRET_FILE=~/pull_secret.json RELEASE_IMAGE_URL=registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-04-01-173804

# Optional parameters
make build-ove-iso PULL_SECRET_FILE=~/pull_secret.json RELEASE_IMAGE_VERSION=4.19.1 ARCH=x86_64

# Direct script usage with all options
./hack/build-ove-image.sh \
  --pull-secret-file ~/pull_secret.json \
  --ocp-version 4.19.1 \
  --arch x86_64 \
  --ssh-key-file ~/.ssh/id_rsa \
  --dir ~/iso_builder

# Container-based build
make build-ove-iso-container PULL_SECRET_FILE=~/pull_secret.json RELEASE_IMAGE_VERSION=4.19.1
```

## Output

Default output location: `/tmp/iso_builder/<OCP_VERSION>/ove/output/agent-ove.<arch>.iso`

Directory structure after build:
```
/tmp/iso_builder/
└── 4.19.0-0.ci-2025-04-01-173804
    ├── agent-artifacts
    ├── appliance
    ├── ignition
    ├── logs
    └── ove
        ├── output
        │   └── agent-ove.x86_64.iso
        └── work
```

## Cleanup

```bash
# Remove all temporary files and generated ISO
make cleanall

# Remove temp dirs for specific OCP version, preserve image cache
make clean-appliance-temp-dir
```

## Build Process

### Main Script: `hack/build-ove-image.sh`

The build follows these steps:

1. **Input Validation** (via `hack/helper.sh`)
   - Exactly one of `--ocp-version` or `--release-image-url` required
   - `--pull-secret-file` is required
   - OCP version must be in format `x.y.z`
   - Architecture defaults to `x86_64` if not specified
   - Directory defaults to `/tmp/iso_builder` if not specified

2. **Variable Setup** (`setup_vars()` in `hack/helper.sh`)
   - Extracts full OCP version from image metadata using `skopeo inspect`
   - Parses `major.minor.patch` version from full version string
   - Extracts `major.minor` version for image selection
   - Creates directory structure: `<DIR>/<full_version>/{ove,appliance,logs}`

3. **Appliance Configuration** (`create_appliance_config()`)
   - Copies version-specific config from `config/<major.minor>/` to appliance work dir
   - Creates `appliance-config.yaml` with:
     - OCP release version/URL and architecture
     - Pull secret (required)
     - SSH key (optional)

4. **ISO Build** (`build_live_iso()`)
   - Uses appliance image: `registry.ci.openshift.org/ocp/<major.minor>:agent-preinstall-image-builder`
   - Runs podman with privileged mode and host networking
   - Mounts appliance work dir to `/assets` in container
   - Executes: `build live-iso --log-level debug`
   - Outputs `appliance.iso` to work dir

5. **Finalization** (`finalize()`)
   - Copies `appliance.iso` to `agent-ove.<arch>.iso` in output directory
   - Reports total execution time

### Helper Script: `hack/helper.sh`

Key functions:
- `parse_inputs()` - Parses command-line arguments with validation
- `validate_inputs()` - Ensures required parameters present and valid
- `setup_vars()` - Establishes all directory paths and version variables
- `usage()` - Prints detailed help text

### Build Steps Parameter

The `--step` parameter controls execution phases:
- `all` (default) - Run complete build process
- `configure` - Only create appliance configuration
- `create-iso` - Only finalize ISO (assumes appliance.iso exists)

This is used internally by containerized builds where configuration and ISO creation happen in different stages.

## Makefile

Variables:
- `ARCH` - Target architecture (default: x86_64)
- `PULL_SECRET_FILE` - Path to pull secret (default: ./pull-secret.json)
- `RELEASE_IMAGE_URL` - Release image URL (optional)
- `RELEASE_IMAGE_VERSION` - OCP version (optional)

The Makefile automatically:
- Determines which flag (`--ocp-version` or `--release-image-url`) to use based on variables set
- Extracts full OCP version using skopeo for container builds
- Parses major.minor and patch versions for build arguments

## Container-Based Build

`make build-ove-iso-container` builds using buildah with:
- Specific capabilities: CAP_SYS_ADMIN, CAP_NET_ADMIN
- VFS storage driver
- container_runtime_t SELinux context
- Pull secret passed as build secret

The Dockerfile uses multi-stage approach with `--step` parameter to separate configuration from ISO creation.

## Logging

Build logs stored in: `<DIR>/<OCP_VERSION>/logs/`

Use `hack/logging.sh` for consistent log formatting across scripts.

## Version-Specific Configuration

Configuration files in `config/<major.minor>/` directory contain version-specific settings for the appliance. These are copied to the appliance work directory before building.

## Common Issues

- Ensure pull secret is valid for the registry hosting the release image
- Container builds require sudo/root for buildah with required capabilities
- Image cache is preserved in appliance directory across builds to speed up rebuilds
- For CI/nightly builds, use `--release-image-url` to specify exact image reference
- For GA releases, use `--ocp-version` which fetches from stable channels
