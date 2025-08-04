FROM registry.access.redhat.com/ubi9/ubi:9.6-1753978585

USER root

ARG PULL_SECRET_CONTAINER_PATH="/run/secrets/my_pull_secret_file" # Good default for secrets

# RUN dnf install -y --enablerepo=ubi-9-baseos-rpms gcc xorriso syslinux coreos-installer && dnf clean all

# Set the working directory inside the container
WORKDIR /app

COPY hack/ hack/
COPY data/ data/

# RUN dnf install -y --enablerepo=ubi-9-baseos-rpms gcc jq podman xorriso syslinux coreos-installer && dnf clean all
RUN dnf install -y gcc jq podman && dnf clean all

RUN chmod +x hack/build-ove-image.sh hack/helper.sh

# Create the final output directory /app/output and ensure it's writable
RUN mkdir -p /app/output && \
    chmod g+rwx /app/output && \
    chgrp 0 /app/output

# Create the temporary ISO build directory as expected by the script
RUN mkdir -p /tmp/iso_builder && \
    chmod g+rwx /tmp/iso_builder && \
    chgrp 0 /tmp/iso_builder

# Define the volume for persistent storage of the generated ISO
VOLUME /app/output

RUN --mount=type=secret,id=my_pull_secret,target=${PULL_SECRET_CONTAINER_PATH} \
    hack/build-ove-image.sh --pull-secret-file "${PULL_SECRET_CONTAINER_PATH}" --ocp-version 4.20.0

