#!/bin/bash

function version() {
	echo "$@" | awk -F. '{ printf("%d%03d%03d%03d\n", $1,$2,$3,$4); }'
}

function go_version_check() {
	declare -r minimum_go_version=1.18
	declare -r current_go_version=$(go version | cut -d' ' -f3)

	if [ "$(version "${current_go_version#go}")" -lt "$(version "$minimum_go_version")" ]; then
		echo "Go version should be greater or equal to $minimum_go_version"
		exit 1
	fi
}

function build() {
	declare -r mode=${MODE:-release}

	case "$mode" in
	release)
		declare ldflags="${LDFLAGS} -w -s"
		;;
	dev)
		declare ldflags="${LDFLAGS}"
		;;
	*)
		(echo >&2 "Unrecognized build mode")
		;;
	esac

	declare -r repo_dir="$(dirname "$(realpath "$(dirname -- "${BASH_SOURCE[0]}")")")"
	declare -r build_dir="${repo_dir}/bin"
	mkdir -p "$build_dir"
	GO_ENABLED=1 CGO_CFLAGS="$(pkg-config nmstate --cflags)" CGO_LDFLAGS="$(pkg-config nmstate --libs)" go build -ldflags "$ldflags" -o "${build_dir}/agent-tui" tools/agent_tui/main/main.go
}

# Only run this if not being sourced
if ! (return 0 2>/dev/null); then
	go_version_check
	build
fi
