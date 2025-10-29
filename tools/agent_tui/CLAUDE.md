# CLAUDE.md - Agent TUI

## Overview

Agent TUI is a Terminal User Interface for agent-based OpenShift installations. It validates that the installation environment meets prerequisites before proceeding. If all checks pass, the TUI exits and installation continues. If checks fail, users can reconfigure the environment interactively.

## Build and Run

### Dependencies

Requires nmstate-devel package (install with unsandboxed Bash):
```bash
sudo dnf install nmstate-devel  # Fedora/RHEL/CentOS
```

The build uses CGO and links against nmstate via pkg-config.

### Build Commands

```bash
# From repository root
make build
RELEASE_IMAGE=quay.io/openshift-release-dev/ocp-release:4.12.0-rc.7-x86_64 ./bin/agent-tui

# Or using make run
make run
```

## Architecture

### Entry Point
- `main/main.go` - Application entry point

### Core Components

**app.go** - Main application initialization
- `App()` function initializes the TUI application
- Sets up logging to config.LogPath
- Creates UI with custom tview theme (newt color scheme)
- Initializes Controller and Engine
- Coordinates UI ↔ Engine communication via channels

**checks/engine.go** - Check execution engine
- `Engine` runs periodic connectivity checks every 5 seconds
- Check types:
  - `ReleaseImagePull` - validates `podman pull` works
  - `ReleaseImageHostDNS` - validates DNS resolution via `nslookup`
  - `ReleaseImageHostPing` - validates network connectivity via `ping -c 4`
  - `ReleaseImageHttp` - validates HTTP connectivity to release image host
- Checks are skipped entirely if `/etc/assisted/registry.env` exists (local registry present)
- Each check runs in its own goroutine, sending results to shared channel
- `CheckFunctions` map allows injecting custom check implementations for testing

**ui/controller.go** - UI coordination
- Mediates between UI components and checks engine
- Receives check results via channel
- Updates UI state based on results

**ui/check_page.go** - Check results display
- Shows status of each check (success/failure)
- Displays failure details when checks fail

**ui/rendezvous_ip*.go** - Rendezvous IP management
- `rendezvous_ip_select.go` - IP selection interface
- `rendezvous_ip.go` - IP validation and configuration
- `rendezvous_ip_connectivity_fail_modal.go` - Error display
- `rendezvous_ip_save_success_modal.go` - Success confirmation
- `rendezvous_ip_template.go` - Template rendering

**ui/netstate*.go** - Network state configuration
- `netstate_treeview.go` - Network interface tree display
- `netstate_modal.go` - Configuration modal
- Uses nmstate library for network configuration

**ui/nmtui.go** - nmtui integration
- Launches external nmtui tool for advanced network configuration

**checks/url_parser.go** - URL utilities
- `ParseHostnameFromURL()` - extracts hostname from release image URL
- `ParseSchemeHostnamePortFromURL()` - extracts scheme://hostname:port

**newt/colors.go** - Color theme definitions
- Custom color scheme (ColorGray, ColorBlue, ColorBlack)

## Configuration

`checks.Config` structure:
- `ReleaseImageURL` - URL to OpenShift release image (required)
- `LogPath` - Path for TUI logs (default: /var/log/agent-tui.log)
- `ReleaseImageHostname` - Parsed from ReleaseImageURL
- `ReleaseImageSchemeHostnamePort` - Parsed from ReleaseImageURL

## Testing

Tests use testify for assertions:
```bash
go test ./...
```

Test files:
- `app_test.go`, `apptester_test.go` - Application testing
- `checks/url_parser_test.go` - URL parsing tests
- `ui/check_page_test.go` - Check page UI tests
- `ui/rendezvous_ip_select_test.go` - Rendezvous IP selection tests

Mock implementations can be injected via `CheckFunctions` parameter to `App()`.

## Key Libraries

- `github.com/rivo/tview` - Terminal UI framework
- `github.com/gdamore/tcell/v2` - Terminal cell-based view
- `github.com/nmstate/nmstate/rust/src/go/nmstate/v2` - Network state management
- `github.com/sirupsen/logrus` - Structured logging
