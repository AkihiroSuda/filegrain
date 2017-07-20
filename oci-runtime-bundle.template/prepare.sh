#!/bin/sh
set -x
set -e
mkdir -p rootfs
mkdir -p volumes/root
chown 0:0 volumes/root
mkdir -p volumes/home
chown 0:0 volumes/home
