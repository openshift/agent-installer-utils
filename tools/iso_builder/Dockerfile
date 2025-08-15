# Use appliance image as the base
FROM registry.ci.openshift.org/ocp/4.20:agent-preinstall-image-builder

USER root

# Copy the configuration script into the container
COPY hack/build-appliance-config.sh hack/helper.sh hack/logging.sh hack/create-ove-iso.sh /tmp/

RUN chmod +x /tmp/build-appliance-config.sh
RUN chmod +x /tmp/create-ove-iso.sh
RUN chmod +x /tmp/helper.sh
RUN chmod +x /tmp/logging.sh

# RUN mkdir -p /assets

# Run the configuration script to set up the appliance config
RUN --mount=type=secret,id=my_pull_secret,target=/run/secrets/pull-secret.txt \
    /tmp/build-appliance-config.sh --pull-secret-file /run/secrets/pull-secret.txt "$RELEASE_FLAG" "$RELEASE_VALUE"

# RUN dnf -y update; yum -y reinstall shadow-utils; \
# yum -y install podman runc fuse-overlayfs --exclude container-selinux; \
RUN dnf -y update; yum -y reinstall shadow-utils; \
yum -y install podman runc ; \
rm -rf /var/cache /var/log/dnf* /var/log/yum.*

# Set up volumes for rootfull or rootless podman
# VOLUME /var/lib/containers

# chmod containers.conf and adjust storage.conf to enable Fuse storage.
# RUN chmod 644 /etc/containers/containers.conf; sed -i -e 's|^#mount_program|mount_program|g' -e '/additionalimage.*/a "/var/lib/shared",' -e 's|^mountopt[[:space:]]*=.*$|mountopt = "nodev,fsync=0"|g' /etc/containers/storage.conf
# RUN mkdir -p /var/lib/shared/overlay-images /var/lib/shared/overlay-layers /var/lib/shared/vfs-images /var/lib/shared/vfs-layers; touch /var/lib/shared/overlay-images/images.lock; touch /var/lib/shared/overlay-layers/layers.lock; touch /var/lib/shared/vfs-images/images.lock; touch /var/lib/shared/vfs-layers/layers.lock

# Use runc as containers engine
RUN mkdir -p /etc/containers/ && echo "[containers]" >> /etc/containers/containers.conf && echo "label=false" >> /etc/containers/containers.conf && echo "[engine]" >> /etc/containers/containers.conf && echo "runtime=\"runc\"" >> /etc/containers/containers.conf

# Run the openshift-appliance build live-iso command during the image build process.
RUN /openshift-appliance build live-iso --log-level debug && rm -r /temp && rm -r /cache && rm -r /root/.oc-mirror

# Do the postprocessing step to generate the OVE ISO
RUN dnf install -y xorriso rsync

RUN --mount=type=secret,id=my_pull_secret,target=/run/secrets/pull-secret.txt \
    /tmp/create-ove-iso.sh --pull-secret-file /run/secrets/pull-secret.txt $RELEASE_FLAG $RELEASE_VALUE --dir / && rm /appliance.iso

# Move the OVE ISO where it can be retrieved
RUN mkdir -p /release && mv /agent-ove.x86_64.iso /release

# Override openshift-appliance
ENTRYPOINT ["/bin/bash"]

CMD ["/bin/bash", "-c", "sleep infinity"]
