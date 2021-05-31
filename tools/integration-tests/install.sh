#!/bin/sh

setup_debian () {
    apt-get update
    apt-get install -y curl gnupg2 apt-transport-https

    export APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=1

    curl https://reubenmiller.jfrog.io/artifactory/api/security/keypair/public/repositories/c8y-debian | apt-key add -
    sh -c "echo 'deb https://reubenmiller.jfrog.io/artifactory/c8y-debian stable main' >> /etc/apt/sources.list"

    apt-get update
    apt-get install -y c8y
}

setup_apk () {
    apk update
    apk add wget sudo
    wget -O /etc/apk/keys/rmiller-rsa-signing.rsa.pub https://reubenmiller.jfrog.io/artifactory/api/security/keypair/public/repositories/c8y-alpine
    sudo sh -c "echo 'https://reubenmiller.jfrog.io/artifactory/c8y-alpine/stable/main'" >> /etc/apk/repositories

    apk update
    apk add c8y
}

setup_rpm () {
    yum update
    yum install -y curl

    cat <<EOT > /etc/yum.repos.d/artifactory.repo
[Artifactory]
name=Artifactory
baseurl=https://reubenmiller.jfrog.io/artifactory/c8y-rpm
enabled=1
gpgcheck=0
gpgkey=https://reubenmiller.jfrog.io/artifactory/c8y-rpm/repodata/repomd.xml.key
repo_gpgcheck=1
EOT

    yum update
    yum install -y c8y
}

setup_linuxbrew () {
    apt-get update
    apt-get install -y git bash sudo curl procps gcc
    export CI=true
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    # import brew paths variables
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

    # install tap then install
    brew tap reubenmiller/go-c8y-cli
    brew install go-c8y-cli
}

check_installation () {
    command -v c8y || exit 1
    c8y version --output json -c
}

main () {
    echo "Setting up repository"
    if command -v yum; then
        setup_rpm
    elif command -v apk; then
        setup_apk
    elif command -v apt; then
        setup_debian
    fi

    echo "Verifying installation"
    check_installation
}

main
