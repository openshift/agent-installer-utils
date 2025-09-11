FROM registry.ci.openshift.org/ocp/4.20:agent-preinstall-image-builder AS builder

# Define the arguments that will be passed in from the pipeline
ARG RELEASE_FLAG
ARG RELEASE_VALUE
ARG MAJOR_MINOR_VERSION
ARG ARCH
# Set the environment variables using the values from the arguments
ENV RELEASE_FLAG=${RELEASE_FLAG}
ENV RELEASE_VALUE=${RELEASE_VALUE}
ENV MAJOR_MINOR_VERSION=${MAJOR_MINOR_VERSION}
ENV ARCH=${ARCH}

USER root
# Copy the configuration script into the container
COPY hack/build-ove-image.sh hack/helper.sh hack/logging.sh /tmp/

RUN chmod +x /tmp/build-ove-image.sh
RUN chmod +x /tmp/helper.sh
RUN chmod +x /tmp/logging.sh

# COPY appliance-config.yaml and operator CRs into container
COPY ./config/${MAJOR_MINOR_VERSION}/. /

# Run the configuration script to finish setting up appliance-config.yaml
RUN --mount=type=secret,id=ove-ui-image-pull-secret/.dockerconfigjson \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/ove-ui-image-pull-secret/.dockerconfigjson "$RELEASE_FLAG" "$RELEASE_VALUE" \
    --dir / --step "configure"

RUN dnf -y update; yum -y reinstall shadow-utils; \
yum -y install podman runc ; \
rm -rf /var/cache /var/log/dnf* /var/log/yum.*

# Use runc as containers engine
RUN mkdir -p /etc/containers/ && \
  echo "[engine]" >> /etc/containers/containers.conf && \
  echo "runtime=\"runc\"" >> /etc/containers/containers.conf

# Run the openshift-appliance build live-iso command during the image build process.
RUN /openshift-appliance build live-iso --log-level debug && rm -r /temp && rm -r /cache && rm -r /root/.oc-mirror

# Do the postprocessing step to generate the OVE ISO
RUN dnf install -y xorriso rsync
RUN --mount=type=secret,id=ove-ui-image-pull-secret/.dockerconfigjson \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/ove-ui-image-pull-secret/.dockerconfigjson $RELEASE_FLAG $RELEASE_VALUE --dir / --step "create-iso" \
    && rm /appliance.iso && rm -rf /isomnt

FROM scratch

ARG ARCH
ENV ARCH=${ARCH}

# Copy final ISO as only thing from builder stage
COPY --from=builder /agent-ove.$ARCH.iso /agent-ove.$ARCH.iso
