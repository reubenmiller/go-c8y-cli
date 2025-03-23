#!/bin/sh
set -e

DRY_RUN=${DRY_RUN:-}

# Set shell used by the script (can be overwritten during dry run mode)
sh_c='sh -c'
user_sh_c='sh -c'
PACKAGE_MANAGER="${PACKAGE_MANAGER:-}"
VERSION="${VERSION:-latest}"
SHELLS="${SHELLS:-}"
ARCH="${ARCH:-}"
INSTALL_PATH="${INSTALL_PATH:-$HOME/bin}"

usage() {
    cat <<EOF
Install go-c8y-cli which is a CLI tool for Cumulocity

USAGE:
    $0 [--dry-run] [--package-manager <apt|apk|rpm|homebrew|golang|tarball>]

OPTIONS:
    -p, --package-manager string  Package manager to use to install thin-edge.io. Defaults to auto detection
                                  Available: apt, apk, rpm, homebrew, golang, tarball
    --version <version>           Version to install. Defaults to 'latest'
    --install-path <path>         Install path (only used with the 'tarball' package manager)
    --dry-run                     Don't install anything, just let me know what it does
    --help, -h                    Show this help

EOF
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

log() {
    echo "$@" >&2
}

get_latest_release() {
    if command_exists curl; then
        latest_version_url="$(curl -s https://api.github.com/repos/reubenmiller/go-c8y-cli/releases/latest | grep "browser_download_url.*deb" | cut -d : -f 2,3 | tr -d \" | tail -n1)"
        TAG="$(basename "$(dirname "$latest_version_url")" ||:)"
    fi

    if [ -z "$TAG" ]; then
        if command_exists git; then
            TAG=$(git ls-remote --tags "https://github.com/reubenmiller/go-c8y-cli" | cut -d/ -f3- | grep "[0-9]$" | sort -V | tail -n1)
        fi
    fi
    echo "$TAG" | sed 's/^v//g'
}

only_support_latest() {
    if [ "$VERSION" != latest ]; then
        log "WARNING: The current package manager (${PACKAGE_MANAGER}) does not support installing older versions. Try adding '-p go' or '-p tarball' instead"
    fi
}

is_dry_run() {
	if [ -z "$DRY_RUN" ]; then
		return 1
	else
		return 0
	fi
}

DEBIAN_PUBLIC_KEY="
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
pNxthT4=
=4iTS
-----END PGP PUBLIC KEY BLOCK-----
"

ALPINE_RSA_PUBLIC_KEY="
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmHLzkJqiTWDYLhlGxxPq
1JyGisSxh4ke+two4auYVuMxpUffBlcOCF6nKcnkPjU6XqajTbGFzn4SrT1ViePi
KLB+2LdnFkQwd3ZYoRNZzNJtNeaa5DJ+o7zrjmW74JRZ/h5YhtM2kJZ6vs2ldIt9
sw98eeOTl7hUG8dfwLtRp6NOAF9xc5OF/sJm3TG7rDbbcni+hTpI8AM31c+m9xun
to80o3j/t4CnCTUUMczXaMN93SKTFT/w8tlG2klGvx9pZZSYN2IEjd7lYFrEuwu4
/6rjdoSrK39G/O1Xbn8nj4dsFAA0EtCxFDxTTcA9FjYANAuvbhWprBvt1fhs67zE
3wIDAQAB
-----END PUBLIC KEY-----
"

configure_shell() {
    if [ "$1" = root ]; then
        # Check if has sudo rights or if it can be requested
        user="$(id -un 2>/dev/null || true)"
        sh_c='sh -c'
        if [ "$user" != 'root' ]; then
            if command_exists sudo; then
                sh_c='sudo -E sh -c'
            elif command_exists su; then
                sh_c='su -c'
            else
                cat >&2 <<-EOF
Error: this installer needs the ability to run commands as root.
We are unable to find either "sudo" or "su" available to make this happen.
EOF
                exit 1
            fi
        fi
    fi

    user_sh_c='sh -c'

    if is_dry_run; then
        sh_c="echo"
        user_sh_c="echo"
    fi
}

install_debian() {
    $sh_c "apt-get update && apt-get install -y gnupg2 apt-transport-https"

    case "$ID" in
        debian)
            if [ "$VERSION_ID" -ge 9 ]; then
                $sh_c "echo '$DEBIAN_PUBLIC_KEY' | gpg --dearmor > /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg"
                $sh_c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            else
                $sh_c "echo '$DEBIAN_PUBLIC_KEY' | apt-key add -"
                $sh_c "echo 'deb https://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            fi
            ;;
        ubuntu)
            VERSION_ID=""
            MAJOR_VERSION=$(echo "$VERSION_ID" | cut -d. -f1)
            if [ -z "$MAJOR_VERSION" ]; then
                MAJOR_VERSION=22
            fi
            if [ "$MAJOR_VERSION" -ge 16 ]; then
                $sh_c "echo '$DEBIAN_PUBLIC_KEY' | gpg --dearmor > /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg"
                $sh_c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            else
                $sh_c "echo '$DEBIAN_PUBLIC_KEY' | apt-key add -"
                $sh_c "echo 'deb https://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            fi
            ;;
    esac

    $sh_c "apt-get update"
    $sh_c "apt-get install -y go-c8y-cli $EXTRA_PACKAGES"
}

