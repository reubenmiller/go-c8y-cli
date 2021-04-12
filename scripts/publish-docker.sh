#!/bin/bash
#######################################################################
# Publish c8y cli docker images to github container registry (ghcr.io)
#
# Required variables:
#  CR_PAT: Github container registry token
#######################################################################
TARGET_OWNER=${TARGET_OWNER:-"reubenmiller"}
CR_PAT=${CR_PAT:-""}

# Get version
if [[ -z $VERSION ]]; then
    VERSION=$( git describe )
fi
# strip git reference prefix
VERSION=${VERSION#refs/*/}
VERSION=${VERSION#v}

echo "Version: $VERSION"

login_ghcr () {
    echo $CR_PAT | docker login ghcr.io -u USERNAME --password-stdin

    if [ $? -ne 0 ]; then
        echo "ghcr.io login failed"
        exit 1
    fi
}

publish_ghcr_docker () {
    SOURCE_IMAGE_NAME=$1
    TARGET_IMAGE_NAME=$1
    tag=${2:-latest}

    docker tag ${SOURCE_IMAGE_NAME} ghcr.io/${TARGET_OWNER}/${TARGET_IMAGE_NAME}:${tag}
    docker push ghcr.io/${TARGET_OWNER}/${TARGET_IMAGE_NAME}:${tag}
}

login_ghcr

publish_ghcr_docker c8y-shell $VERSION
publish_ghcr_docker c8y-pwsh $VERSION

# also use latest (or next) tag
if [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$  ]]; then
    # official release
    publish_ghcr_docker c8y-shell latest
    publish_ghcr_docker c8y-pwsh latest
else
    # beta release
    publish_ghcr_docker c8y-shell next
    publish_ghcr_docker c8y-pwsh next
fi
