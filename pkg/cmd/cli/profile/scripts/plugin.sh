#!/bin/bash
# Force encoding
export LANG=C.UTF-8
export LC_ALL=C.UTF-8

__setup_bash() {
    #
    # Install bash dependencies
    #
    if [ ! -d "$HOME/.bash_completion.d" ]; then
        mkdir -p "$HOME/.bash_completion.d"
    fi

    if [ ! -f "$HOME/.bash_completion.d/complete_alias" ]; then
        if command -V curl >/dev/null 2>&1; then
            echo "Installing bash completion for aliases"
            curl -sfL https://raw.githubusercontent.com/cykerway/complete-alias/master/complete_alias > "$HOME/.bash_completion.d/complete_alias"
        elif command -V curl >/dev/null 2>&1; then
            echo "Installing bash completion for aliases"
            wget -O - https://raw.githubusercontent.com/cykerway/complete-alias/master/complete_alias > "$HOME/.bash_completion.d/complete_alias"
        fi
    fi

    # Enable completion for aliases
    # shellcheck disable=SC1091
    [ -f /usr/share/bash-completion/bash_completion ] && source /usr/share/bash-completion/bash_completion
    # shellcheck disable=SC1091
    [ -f "$HOME/.bash_completion.d/complete_alias" ] && source "$HOME/.bash_completion.d/complete_alias"

    if [ -f /etc/bash_completion ]; then
        # shellcheck disable=SC1091
        . "/etc/bash_completion"
    elif command -v brew >/dev/null 2>&1; then
        if [ -f "$(brew --prefix)/etc/bash_completion" ]; then
            # shellcheck disable=SC1091
            . "$(brew --prefix)/etc/bash_completion"
        fi
    fi

    # shellcheck disable=SC1090
    source <(c8y completion bash)
}

__setup_zsh() {
    # homebrew only: Add completions folder to the fpath so zsh can find it
    if command -v brew >/dev/null 2>&1; then
        FPATH="$(brew --prefix)/share/zsh/site-functions:$FPATH"
        # TODO: Check if this line is really required
        # chmod -R go-w "$(brew --prefix)/share"
    fi

    # init completions
    autoload -U compinit; compinit

    IMPORT_LOCAL=1

    if [ -n "$ZSH_CUSTOM" ]; then
        # oh-my-zsh
        mkdir -p "$ZSH_CUSTOM/plugins/c8y"
        if [ ! -f "$ZSH_CUSTOM/plugins/c8y/_c8y" ]; then
            echo "Updating c8y completions: $ZSH_CUSTOM/plugins/c8y/_c8y"
            c8y completion zsh > "$ZSH_CUSTOM/plugins/c8y/_c8y"
            IMPORT_LOCAL=0
        fi
    else
        # zsh (vanilla)
        if [ ! -f "/usr/share/zsh/site-functions/_c8y" ]; then
            if [ "$EUID" = "0" ]; then
                if [ -d /usr/share/zsh/site-functions ]; then
                    echo "Updating c8y completions: /usr/share/zsh/site-functions/_c8y"
                    c8y completion zsh > "/usr/share/zsh/site-functions/_c8y"
                    IMPORT_LOCAL=0
                fi
            fi
        fi
    fi

    if [ "$IMPORT_LOCAL" = 1 ]; then
        # shellcheck disable=SC1090
        source <(c8y completion zsh)
    fi
}


#
# init
#
C8Y_SHELL=bash
if [ -n "$BASH_VERSION" ]; then
    C8Y_SHELL=bash
elif [ -n "$ZSH_VERSION" ]; then
    C8Y_SHELL=zsh
fi
if command -v c8y >/dev/null 2>&1; then
    case "$C8Y_SHELL" in
        bash)
            __setup_bash
            ;;
        zsh)
            __setup_zsh
            ;;
    esac
fi

########################################################################
# c8y helpers
########################################################################
# -----------
# set-session
# -----------
# Description: Switch Cumulocity session interactively
# Usage:
#   set-session
#
set-session() {
    c8yenv=$( c8y sessions set --noColor=false $@ )
    code=$?
    if [ $code -ne 0 ]; then
        echo "Set session failed"
        exit $code
    fi
    eval "$c8yenv"
}

# -------------
# clear-session
# -------------
# Description: Clear all cumulocity session variables
# Usage:
#   clear-session
#
clear-session() {
    c8yenv=$(c8y sessions clear)
    eval "$c8yenv"
}

# -------------------
# clear-c8ypassphrase
# -------------------
# Description: Clear the encryption passphrase environment variables
# Usage:
#   clear-c8ypassphrase
#
clear-c8ypassphrase() {
    unset C8Y_PASSPHRASE
    unset C8Y_PASSPHRASE_TEXT
}

# ----------------
# set-c8ymode-xxxx
# ----------------
# Description: Set temporary mode by setting the environment variables
# Usage:
#   set-c8ymode-dev     (enable PUT, POST and DELETE)
#   set-c8ymode-qual    (enable PUT, POST)
#   set-c8ymode-prod    (disable PUT, POST and DELETE)
#
set-c8ymode() {
    eval "$(c8y settings update --shell auto mode "$1")"
    printf "\e[32mEnabled %s mode (temporarily)\e[0m\n" "$1";
}
set-c8ymode-dev() { set-c8ymode dev; }
set-c8ymode-qual() { set-c8ymode qual; }
set-c8ymode-prod() { set-c8ymode prod; }
