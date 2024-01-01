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

# -------------
# clear-session
# -------------
# Description: Clear all cumulocity session variables
# Usage:
#   clear-session
#
function clear-session --description "Clear all cumulocity session variables"
    c8y sessions clear | source
end

# -------------------
# clear-c8ypassphrase
# -------------------
# Description: Clear the encryption passphrase environment variables
# Usage:
#   clear-c8ypassphrase
#
function clear-c8ypassphrase --description "Clear the encryption passphrase environment variables"
    set -u C8Y_PASSPHRASE
    set -u C8Y_PASSPHRASE_TEXT
end

# ----------------
# set-c8ymode-xxxx
# ----------------
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
