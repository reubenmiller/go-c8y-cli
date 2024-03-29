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
EOF
}

# Print log messages on stderr so it does not mix with results which is generally printed on stdout
echo "Running custom %[2]s command" >&2

# Snippets to help get started:

# Get the script's directory (useful if you need to reference some other assets provided by the extension using a relative path)
# SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# cat "$SCRIPT_DIR/../templates/mytemplate.jsonnet"

# Check for minimal positional arguments
# if [ $# -lt 1 ]; then
#     echo "Missing positional arguments. This command requires at least 1 argument"
#     help
#     exit 1
# fi

# Determine if an executable is in the PATH
# if ! type -p python3 >/dev/null; then
#     echo "python3 not found on the system" >&2
#     exit 1
# fi

# Pass arguments through to another command
# c8y inventory find "$@"
#

# Using the c8y api command to send to a custom endpoint
# TEMPLATE='
# {
#     someNestedFragment: {
#         info: "value",
#     },
# }
# '
# exec c8y api POST "/service/my-service/endpoint" --template="${TEMPLATE}"

# Inventory query builder using
#
# FLAGS=()
# QUERY_PARTS=()

# function join_by {
#     local d=${1-} f=${2-}
#     if shift 2; then
#         printf %%s "$f" "${@/#/$d}"
#     fi
# }

# # Parse options for flags with values: --flag <value>, or boolean/switch flags: --help|-h
# while [ $# -gt 0 ]; do
#     case "$1" in
#         --name)
#             QUERY_PARTS+=(
#                 "name eq '$2'"
#             )
#             shift
#             ;;
#         --onlyAgents)
#             QUERY_PARTS+=(
#                 "has(com_cumulocity_model_Agent)"
#             )
#             ;;
#        Support showing the help when users provide '-h' or '--help'
#        -h|--help)
#            help
#            exit 0
#            ;;
#        # Support showing just the examples using '--examples'
#        --examples)
#            examples
#            exit 0
#            ;;
#         *)
#             REST_ARGS+=("$1")
#             ;;
#     esac
#     shift
# done

# # Reset additional arguments which can then be referenced via "$@"
# set -- "${REST_ARGS[@]}"

# if [ "${#QUERY_PARTS[@]}" -gt 0 ]; then
#     FLAGS+=(
#         --query
#         "$(join_by " and " "${QUERY_PARTS[@]}")"
#     )
# fi

# c8y inventory find "${FLAGS[@]}" "$@"
