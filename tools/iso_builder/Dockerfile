# Use appliance image as the base
FROM registry.ci.openshift.org/ocp/4.20:agent-preinstall-image-builder
# TEST ONLY of podman run command
# FROM quay.io/bfournie/openshift-appliance:latest

USER root

# TEST of running podman to see if cgroup write failure occurs
# RUN podman run --privileged --cgroupns=host --network=none --rm alpine sh -c "mkdir -p /sys/fs/cgroup/test && echo 'hello' > /sys/fs/cgroup/test/file"
RUN podman run --network=none --rm alpine:latest ls

# Define the arguments that will be passed in from the pipeline
ARG RELEASE_FLAG
ARG RELEASE_VALUE
ARG MAJOR_MINOR_VERSION
# Set the environment variables using the values from the arguments
ENV RELEASE_FLAG=${RELEASE_FLAG}
ENV RELEASE_VALUE=${RELEASE_VALUE}
ENV MAJOR_MINOR_VERSION=${MAJOR_MINOR_VERSION}

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
# RUN mkdir -p /etc/containers/ && \
#  echo "[engine]" >> /etc/containers/containers.conf && \
#  echo "runtime=\"runc\"" >> /etc/containers/containers.conf

# Run the openshift-appliance build live-iso command during the image build process.
RUN /openshift-appliance build live-iso --log-level debug && rm -r /temp && rm -r /cache && rm -r /root/.oc-mirror

# Do the postprocessing step to generate the OVE ISO
RUN dnf install -y xorriso rsync

RUN --mount=type=secret,id=ove-ui-image-pull-secret/.dockerconfigjson \
    /tmp/build-ove-image.sh --pull-secret-file /run/secrets/ove-ui-image-pull-secret/.dockerconfigjson $RELEASE_FLAG $RELEASE_VALUE --dir / --step "create-iso" \
    && rm /appliance.iso && rm -rf /isomnt

# Override openshift-appliance entrypoint for testing purposes
ENTRYPOINT ["/bin/bash"]

CMD ["/bin/bash", "-c", "sleep infinity"]
