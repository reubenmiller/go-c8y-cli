#!/bin/bash

set -euo pipefail

check_image () {
    local image=$1
    docker build -t c8y-inttest --build-arg IMAGE="$image" -f tools/integration-tests/linux.dockerfile tools/integration-tests
    docker run --rm c8y-inttest /test/install.sh
}

check_linuxbrew_image () {
    local image=$1
    docker build -t c8y-linuxbrew-inttest --build-arg IMAGE="$image" -f tools/integration-tests/linuxbrew.dockerfile tools/integration-tests
    docker run --rm c8y-linuxbrew-inttest /test/install.homebrew.sh
}

# Home brew (linux) installation
check_linuxbrew_image "ubuntu:latest"

# apk
check_image "alpine:3.13"
check_image "alpine:3.12"
check_image "alpine:3.11"

# Debian based
check_image "debian:10"
check_image "debian:9"
check_image "debian:8"
check_image "ubuntu:latest"
check_image "ubuntu:focal"
check_image "ubuntu:bionic"

# rpm
check_image "fedora:35"
check_image "fedora:34"
check_image "centos:latest"
check_image "centos:7"
