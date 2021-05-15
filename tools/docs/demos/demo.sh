#!/bin/bash

#
# Replace host information with generic example information
# i.e. tenant id, host url etc.
#
sub_data () {
    local file="$1"
    local tenant=$( c8y sessions get --select tenant -o csv )
    local host=$( c8y sessions get --select host -o csv )
    local domain=$( c8y sessions get --select host -o csv | sed -E 's/[^.]+\.(.+)/\1/g' )

    sed -i "s/$tenant/t12345/g" "$file"
    sed -i "s/$host/example.cumulocity.com/g" "$file"
    sed -i "s/$domain/cumulocity.com/g" "$file"
}

#
# Start an automated demo
#
start_automated_demo () {
    local demo_script="$(readlink -f $1)"
    shift
    local options="$@"
    pushd ~
    export C8Y_JSONNET_DEBUG=false
    export C8Y_HOME=/workspaces/go-c8y-cli/.cumulocity/sessions
    export C8Y_SESSION_HOME=/workspaces/go-c8y-cli/.cumulocity/sessions
    export C8Y_SETTINGS_DEFAULTS_VIEW=auto
    export C8Y_SETTINGS_LOGGER_HIDESENSITIVE=true
    export PATH="/usr/games/:$PATH"
    local name=$(basename "$demo_script")
    local output_file="/workspaces/go-c8y-cli/tools/docs/demos/${name}.asc"
    local output_folder=$(dirname "$output_file")
    mkdir -p "$output_folder"
    clear
    asciinema rec "$output_file" --overwrite --command "zsh -c \"$demo_script\""
    popd
    echo ""
    echo "Saved demo to: $output_file"

    if [[ -f "$output_file" ]]; then
        sub_data "$output_file"
    fi

    if [[ -f "$output_file" && "$options" =~ "--svg" ]]; then
        echo "Converting demo to svg: $output_file.svg"
        cat "$output_file" | svg-term --out "${output_file}.svg" --window --term iterm2 --profile  "Afterglow.itermcolors"
    fi
    echo ""
}

start_automated_demo $@