install_rpm() {
    CONTENTS=$(cat <<EOT
[go-c8y-cli]
name=go-c8y-cli packages
baseurl=https://reubenmiller.github.io/go-c8y-cli-repo/rpm/stable
enabled=1
gpgcheck=1
gpgkey=https://reubenmiller.github.io/go-c8y-cli-repo/rpm/PUBLIC.KEY
EOT
    )
    $sh_c "echo '$CONTENTS' > /etc/yum.repos.d/go-c8y-cli.repo"
    $sh_c "dnf update"
    $sh_c "dnf install -y go-c8y-cli $EXTRA_PACKAGES"
}

install_alpine() {
    $sh_c "echo '$ALPINE_RSA_PUBLIC_KEY' > /etc/apk/keys/reuben.d.miller\@gmail.com-61e3680b.rsa.pub"
    $sh_c "echo 'https://reubenmiller.github.io/go-c8y-cli-repo/alpine/stable/main' >> /etc/apk/repositories"
    $sh_c "apk update"
    $sh_c "apk add go-c8y-cli $EXTRA_PACKAGES"
}

install_homebrew() {
    $user_sh_c "brew tap reubenmiller/go-c8y-cli"
    if brew list go-c8y-cli >/dev/null 2>&1; then
        $user_sh_c "brew update"
        $user_sh_c "brew upgrade go-c8y-cli"
    else
        $user_sh_c "brew update"
        $user_sh_c "brew install go-c8y-cli"
    fi
}

check_bin_path_in_shells() {
    CHECK_BIN_PATH=
    if command_exists bash; then
        CHECK_BIN_PATH=$(bash -c 'which c8y' 2>/dev/null ||:)
    elif command_exists zsh; then
        CHECK_BIN_PATH=$(zsh -c 'which c8y' 2>/dev/null ||:)
    elif command_exists fish; then
        CHECK_BIN_PATH=$(fish -c 'which c8y' 2>/dev/null ||:)
    fi
    echo "$CHECK_BIN_PATH"
}

install_golang() {
    if ! command_exists go; then
        log "could not find golang (go) on your system. Please install it and try again"
        exit 1
    fi

    $user_sh_c "go install github.com/reubenmiller/go-c8y-cli/v2/cmd/c8y@$VERSION"

    INSTALL_PATH=$(go env GOBIN)
    if [ -z "$INSTALL_PATH" ]; then
        INSTALL_PATH="$(go env GOPATH)/bin"
    fi
    export PATH="${INSTALL_PATH}:$PATH"

    CHECK_BIN_PATH=$(check_bin_path_in_shells)
    if [ -z "$CHECK_BIN_PATH" ]; then
        cat <<EOT >&2
Add the following line to your shell profile, and then reload it

  export PATH="\$(go env GOPATH)/bin:$PATH"

EOT
    fi
}

get_arch() {
    if [ -z "$ARCH" ]; then
        if [ -n "$TARGETARCH" ]; then
            # Support reading from TARGETARCH as it is used by docker buildx
            case "$TARGETARCH" in
                *amd64*) ARCH="amd64" ;;
                *arm64*) ARCH="arm64" ;;
                *arm/v7*) ARCH="armv7" ;;
                *arm/v6*) ARCH="armv6" ;;
                *)
                    # Unknown normalization, but still try to use it
                    ARCH="$TARGETARCH"
                    ;;
            esac
            log "Detected buildx TARGETARCH=$TARGETARCH. Using ARCH=$ARCH"
        else
            ARCH="$(uname -m)"
        fi
    fi
    echo "$ARCH"
}

