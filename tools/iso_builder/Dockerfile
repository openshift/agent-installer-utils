# Use the appliance image as the base
FROM registry.ci.openshift.org/ocp/4.20:agent-preinstall-image-builder

RUN yum install -y rsync

# Set the ENTRYPOINT to the openshift-appliance executable.
# This makes it so that 'podman run' commands can use sub-commands directly.
ENTRYPOINT ["openshift-appliance"]

# The CMD defines the default command to run when the container starts.
# In this case, it's the 'build live-iso' command.
CMD ["build", "live-iso", "--log-level", "debug", "--dir", "/tmp/iso_builder/4.20.0/appliance"]
