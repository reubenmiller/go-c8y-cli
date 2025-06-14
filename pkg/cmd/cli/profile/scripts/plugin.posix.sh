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
session() {
    c8y sessions get "$@"
}

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

# ----------------
# set-c8ymode-xxxx
# ----------------
# Description: Set temporary mode by setting the environment variables
# Usage:
#   set-c8ymode-dev     (enable PUT, POST and DELETE)
#   set-c8ymode-qual    (enable PUT, POST)
#   set-c8ymode-prod    (disable PUT, POST and DELETE)
#
set_c8ymode() {
    eval "$(c8y settings update --shell auto mode "$1")"
    printf "\e[32mEnabled %s mode (temporarily)\e[0m\n" "$1";
}
alias set-c8ymode=set_c8ymode
alias set-c8ymode-dev='set_c8ymode dev'
alias set-c8ymode-qual='set_c8ymode qual'
alias set-c8ymode-prod='set_c8ymode prod'
