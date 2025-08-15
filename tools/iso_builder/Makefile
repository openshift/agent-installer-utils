.PHONY: cleanall clean-appliance-temp-dir build-ove-iso build-ove-iso-container

ARCH ?= x86_64
PULL_SECRET_FILE ?= ./pull-secret.json
RELEASE_IMAGE_URL ?=
RELEASE_IMAGE_VERSION ?=

ifdef RELEASE_IMAGE_VERSION
    RELEASE_FLAG := --ocp-version
    RELEASE_VALUE := $(RELEASE_IMAGE_VERSION)
else ifdef RELEASE_IMAGE_URL
    RELEASE_FLAG := --release-image-url
    RELEASE_VALUE := $(RELEASE_IMAGE_URL)
endif

cleanall:
	hack/cleanup.sh cleanall

clean-appliance-temp-dir:
	hack/cleanup.sh clean-appliance-temp-dir
	
build-ove-iso:
	hack/build-ove-image.sh --release-image-url ${RELEASE_IMAGE_URL} --pull-secret-file ${PULL_SECRET_FILE}

build-ove-iso-container:
	sudo buildah bud \
		--env RELEASE_FLAG="$(RELEASE_FLAG)" \
		--env RELEASE_VALUE="$(RELEASE_VALUE)" \
		--authfile ${PULL_SECRET_FILE} \
		--secret id=my_pull_secret,src=$(PULL_SECRET_FILE) \
		--cap-add CAP_SYS_ADMIN \
		--storage-driver vfs \
		-t appliance-test \
		-f Dockerfile
