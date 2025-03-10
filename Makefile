# See pkg/version/version.go for details
SOURCE_GIT_COMMIT ?= $(shell git rev-parse --verify 'HEAD^{commit}')
BUILD_VERSION ?= $(shell git describe --always --abbrev=40 --dirty)

VERSION_URI ?= github.com/openshift/agent-installer-utils/pkg/version
RELEASE_IMAGE ?= registry.ci.openshift.org/ocp/release:4.19.0-0.nightly-2025-03-05-160850
ARCH ?= x86_64
PULL_SECRET ?= /home/test/dev-scripts/pull_secret.json


.PHONY:clean
clean:
	rm -rf bin/

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: build
build: clean lint
	hack/build.sh ${VERSION_URI} ${SOURCE_GIT_COMMIT} ${BUILD_VERSION}

.PHONY: run
run: build
	RELEASE_IMAGE=${RELEASE_IMAGE} SOURCE_GIT_COMMIT=${SOURCE_GIT_COMMIT} BUILD_VERSION=${BUILD_VERSION} ./bin/agent-tui

.PHONY: build-ove-iso
build-ove-iso:
	tools/iso_builder/hack/build-ove-image.sh --release-image ${RELEASE_IMAGE} --arch ${ARCH} --pull-secret ${PULL_SECRET}