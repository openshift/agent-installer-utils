#!/bin/bash
set -e

/usr/sbin/unsquashfs -f -d /usr/local/bin /run/media/iso/agent-artifacts/agent-artifacts.squashfs

files=(/usr/local/bin/agent-tui /usr/local/bin/libnmstate.so.*)

for file in "${files[@]}"; do
    chcon system_u:object_r:bin_t:s0 "$file"
done
