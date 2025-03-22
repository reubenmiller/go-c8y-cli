#!/bin/sh
set -e

DRY_RUN=${DRY_RUN:-}

# Set shell used by the script (can be overwritten during dry run mode)
sh_c='sh -c'
user_sh_c='sh -c'

usage() {
    cat <<EOF
USAGE:
    $0 [--dry-run] [--package-manager <apt|apk|dnf|microdnf|zypper|tarball>]

OPTIONS:
    -p, --package-manager string  Package manager to use to install thin-edge.io. Defaults to auto detection
                                  Available: apt, apk, dnf, microdnf, zypper, tarball
    --dry-run                     Don't install anything, just let me know what it does
    --help                        Show this help

EOF
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

log() {
    echo "$@" >&2
}

is_dry_run() {
	if [ -z "$DRY_RUN" ]; then
		return 1
	else
		return 0
	fi
}

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

if [ -f /etc/os-release ]; then
    . /etc/os-release
fi

install_debian() {
    $sh_c "apt-get install -y curl gnupg2 apt-transport-https"

    case "$ID" in
        debian)
            if [ "$VERSION_ID" -ge 9 ]; then
                $sh_c "curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | gpg --dearmor > /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg"
                $sh_c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            else
                $sh_c "curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | sudo apt-key add -"
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
                $sh_c "curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | gpg --dearmor > /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg"
                $sh_c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            else
                $sh_c "curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | sudo apt-key add -"
                $sh_c "echo 'deb https://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
            fi
            ;;
    esac

    sudo apt-get update
    sudo apt-get install go-c8y-cli
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
    $sh_c "dnf install go-c8y-cli"
}

install_alpine() {
    if ! command_exists wget; then
        $sh_c "apk add wget"
    fi
    $sh_c "wget -O /etc/apk/keys/reuben.d.miller\@gmail.com-61e3680b.rsa.pub https://reubenmiller.github.io/go-c8y-cli-repo/alpine/PUBLIC.KEY"
    $sh_c "echo 'https://reubenmiller.github.io/go-c8y-cli-repo/alpine/stable/main' >> /etc/apk/repositories"
    $sh_c "apk update"
    $sh_c "apk add go-c8y-cli"
}

install_homebrew() {
    $user_sh_c "brew tap reubenmiller/go-c8y-cli"
    if command_exists c8y; then
        $user_sh_c "brew update"
        $user_sh_c "brew upgrade go-c8y-cli"
    else
        $user_sh_c "brew update"
        $user_sh_c "brew install go-c8y-cli"
    fi
}

post_install() {
    $user_sh_c "c8y cli install"
}

main() {
    # http://localhost:3000/install.sh
    if command_exists apt-get; then
        configure_shell root
        install_debian
    elif command_exists apk; then
        configure_shell root
        install_alpine
    elif command_exists dnf; then
        configure_shell root
        install_rpm
    elif command_exists brew; then
        configure_shell user
        install_homebrew
    else
        log "Unsupported environment"
        exit 1
    fi

    post_install
}



while [ $# -gt 0 ]; do
    case $1 in
        --dry-run)
            DRY_RUN=1
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
