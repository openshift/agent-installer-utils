# ABI OVE Image Builder
This script generates 'agent-ove-\<arch\>.iso'

## Custom OpenShift Installer Script

This script allows you to test a locally built OpenShift installer by patching it with a release version.

### Setup

1. **Build the Installer**: Checkout and build the OpenShift installer binary.
2. **Set Custom Installer Path**: Set the `CUSTOM_OPENSHIFT_INSTALLER_PATH` environment variable to point to your custom installer binary:
   ```bash
    CUSTOM_OPENSHIFT_INSTALLER_PATH=/path/to/installerCode ./tools/iso_builder/hack/build-ove-image.sh 
    ```