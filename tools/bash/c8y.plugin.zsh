#!/bin/zsh

#mkdir -p "$ZSH/completions"
#c8y completion zsh > ~/.oh-my-zsh/completions/_c8y

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

    # Export session as individual settings
    # to support other 3rd party applicatsion (i.e. java c8y sdk apps)
    # which will read these variables
    session_info=$( cat "$C8Y_SESSION" )
    if [[ $(command -v jq) ]]; then
        export C8Y_HOST=$( echo $session_info | jq -r ".host" )
        export C8Y_TENANT=$( echo $session_info | jq -r ".tenant" )
        export C8Y_USER=$( echo $session_info | jq -r ".username" )
        export C8Y_USERNAME=$( echo $session_info | jq -r ".username" )
        export C8Y_PASSWORD=$( echo $session_info | jq -r ".password" )
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
    curl -L --silent https://github.com/reubenmiller/go-c8y-cli/releases/download/$VERSION/$BINARY_NAME -o $c8ytmp
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
    
    # update completions
    [ ! -d ~/.oh-my-zsh/completions ] && mkdir -p ~/.oh-my-zsh/completions

    if [ $(command -v c8y) ]; then
        echo -n "updating completions..."
        c8y completion zsh > $ZSH/completions/_c8y
        echo -e "${green}ok${normal}"
    fi

    echo -e "${green}Updated c8y completions. \n\n${bold}Please load your zsh profile again using 'source ~/.zshrc'${normal}"

    # show new version to user
    c8y version
}

########################################################################
# c8y aliases
########################################################################

# alarms
alias alarms=c8y\ alarms\ list

# apps
alias apps=c8y\ applications\ list

# devices
alias devices=c8y\ devices\ list

# events
alias events=c8y\ events\ list

# fmo
alias fmo=c8y\ inventory\ find\ --query

# measurements
alias measurements=c8y\ measurements\ list

# operations
alias ops=c8y\ operations\ list

# series
alias series=c8y\ measurements\ getSeries

#
# Single item getters
#
# alarm
alias alarm=c8y\ alarms\ get\ --id

# app
alias app=c8y\ applications\ get\ --id

# event
alias event=c8y\ events\ get\ --id

# m
alias m=c8y\ measurements\ get\ --id

# mo
alias mo=c8y\ inventory\ get\ --id

# op
alias op=c8y\ operations\ get\ --id
