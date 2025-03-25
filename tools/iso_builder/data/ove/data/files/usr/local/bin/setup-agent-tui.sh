#!/bin/bash
set -e

/usr/sbin/unsquashfs -f -d /usr/local/bin /run/media/iso/agent-artifacts/agent-artifacts.squashfs
restorecon -FRv /usr/local/bin
