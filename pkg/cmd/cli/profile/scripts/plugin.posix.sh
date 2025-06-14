#!/bin/bash
# Force encoding
export LANG=C.UTF-8
export LC_ALL=C.UTF-8

# Note: Bash v3, and some older posix shells don't support function names with a hyphen
# so underscores need to be used, then define an alias which can use hyphens


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
alias session='c8y sessions get'

# -----------
# set-session
# -----------
# Description: Switch Cumulocity session interactively
# Usage:
#   set-session
#
set_session() {
    c8yenv=$( c8y sessions login --noColor=false $@ )
    code=$?
    if [ $code -ne 0 ]; then
        echo "Set session failed"
        return 1
    fi
    eval "$c8yenv"
}
alias set-session=set_session

# -------------
# clear-session
# -------------
# Description: Clear all cumulocity session variables
# Usage:
#   clear-session
#
clear_session() {
    c8yenv=$(c8y sessions clear)
    eval "$c8yenv"
}
alias clear-session=clear_session
