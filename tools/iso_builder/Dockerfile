# Use appliance image as the base
FROM registry.ci.openshift.org/ocp/4.20:agent-preinstall-image-builder AS builder

USER root

# Copy the configuration script into the container
COPY hack/build-ove-image.sh hack/helper.sh hack/logging.sh /tmp/

RUN chmod +x /tmp/build-ove-image.sh
RUN chmod +x /tmp/helper.sh
RUN chmod +x /tmp/logging.sh

# COPY appliance-config.yaml and operator CRs into container
COPY ./config/${MAJOR_MINOR_VERSION}/. /

# Run the configuration script to finish setting up appliance-config.yaml
RUN --mount=type=secret,id=image_pull_secrets,target=/run/secrets/pull-secret.txt \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/pull-secret.txt "$RELEASE_FLAG" "$RELEASE_VALUE" \
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
RUN --mount=type=secret,id=image_pull_secrets,target=/run/secrets/pull-secret.txt \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/pull-secret.txt $RELEASE_FLAG $RELEASE_VALUE --dir / --step "create-iso" \
    && rm /appliance.iso && rm -rf /isomnt

FROM scratch

# Copy final ISO as only thing from builder stage
COPY --from=builder /agent-ove.$ARCH.iso /agent-ove.$ARCH.iso
