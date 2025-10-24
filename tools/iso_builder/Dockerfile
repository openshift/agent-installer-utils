ARG MAJOR_MINOR_VERSION
FROM registry.ci.openshift.org/ocp/${MAJOR_MINOR_VERSION}:agent-preinstall-image-builder AS builder

# Define the arguments that will be passed in from the pipeline
ARG RELEASE_FLAG
ARG RELEASE_VALUE
ARG MAJOR_MINOR_VERSION
ARG PATCH_VERSION
ARG ARCH

# Copy the configuration script into the container
COPY hack/build-ove-image.sh hack/helper.sh hack/logging.sh /tmp/
RUN chmod +x /tmp/build-ove-image.sh
RUN chmod +x /tmp/helper.sh
RUN chmod +x /tmp/logging.sh

# COPY appliance-config.yaml and operator CRs into container
COPY config/${MAJOR_MINOR_VERSION}/. /

# Run the configuration script to finish setting up appliance-config.yaml
RUN --mount=type=secret,id=ove-ui-image-pull-secret/.dockerconfigjson \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/ove-ui-image-pull-secret/.dockerconfigjson "${RELEASE_FLAG}" "${RELEASE_VALUE}" \
    --dir / --step "configure"

# Run the openshift-appliance build live-iso command during the image build process.
RUN /openshift-appliance build live-iso --log-level debug && rm -r /temp && rm -r /cache && rm -r /root/.oc-mirror

# Do the postprocessing step to generate the OVE ISO
RUN --mount=type=secret,id=ove-ui-image-pull-secret/.dockerconfigjson \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/ove-ui-image-pull-secret/.dockerconfigjson ${RELEASE_FLAG} ${RELEASE_VALUE} --dir / --step "create-iso" \
    && rm /appliance.iso && rm -rf /isomnt

FROM scratch

ARG ARCH
ARG MAJOR_MINOR_VERSION
ARG PATCH_VERSION
# Copy final ISO as only thing from builder stage
COPY --from=builder /agent-ove.$ARCH.iso /agent-ove.$ARCH.iso

LABEL summary="OpenShift Agent-Based Installer for OVE UI ISO" \
      name="Agent Installer for OpenShift-Virt" \
      description="An ISO that provides a UI-driven installation for Openshift Virtualization using the Openshift Appliance." \
      io.k8s.description="An ISO that provides a UI-driven installation for Openshift Virtualization using the Openshift Appliance." \
      io.k8s.display-name="Agent Installer for OpenShift-Virt" \
      io.openshift.tags="openshift,installer,agent" \
      com.redhat.component="Installer/Agent based installation" \
      distribution-scope="public" \
      url="https://github.com/openshift/agent-installer-utils" \
      vendor="Red Hat, Inc." \
      version="${MAJOR_MINOR_VERSION}" \
      release="${PATCH_VERSION}"
