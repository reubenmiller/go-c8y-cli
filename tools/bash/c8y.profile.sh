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
source /usr/share/bash-completion/bash_completion
source ~/.bash_completion.d/complete_alias

source <(c8y completion bash)

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
    args=""
    if [ $# -gt 0 ]; then
        args="--sessionFilter \"$@\""
    fi
    echo "c8y sessions list $args"
    export C8Y_SESSION=$( c8y sessions list $args)
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
