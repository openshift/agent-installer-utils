ARCH ?= x86_64

.PHONY:clean
clean:
	rm -rf /tmp/ove/
	rm -rf ove-assets/

cleanall: clean
	rm -rf /tmp/appliance-assets*
	
.PHONY: build-ove-iso
build-ove-iso:
	OCP_RELEASE_IMAGE=${OCP_RELEASE_IMAGE} \
	PULL_SECRET=${PULL_SECRET} \
	ARCH=${ARCH} \
	hack/build-ove-image.sh --release-image ${OCP_RELEASE_IMAGE} --arch ${ARCH} --pull-secret ${PULL_SECRET}
