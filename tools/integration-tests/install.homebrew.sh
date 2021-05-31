#!/bin/sh

setup_linuxbrew () {
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
    setup_linuxbrew

    echo "Verifying installation"
    check_installation
}

main
