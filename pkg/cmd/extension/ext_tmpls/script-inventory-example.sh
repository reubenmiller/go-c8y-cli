#!/usr/bin/env bash
set -e

# Describe how to use your command
help() {
  cat <<EOF
Example description about what your command does

Usage:
    c8y %[1]s %[2]s [FLAGS]

$(examples)

Flags:
    --name <string>           Match by name
    --label <string>          Match by c8y_Kpi.label
    --onlyAgents              Only include managed objects with the 'com_cumulocity_model_Agent' fragment
EOF
}

examples () {
  cat <<EOF
Examples:
    c8y %[1]s %[2]s
    # List items

    c8y %[1]s %[2]s --pageSize 100
    # List the first 100 items

    c8y %[1]s %[2]s --name "MyLinuxDevice" --onlyAgents
    # List items matching a given name and are agents
EOF
}

# Print log messages on stderr so it does not mix with results which is generally printed on stdout
echo "Running custom %[2]s command" >&2

# Inventory query builder using
FLAGS=()
QUERY_PARTS=()
POSITIONAL_ARGS=()

function join_by {
    local d=${1-} f=${2-}
    if shift 2; then
        printf %%s "$f" "${@/#/$d}"
    fi
}

# # Parse options for flags with values: --flag <value>, or boolean/switch flags: --help|-h
while [ $# -gt 0 ]; do
    case "$1" in
        --name)
            QUERY_PARTS+=(
                "name eq '$2'"
            )
            shift
            ;;
        --label)
            QUERY_PARTS+=(
                "c8y_Kpi.label eq '$2'"
            )
            shift
            ;;
        --onlyAgents)
            QUERY_PARTS+=(
                "has(com_cumulocity_model_Agent)"
            )
            ;;
        # Support showing the help when users provide '-h' or '--help'
        -h|--help)
            help
            exit 0
            ;;
        # Support showing just the examples using '--examples'
        --examples)
            examples
            exit 0
            ;;
        *)
            POSITIONAL_ARGS+=("$1")
            ;;
    esac
    shift
done

# Restore additional arguments which can then be referenced via "$@" and "$1" etc.
set -- "${POSITIONAL_ARGS[@]}"

if [ "${#QUERY_PARTS[@]}" -gt 0 ]; then
    FLAGS+=(
        --query
        "$(join_by " and " "${QUERY_PARTS[@]}")"
    )
fi

# Call another c8y command which actually does the heavy lifting
c8y inventory find "${FLAGS[@]}" "$@"
