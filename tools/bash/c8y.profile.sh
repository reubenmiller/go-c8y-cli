#!/bin/bash

if [[ ! -d ~/.bash_completion.d ]]; then
    mkdir -p ~/.bash_completion.d
fi

if [ ! -f ~/.bash_completion.d/complete_alias ]; then
    echo "Installing bash completion for aliases"
    curl https://raw.githubusercontent.com/cykerway/complete-alias/master/complete_alias \
            > ~/.bash_completion.d/complete_alias
fi

# Enable completion for aliases
[ -f /usr/share/bash-completion/bash_completion ] && source /usr/share/bash-completion/bash_completion
[ -f ~/.bash_completion.d/complete_alias ] && source ~/.bash_completion.d/complete_alias

if [[ $(command -v c8y) ]]; then
    source <(c8y completion bash)
fi

########################################################################
# c8y helpers
########################################################################

# -----------
# test-c8ypassphrase
# -----------
# Description: Set the encryption passphrase interactively
# Usage:
#   test-c8ypassphrase
#
test-c8ypassphrase () {
    passphraseCheck=$( c8y sessions checkPassphrase --json )
    if [ $? -ne 0 ]; then
        echo "Encryption check failed"
        exit 2
    fi

    if [[ $(command -v jq) ]]; then
        export C8Y_PASSPHRASE=$( echo $passphraseCheck | jq -r ".passphrase" )
        export C8Y_PASSPHRASE_TEXT=$( echo $passphraseCheck | jq -r ".secretText" )
    fi
}

# -----------
# set-session
# -----------
# Description: Switch Cumulocity session interactively
# Usage:
#   set-session
#
set-session () {
    if [ $# -gt 0 ]; then
        resp=$( c8y sessions list --sessionFilter "$1 $2 $3 $4 $5" )
    else
        resp=$( c8y sessions list )
    fi

    if [ $? -ne 0 ]; then
        echo "Set session aborted"
        return
    fi

    export C8Y_SESSION=$resp

    # Check encryption passphrase
    passphraseCheck=$( c8y sessions checkPassphrase --json )

    if [ $? -ne 0 ]; then
        echo "Encryption check failed"
        exit 2
    fi

    # Export session as individual settings
    # to support other 3rd party applicatsion (i.e. java c8y sdk apps)
    # which will read these variables
    session_info=$( cat "$C8Y_SESSION" )
    if [[ $(command -v jq) ]]; then
        export C8Y_HOST=$( echo $session_info | jq -r ".host" )
        export C8Y_TENANT=$( echo $session_info | jq -r ".tenant" )
        export C8Y_USER=$( echo $session_info | jq -r ".username" )
        export C8Y_USERNAME=$( echo $session_info | jq -r ".username" )

        export C8Y_PASSPHRASE=$( echo $passphraseCheck | jq -r ".passphrase" )
        export C8Y_PASSPHRASE_TEXT=$( echo $passphraseCheck | jq -r ".secretText" )

        password=$( echo $session_info | jq -r ".password" )
        if [[ "$password" != "" ]]; then
            password=$( c8y sessions decryptText --text "$password" )
        fi
        export C8Y_PASSWORD=$password

        # login / test session credentials
        c8y sessions login
    fi

    # reset any enabled side-effect commands
    unset C8Y_SETTINGS_MODE_ENABLECREATE
    unset C8Y_SETTINGS_MODE_ENABLEUPDATE
    unset C8Y_SETTINGS_MODE_ENABLEDELETE
}

# ----------
# c8y-update
# ----------
# Description: Update the c8y binary
# The latest binary version will be downloaded from github
#
# Usage:
#   c8y-update
#
c8y-update () {
    bold="\e[1m"
    normal="\e[0m"
    red="\e[31m"
    green="\e[32m"

    VERSION=${1:-latest}
    INSTALL_PATH=${2:-~/bin}

    if [ ! -d "$INSTALL_PATH" ]; then
        mkdir -p "$INSTALL_PATH"
    fi

    current_version=$(c8y version 2> /dev/null | tail -1)

    # Get binary name based on os type
    BINARY_NAME=c8y.linux

    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        BINARY_NAME=c8y.linux
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        BINARY_NAME=c8y.macos
    elif [[ "$OSTYPE" == "cygwin" ]]; then
        BINARY_NAME=c8y.windows.exe
    elif [[ "$OSTYPE" == "msys" ]]; then
        BINARY_NAME=c8y.windows.exe
    elif [[ "$OSTYPE" == "linux"* ]]; then
        BINARY_NAME=c8y.linux
    else
        # assume windows
        BINARY_NAME=c8y.windows.exe
    fi

    # try to download latest c8y version
    echo -n "downloading ($BINARY_NAME)..."

    c8ytmp=./.c8y.tmp
    if [[ "$VERSION" = "latest" ]]; then
        curl -L --silent https://github.com/reubenmiller/go-c8y-cli/releases/latest/download/$BINARY_NAME -o $c8ytmp
    else
        curl -L --silent https://github.com/reubenmiller/go-c8y-cli/releases/download/$VERSION/$BINARY_NAME -o $c8ytmp
    fi
    chmod +x $c8ytmp

    new_version=$($c8ytmp version 2>/dev/null | tail -1)

    if [ "$new_version" = "" ]; then
        if [[ $(cat $c8ytmp | head -1 | grep ELF) ]]; then
            echo -e "${red}Failed download latest version: err=Unknown binary error${normal}"
        else
            echo -e "${red}Failed download latest version: err=$(cat .c8y.tmp | head -1)${normal}"
        fi
        rm -f .c8y.tmp
        return 1
    else
        echo -e "${green}ok${normal}"
        mv $c8ytmp $INSTALL_PATH/c8y
    fi

    if [[ ! $(command -v c8y) ]]; then
        echo "Adding install path ($INSTALL_PATH) to PATH variable"
        export PATH=${PATH}:$INSTALL_PATH
    fi

    if [ "$current_version" = "$new_version"]; then
        echo -e "${green}c8y is already up to date: $(current_version)${normal}"
        return 0
    fi

    # show new version to user
    c8y version
}

########################################################################
# c8y aliases
########################################################################

# alarms
alias alarms=c8y\ alarms\ list
complete -F _complete_alias alarms

# apps
alias apps=c8y\ applications\ list
complete -F _complete_alias apps

# devices
alias devices=c8y\ devices\ list
complete -F _complete_alias devices

# events
alias events=c8y\ events\ list
complete -F _complete_alias events

# fmo
alias fmo=c8y\ inventory\ find\ --query
complete -F _complete_alias fmo

# measurements
alias measurements=c8y\ measurements\ list
complete -F _complete_alias measurements

# operations
alias ops=c8y\ operations\ list
complete -F _complete_alias ops

# series
alias series=c8y\ measurements\ getSeries
complete -F _complete_alias series

#
# Single item getters
#
# alarm
alias alarm=c8y\ alarms\ get\ --id
complete -F _complete_alias alarm

# app
alias app=c8y\ applications\ get\ --id
complete -F _complete_alias app

# event
alias event=c8y\ events\ get\ --id
complete -F _complete_alias event

# m
alias m=c8y\ measurements\ get\ --id
complete -F _complete_alias m

# mo
alias mo=c8y\ inventory\ get\ --id
complete -F _complete_alias mo

# op
alias op=c8y\ operations\ get\ --id
complete -F _complete_alias op

# init passphrase (if not already set)
if [ -t 0 ]; then
    test-c8ypassphrase
fi
