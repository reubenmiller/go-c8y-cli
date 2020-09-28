#!/bin/bash
#######################################################################
# Publish c8y cli docker images to github container registry (ghcr.io)
#
# Required variables:
#  CR_PAT: Github container registry token
#######################################################################
TARGET_OWNER=${TARGET_OWNER:-"reubenmiller"}
CR_PAT=${CR_PAT:-""}

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
    VERSION=${2:-latest}

    docker tag ${SOURCE_IMAGE_NAME}:${VERSION} ghcr.io/${TARGET_OWNER}/${TARGET_IMAGE_NAME}:${VERSION}
    docker push ghcr.io/${TARGET_OWNER}/${TARGET_IMAGE_NAME}:${VERSION}
}

login_ghcr

publish_ghcr_docker c8y-bash $VERSION
publish_ghcr_docker c8y-zsh $VERSION
publish_ghcr_docker c8y-bash $VERSION

# also use latest tag
publish_ghcr_docker c8y-bash latest
publish_ghcr_docker c8y-zsh latest
publish_ghcr_docker c8y-bash latest
