#!/bin/bash

SOURCE_PATH="${1:-"./dist"}"
JFROG_URL="https://reubenmiller.jfrog.io/artifactory"
COMPONENT="main"
DEBIAN_REPO="c8y-debian"
ALPINE_REPO="c8y-alpine"
RPM_REPO="c8y-rpm"
POOL_PATH="pool/go-c8y-cli"
export CI=true

if [[ "$JFROG_APIKEY" == "" ]]; then
    echo "JFROG_APIKEY environment variable is not defined"
    exit 1
fi

if [[ "$JFROG_URL" == "" ]]; then
    echo "JFROG_URL environment variable is not defined"
    exit 1
fi

install_dependencies () {
    if ! command -v jfrog; then
        curl -fL https://getcli.jfrog.io | sh
        chmod a+x jfrog
        mkdir -p /tmp/bin/
        mv jfrog /tmp/bin/
        export PATH="/tmp/bin:$PATH"
    fi
}

#----------------------------------------------#
# Debian
#----------------------------------------------#
publish_debian_package () {
    local file="$1"
    local arch="$2"
    local DISTRIBUTIONS="cosmic eoan disco groovy focal stable oldstable testing sid unstable buster bullseye stretch jessie bionic trusty precise xenial hirsute impish kali-rolling"
    local TARGET_PROPS="$(echo "$DISTRIBUTIONS" | sed 's/ /,/g')"
    jfrog rt upload \
        --url "$JFROG_URL/$DEBIAN_REPO" \
        --target-props "deb.distribution=$TARGET_PROPS;deb.component=$COMPONENT;deb.architecture=$arch" \
        --access-token "$JFROG_APIKEY" \
        "$file" \
        "$POOL_PATH/"
}

publish_all_debian_packages () {
    ARCHITECTURES="i386:386 amd64 arm64 armel:armv5 armel:armv6 armhf:armv7"

    for arch in $ARCHITECTURES; do
        arch_name=$(echo "$arch" | cut -d':' -f1)
        arch_filter=$(echo "$arch" | rev | cut -d':' -f 1 | rev)
        for file in dist/*$arch_filter*.deb; do
            publish_debian_package "$file" "$arch_name"
        done
    done
}

#----------------------------------------------#
# Alpine
#----------------------------------------------#
publish_alpine_package () {
    local file="$1"
    local arch="$2"
    local BRANCHES="stable"
    # local BRANCHES="v3.9 v3.10 v3.11 v3.12 v3.13 v3.14"
    local BRANCH="$(echo "$BRANCHES" | sed 's/ /,/g')"
    local REPO_NAME="main"

    # package name needs to be specially formatted
    # otherwise the index will not match, and apk will have
    # issues
    local name="$(basename "$file" | cut -d'_' -f1 )"
    local version="$(basename "$file" | cut -d'_' -f2 | sed 's/-/~/')"
    local package_name="$name-$version.apk"

    jfrog rt upload \
        --url "$JFROG_URL/$ALPINE_REPO" \
        --target-props "alpine.branch=$BRANCH;alpine.repository=$REPO_NAME;alpine.architecture=$arch;alpine.name=$name;alpine.version=$version" \
        --access-token "$JFROG_APIKEY" \
        "$file" \
        "$BRANCH/$REPO_NAME/$arch/$package_name"
}

publish_all_alpine_packages () {
    ARCHITECTURES="x86:386 x86_64:amd64 aarch64:arm64 armel:armv5 armhf:armv6 armv7:armv7"

    for arch in $ARCHITECTURES; do
        arch_name=$(echo "$arch" | cut -d':' -f1)
        arch_filter=$(echo "$arch" | rev | cut -d':' -f 1 | rev)
        for file in dist/*$arch_filter*.apk; do
            publish_alpine_package "$file" "$arch_name"
        done
    done
}

#----------------------------------------------#
# RPM
#----------------------------------------------#
publish_rpm_package () {
    local file="$1"
    local arch="$2"
    local DISTRIBUTIONS="cosmic eoan disco groovy focal stable oldstable testing sid unstable buster bullseye stretch jessie bionic trusty precise xenial hirsute impish kali-rolling"
    local TARGET_PROPS="$(echo "$DISTRIBUTIONS" | sed 's/ /,/g')"
    jfrog rt upload \
        --url "$JFROG_URL/$RPM_REPO" \
        --access-token "$JFROG_APIKEY" \
        "$file" \
        "$POOL_PATH/"
}

publish_all_rpm_packages () {
    ARCHITECTURES="i386:386 amd64 arm64 armel:armv5 armel:armv6 armhf:armv7"

    for arch in $ARCHITECTURES; do
        arch_name=$(echo "$arch" | cut -d':' -f1)
        arch_filter=$(echo "$arch" | rev | cut -d':' -f 1 | rev)
        for file in dist/*$arch_filter*.rpm; do
            publish_rpm_package "$file" "$arch_name"
        done
    done
}


#----------------------------------------------#
# Main
#----------------------------------------------#
install_dependencies
publish_all_debian_packages
publish_all_alpine_packages
publish_all_rpm_packages
