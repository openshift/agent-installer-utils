.PHONY: cleanall clean-appliance-temp-dir build-ove-iso build-ove-iso-container

ARCH ?= x86_64
PULL_SECRET_FILE ?= ./pull-secret.json
RELEASE_IMAGE_URL ?=
RELEASE_IMAGE_VERSION ?=

ifdef RELEASE_IMAGE_VERSION
    RELEASE_FLAG := --ocp-version
    RELEASE_VALUE := $(RELEASE_IMAGE_VERSION)
    full_ocp_version :=${RELEASE_IMAGE_VERSION}
else ifdef RELEASE_IMAGE_URL
    RELEASE_FLAG := --release-image-url
    RELEASE_VALUE := $(RELEASE_IMAGE_URL)
    full_ocp_version := $(shell skopeo inspect --authfile $(PULL_SECRET_FILE) docker://$(RELEASE_IMAGE_URL) | jq -r '.Labels["io.openshift.release"]')
endif

MAJOR_MINOR_VERSION := $(shell echo $(full_ocp_version) | cut -d'.' -f1,2)
PATCH_VERSION := $(shell echo $(full_ocp_version) | cut -d: -f2 | cut -d'.' -f3 | cut -d'-' -f1)

cleanall:
	hack/cleanup.sh cleanall

clean-appliance-temp-dir:
	hack/cleanup.sh clean-appliance-temp-dir
	
build-ove-iso:
	hack/build-ove-image.sh $(RELEASE_FLAG) $(RELEASE_VALUE) --pull-secret-file $(PULL_SECRET_FILE)

build-ove-iso-container:
	# Build the container with specific capabilities to support skopeo used by openshift-appliance
	@sudo buildah bud \
		--cap-add=all \
		--build-arg RELEASE_FLAG="$(RELEASE_FLAG)" \
		--build-arg RELEASE_VALUE="$(RELEASE_VALUE)" \
		--build-arg MAJOR_MINOR_VERSION="$(MAJOR_MINOR_VERSION)" \
		--build-arg PATCH_VERSION="$(PATCH_VERSION)" \
		--build-arg ARCH="$(ARCH)" \
		--authfile $(PULL_SECRET_FILE) \
		--secret id=ove-ui-image-pull-secret/.dockerconfigjson,src=$(PULL_SECRET_FILE) \
		--storage-driver vfs \
		-t appliance-test \
		-f Dockerfile