install_tarball() {
    $user_sh_c "brew tap reubenmiller/go-c8y-cli"
    mkdir -p "$INSTALL_PATH"

    OS=$(uname | tr '[:upper:]' '[:lower:]' ||:)
    TARGET_OS=
    case "$OS" in
        darwin)
            TARGET_OS="macOS"
            ;;
        *win*)
            TARGET_OS="windows"
            ;;
        *)
            TARGET_OS="linux"
            ;;
    esac

    ARCH=$(get_arch)
    TARGET_ARCH=
    case "$ARCH" in
        *86_64*|*amd64*)
            TARGET_ARCH="amd64"
            ;;
        *aarch64*|*arm64*)
            TARGET_ARCH="arm64"
            ;;
        *armv7*)
            TARGET_ARCH="armv7"
            ;;
        *armv6*)
            TARGET_ARCH="armv6"
            ;;
        *)
            fail 1 "Unsupported architecture: $ARCH. Supported architectures are: amd64, arm64, armv7, armv6"
            ;;
    esac

    # strip prefix
    case "$VERSION" in
        v*)
            VERSION=$(echo "$VERSION" | sed 's/^v//g')
            ;;
        latest)
            VERSION=$(get_latest_release)
            ;;
    esac

    BASEDIR="c8y_${VERSION}_${TARGET_OS}_${TARGET_ARCH}"
    DOWNLOAD_URL="https://github.com/reubenmiller/go-c8y-cli/releases/download/v${VERSION}/${BASEDIR}.tar.gz"
    log "Downloading version: $VERSION, url=$DOWNLOAD_URL"

    if command_exists curl; then
        $user_sh_c "curl -sSLf '$DOWNLOAD_URL' > '$INSTALL_PATH/c8y.tar.gz'"
        
    elif command_exists wget; then
        $user_sh_c "wget -O - '$DOWNLOAD_URL' > '$INSTALL_PATH/c8y.tar.gz'"
    fi

    $user_sh_c "cd '$INSTALL_PATH' && tar xzf '$INSTALL_PATH/c8y.tar.gz' --strip-components=2 $BASEDIR/bin/c8y 2>/dev/null"
    $user_sh_c "rm -f '$INSTALL_PATH/c8y.tar.gz'"

    chmod +x "$INSTALL_PATH/c8y"
    export PATH="${INSTALL_PATH}:$PATH"

    CHECK_BIN_PATH=$(check_bin_path_in_shells)
    if [ -z "$CHECK_BIN_PATH" ]; then
        cat <<EOT >&2
Add the following line to your shell profile, and then reload it

  export PATH="\$(go env GOPATH)/bin:$PATH"

EOT
    fi
}

post_install() {
    $user_sh_c "c8y cli install"
}

main() {
    if [ -f /etc/os-release ]; then
        # shellcheck disable=SC1091
        . /etc/os-release
    fi

    add_extra_packages

    if [ -z "$PACKAGE_MANAGER" ]; then
        # auto detect pakage
        if command_exists apt-get; then
            PACKAGE_MANAGER=debian
        elif command_exists apk; then
            PACKAGE_MANAGER=alpine
        elif command_exists dnf; then
            PACKAGE_MANAGER=rpm
        elif command_exists brew; then
            PACKAGE_MANAGER=brew
        elif command_exists go; then
            PACKAGE_MANAGER=golang
        else
            PACKAGE_MANAGER=tarball
        fi
    fi

    log "Using package manager: $PACKAGE_MANAGER"
    case "$PACKAGE_MANAGER" in
        debian)
            only_support_latest
            configure_shell root
            install_debian
            ;;
        alpine)
            only_support_latest
            configure_shell root
            install_alpine
            ;;
        rpm)
            only_support_latest
            configure_shell root
            install_rpm
            ;;
        brew|homebrew)
            only_support_latest
            configure_shell user
            install_homebrew
            ;;
        go|golang)
            configure_shell user
            install_golang
            ;;
        tarball|*)
            configure_shell user
            install_tarball
            ;;
    esac

    post_install
}

add_extra_packages() {
    if [ -n "$SHELLS" ]; then
        for name in $SHELLS; do
            case "$name" in
                zsh)
                    EXTRA_PACKAGES="$EXTRA_PACKAGES zsh"
                    ;;
                bash)
                    EXTRA_PACKAGES="$EXTRA_PACKAGES bash"
                    ;;
                fish)
                    EXTRA_PACKAGES="$EXTRA_PACKAGES fish"
                    ;;
                *)
                    echo "WARNING: Unknown shell. $name" >&2
                    ;;
            esac
        done
    fi

    # Add bash-completion if bash is detected (as a minimum)
    if command_exists bash; then
        EXTRA_PACKAGES="$EXTRA_PACKAGES bash-completion"
    fi
}

while [ $# -gt 0 ]; do
    case $1 in
        --dry-run)
            DRY_RUN=1
            ;;
        --shell)
            if [ -n "$2" ]; then
                SHELLS="$2"
            fi
            shift
            ;;
        --version)
            if [ -n "$2" ]; then
                VERSION="$2"
            fi
            shift
            ;;
        -p|--package-manager)
            if [ -n "$2" ]; then
                PACKAGE_MANAGER="$2"
            fi
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        --*|-*)
            log "Unknown option: $1"
            usage
            exit 1
            ;;
        *)
            ;;
    esac
    shift $(( $# > 0 ? 1 : 0 ))
done

main "$@"
