# See pkg/version/version.go for details
SOURCE_GIT_COMMIT ?= $(shell git rev-parse --verify 'HEAD^{commit}')
BUILD_VERSION ?= $(shell git describe --always --abbrev=40 --dirty)

VERSION_URI ?= github.com/openshift/agent-installer-utils/pkg/version
RELEASE_IMAGE ?= quay.io/openshift-release-dev/ocp-release:4.18.4-x86_64
ARCH ?= x86_64

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