ARCH ?= x86_64
IMAGE := $(or ${IMAGE}, localhost/isobuilder:latest)
PWD = $(shell pwd)
CMD := $(or ${CMD}, build-isobuilder)
OUTPUT := $(if ${OUTPUT},${OUTPUT})

# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
# From: https://stackoverflow.com/questions/10858261/how-to-abort-makefile-if-variable-not-set
check_defined = \
	$(strip $(foreach 1,$1, \
		$(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
	$(if $(value $1),, \
		$(error Undefined environment variable $1$(if $2, ($2))))

.PHONY:clean
clean:
	rm -rf /tmp/iso_builder/ove-iso
	rm -rf ove-assets/

cleanall: clean
	sudo rm -rf /tmp/iso_builder
	
.PHONY: build-ove-iso
build-ove-iso:
	OCP_RELEASE_IMAGE=${OCP_RELEASE_IMAGE} \
	PULL_SECRET=${PULL_SECRET} \
	ARCH=${ARCH} \
	hack/build-ove-image.sh --release-image ${OCP_RELEASE_IMAGE} --arch ${ARCH} --pull-secret ${PULL_SECRET}

clean-isobuilder:
	rm -rf ${OUTPUT}

build-isobuilder:
	podman build -f ../../images/Dockerfile.isobuilder . -t $(IMAGE)

run-isobuilder:
	@:$(call check_defined, OUTPUT)
	@:$(call check_defined, PULL_SECRET_FILE)
	@:$(call check_defined, RELEASE_IMAGE_URL)
	podman run --rm -it \
		-v $(OUTPUT):/tmp/iso_builder:Z \
		-v $(PULL_SECRET_FILE):/secrets/pull-secret.json:Z \
		--privileged \
		--net=host \
		$(IMAGE) --pull-secret /secrets/pull-secret.json --release-image $(RELEASE_IMAGE_URL)
