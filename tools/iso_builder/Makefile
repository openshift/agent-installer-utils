.PHONY: cleanall clean-appliance-temp-dir build-ove-iso
ARCH ?= x86_64

cleanall:
	hack/cleanup.sh cleanall

clean-appliance-temp-dir:
	hack/cleanup.sh clean-appliance-temp-dir
	
build-ove-iso:
	hack/build-ove-image.sh --release-image-url ${RELEASE_IMAGE_URL} --pull-secret-file ${PULL_SECRET_FILE}
