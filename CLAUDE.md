# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agent-based installation tools for OpenShift, supporting the `openshift-install agent` workflow. This repository contains utilities and tools for agent-based OpenShift installations.

## Repository Structure

- `tools/agent_tui/` - Terminal User Interface for validating installation prerequisites (see `tools/agent_tui/CLAUDE.md`)
- `tools/iso_builder/` - ISO builder for creating bootable agent-based installer images (see `tools/iso_builder/CLAUDE.md`)
- `pkg/version/` - Version information package
- `hack/` - Build and utility scripts
- `vendor/` - Vendored Go dependencies

## Top-Level Build Commands

```bash
# Build agent-tui (requires nmstate-libs)
make build          # Runs clean, lint, and build via hack/build.sh
make lint           # Run golangci-lint
make clean          # Remove bin/ directory
make run            # Build and run agent-tui with RELEASE_IMAGE
```

The top-level Makefile builds the agent-tui tool. For iso_builder, change to `tools/iso_builder/` and use its Makefile.

## Development Requirements

- Go >= 1.18 (enforced in `hack/build.sh`)
- For agent-tui: nmstate-devel package (CGO dependency) - install with `sudo dnf install nmstate-devel` (requires unsandboxed Bash)
- For iso_builder: xorriso, isohybrid, oc, podman, skopeo - install with `sudo dnf install -y xorriso syslinux skopeo` (requires unsandboxed Bash)

## Version Management

Version information is injected via LDFLAGS during build:
- `pkg/version/version.go` defines `Raw` and `Commit` variables
- `hack/build.sh` sets these via `-ldflags` using `VERSION_URI`, `SOURCE_GIT_COMMIT`, and `BUILD_VERSION`
- Default values from Makefile: `SOURCE_GIT_COMMIT` from git HEAD, `BUILD_VERSION` from git describe

## Common Patterns

- CGO is enabled for agent-tui builds with nmstate linking
- Vendored dependencies are committed to the repository
