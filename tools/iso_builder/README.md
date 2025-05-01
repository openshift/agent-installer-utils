# ISO Builder

The ISOBuilder tool creates an ISO using the appliance and interactive unconfigured ignition commands from the agent-based installer. This ISO is designed for disconnected environments, enabling an interactive installation workflow.

The generated ISO will include essential components such as:

- Registry images
- Agent TUI 
- The new web UI
- Embedded agent-based installer capabilities

This ensures a seamless installation experience with all necessary resources pre-packaged.

The generated ISO is saved at `/tmp/iso_builder/<OCP_VERSION>/ove/output/agent-ove.<arch>.iso`

By default:

- The architecture is `x86_64`

- The default output directory is `/tmp/iso_builder`

## Pre-requisites

1. xorriso (see [xorriso](https://www.gnu.org/software/xorriso/) for more information). On CentOS, RHEL, run `sudo dnf install -y xorriso`
2. coreos-installer (see [coreos-installer](https://coreos.github.io/coreos-installer/getting-started/) for more information). On CentOS, RHEL, run `sudo dnf install -y coreos-installer`
3. isohybrid (see [isohybrid](https://man.archlinux.org/man/core/syslinux/isohybrid.1.en) for more information). On CentOS, RHEL, run `sudo dnf install -y syslinux`
4. oc (see [oc](https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/cli_tools/openshift-cli-oc#cli-installing-cli_cli-developer-commands) for more information).
5. podman (see [podman](https://podman.io/docs/installation) for more information).
6. skopeo (see [skopeo](https://github.com/containers/skopeo/blob/main/install.md) for more information). On RHEL / CentOS Stream ≥ 8, run `sudo dnf -y install skopeo`


## Quick Start

Run as simple bash script. Uses default arch and output dir location.
```bash
cd tools/iso_builder
make build-ove-iso
```
Alternatively,

```bash
cd tools/iso_builder
./hack/build-ove-image.sh --pull-secret-file <pull-secret> --ocp-version <openshift-major.minor.patch-version>
```
## Examples:
- Specify release-image-url, useful for dev/test purposes:

```bash
cd tools/iso_builder
make build-ove-iso
./hack/build-ove-image.sh --pull-secret-file <pull-secret> --release-image-url <openshift-release-image-url>
```

- Specify a custom output dir:
```bash
cd tools/iso_builder
./hack/build-ove-image.sh --pull-secret-file <pull-secret> --ocp-version <openshift-major.minor.patch-version> --dir <path>
```

- Specify a ssh file:
```bash
cd tools/iso_builder
./hack/build-ove-image.sh --pull-secret-file <pull-secret> --ocp-version <openshift-major.minor.patch-version> --ssh-key-file <path>
```

## Outputs:
`agent-ove.x86_64.iso`: Bootable agent OVE ISO image.
Directory Structure After Running the Script:
```bash
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
## Cleanup:

To clean default ISOBuilder tmp directory ( /tmp/iso_builder) and all the intermediate files along with the generated ISO 
```bash
cd tools/iso_builder
make cleanall
```

This cleanup target is intended for development use—it removes only temporary working directories for a specific OCP_VERSION, preserving reusable artifacts like the image cache to speed up future rebuilds. It's a safer, more targeted alternative to cleanall, helping you save disk space without losing valuable cached content.
```bash
cd tools/iso_builder
make clean-appliance-temp-dir
```

## Using a Custom OpenShift Installer:
You can use a locally built OpenShift installer with the build-ove-image.sh script by following these steps:

1. Build the Custom Installer
Clone and build the OpenShift Installer from source:

```
git clone https://github.com/openshift/installer.git
cd installer
./hack/build.sh
```
This will produce the openshift-install binary in the bin/ directory.

2. Set the Installer Path and Run the Script
Set the CUSTOM_OPENSHIFT_INSTALLER_PATH environment variable to the directory containing your custom-built openshift-install binary, then run the build script:

```
CUSTOM_OPENSHIFT_INSTALLER_PATH=~/installer \
./hack/build-ove-image.sh \
  --release-image-url registry.ci.openshift.org/ocp/release:4.19.0-0.ci-2025-05-01-113742 \
  --pull-secret-file ~/pull_secret.json
```