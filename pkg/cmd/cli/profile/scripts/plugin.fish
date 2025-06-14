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
    set c8yenv ( c8y sessions login --noColor=false $argv )
    if test $status -ne 0
        echo "Set session failed"
        return 1
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
