#!/usr/bin/env fish

if type -q c8y
    c8y completion fish | source

    # create session home folder (if it does not exist)
    set sessionhome ( c8y settings list --select "session.home" --output csv )
    if test ! -e "$sessionhome"
        echo "creating folder"
        mkdir -p "$sessionhome"
    end
end

########################################################################
# c8y helpers
########################################################################
# -------------
# session
# -------------
# Description: Get the current cumulocity session
# Usage:
#   session
#
function session --description "Get the current Cumulocity session"
    c8y sessions get $argv
end

# -----------
# set-session
# -----------
# Description: Switch Cumulocity session interactively
# Usage:
#   set-session
#
function set-session --description "Switch Cumulocity session interactively"
    set c8yenv ( c8y sessions set --noColor=false $argv )
    if test $status -ne 0
        echo "Set session failed"
        return
    end
    echo "$c8yenv" | source
end

# -----------
# clear-session
# -----------
# Description: Clear all cumulocity session variables
# Usage:
#   clear-session
#
function clear-session --description "Clear all cumulocity session variables"
    c8y sessions clear | source
end


# -----------
# set-c8ymode-xxxx
# -----------
# Description: Set temporary mode by setting the environment variables
# Usage:
#   set-c8ymode-dev     (enable PUT, POST and DELETE)
#   set-c8ymode-qual    (enable PUT, POST)
#   set-c8ymode-prod    (disable PUT, POST and DELETE)
#
function set-c8ymode --description "Enable a c8y temporary mode by setting the environment variables"
    c8y settings update --shell auto mode $argv[1] | source
    echo (set_color green)"Enabled "$argv[1]" mode (temporarily)"(set_color normal)
end

function set-c8ymode-dev --description "Enable dev mode (All enabled)"
    set-c8ymode dev
end

function set-c8ymode-qual --description "Enable qual mode (DELETE disabled)"
    set-c8ymode qual
end

function set-c8ymode-prod --description "Enable prod mode (POST/PUT/DELETE disabled)"
    set-c8ymode prod
end

# ----------
# c8y-update
# ----------
# Description: Update the c8y binary
# The latest binary version will be downloaded from github
#
# Usage:
#   c8y-update
#
function c8y-update --description "Update the c8y binary"
    set VERSION $argv[1]
    set -q VERSION || set VERSION "latest"

    set INSTALL_PATH $argv[2]
    set -q INSTALL_PATH || set INSTALL_PATH "~/bin"

    if test ! -d "$INSTALL_PATH"
        mkdir -p "$INSTALL_PATH"
    end

    set current_version (c8y version 2> /dev/null | tail -1)

    # Get binary name based on os type
    set BINARY_NAME c8y.linux

    switch $OSTYPE
        case 'linux-gnu*'
            set BINARY_NAME c8y.linux

        case 'darwin*'
            set BINARY_NAME c8y.macos
        
        case 'cygwin'
            set BINARY_NAME c8y.windows.exe
        
        case 'msys'
            set BINARY_NAME c8y.windows.exe
        
        case 'linux*'
            set BINARY_NAME c8y.linux

        case '*'
            # assume windows
            set BINARY_NAME c8y.windows.exe
    end

    # try to download latest c8y version
    echo -n "downloading ($BINARY_NAME)..."

    set c8ytmp ./.c8y.tmp
    if test "$VERSION" = "latest"
        curl -L --silent https://github.com/reubenmiller/go-c8y-cli/releases/latest/download/$BINARY_NAME -o $c8ytmp
    else
        curl -L --silent https://github.com/reubenmiller/go-c8y-cli/releases/download/$VERSION/$BINARY_NAME -o $c8ytmp
    end
    chmod +x $c8ytmp

    set new_version ($c8ytmp version 2>/dev/null | tail -1)

    if test "$new_version" = ""
        if test (cat $c8ytmp | head -1 | grep ELF) then
            set_color red
            echo "Failed download latest version: err=Unknown binary error"
            set_color normal
        else
            set_color red
            echo "Failed download latest version: err="(cat .c8y.tmp | head -1)
            set_color normal
        end
        rm -f .c8y.tmp
        return 1
    else
        set_color green
        echo "ok"
        set_color normal
        mv $c8ytmp $INSTALL_PATH/c8y
    end

    if test ! (command -v c8y)
        echo "Adding install path ($INSTALL_PATH) to PATH variable"
        set -x PATH "$PATH:$INSTALL_PATH"
    end

    if test "$current_version" = "$new_version"
        set_color green
        echo "c8y is already up to date: $current_version"
        set_color normal
        return 0
    end
    
    # update completions
    if test (command -v c8y)
        echo -n "updating completions..."
        c8y completion fish | source
        set_color green
        echo "ok"
        set_color normal
    end

    # show new version to user
    c8y version
end
