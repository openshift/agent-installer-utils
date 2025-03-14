.PHONY: cleanall clean clean-appliance-temp clean-agent-artifacts delete-ove-work-dir delete-ove-iso build-ove-iso
ARCH ?= x86_64

# Function to check if OCP_RELEASE_IMAGE and PULL_SECRET are set for build-ove-iso, or just OCP_RELEASE_IMAGE for other targets
define check_ocp_release_image
	@if [ "$(firstword $(MAKECMDGOALS))" = "build-ove-iso" ]; then \
		if [ -z "${OCP_RELEASE_IMAGE}" ] || [ -z "${PULL_SECRET}" ]; then \
			echo "Error: Both OCP_RELEASE_IMAGE and PULL_SECRET must be set!"; \
			echo "Usage: make build-ove-iso OCP_RELEASE_IMAGE=<RELEASE_IMAGE> PULL_SECRET=</path/to/pull-secret.json>"; \
			exit 1; \
		fi; \
	fi; \
	if [ -z "$(OCP_RELEASE_IMAGE)" ]; then \
		echo "OCP_RELEASE_IMAGE is not set! Please pass it like: make $(strip $(1)) OCP_RELEASE_IMAGE=registry.ci.openshift.org/ocp/release:4.18.0"; \
		exit 1; \
	fi
endef


# Helper to clean directories based on the OCP_RELEASE_IMAGE
define clean_dir
	@$(call check_ocp_release_image, $1)

	@OCP_VERSION=$(shell echo $(OCP_RELEASE_IMAGE) | awk -F ':' '{print $$2}' | awk -F '-' '{print $$1}'); \
	if [ -z "$$OCP_VERSION" ]; then \
		echo "Failed to determine OCP version from OCP_RELEASE_IMAGE"; \
		exit 1; \
	fi; \
	if [ "$(firstword $(MAKECMDGOALS))" = "cleanall" ]; then \
		sudo umount /tmp/iso_builder/$$OCP_VERSION/appliance/mnt; \
		sudo rm -rf /tmp/iso_builder/$$OCP_VERSION; \
		echo "Cleaned /tmp/iso_builder/$$OCP_VERSION directory."; \
		exit 0; \
	fi; \
	dir="/tmp/iso_builder/$$OCP_VERSION/$(strip $(2))"; \
	if [ ! -d "$$dir" ]; then \
		echo "Directory $$dir does not exist!"; \
		exit 1; \
	fi; \
	$(3); \
	echo "Cleaned $$dir directory."
endef

cleanall:
	$(call clean_dir, cleanall, /tmp/iso_builder, "")

clean: 
	@$(call check_ocp_release_image, clean)
	echo "Cleaning appliance temp files"; \
	$(MAKE) clean-appliance-temp ;\
	echo "Cleaning agent tui artifacts directory"; \
	$(MAKE) clean-agent-artifacts; \

clean-appliance-temp:
	$(call clean_dir, clean-appliance-temp, appliance/work, sudo podman run --rm -it --privileged --net=host -v "$$dir/:/assets:Z" quay.io/edge-infrastructure/openshift-appliance:latest clean)

clean-agent-artifacts:
	$(call clean_dir, clean-agent-artifacts, agent-artifacts, rm -rf $$dir)

delete-ove-work-dir:
	$(call clean_dir, delete-ove-work-dir, ove/work, rm -rf $$dir)

delete-ove-iso:
	$(call clean_dir, delete-ove-iso, ove/output, rm -rf $$dir)

build-ove-iso:
	@$(call check_ocp_release_image, build-ove-iso)
	hack/build-ove-image.sh --release-image ${OCP_RELEASE_IMAGE} --arch ${ARCH} --pull-secret ${PULL_SECRET}
