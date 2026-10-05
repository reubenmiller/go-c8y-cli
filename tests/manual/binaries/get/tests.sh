#!/bin/bash
#
# Regression test for https://github.com/reubenmiller/go-c8y-cli/issues/711
# Downloading multiple piped binaries panicked with
# "*mpb.Progress instance can't be reused after it's done!"
#
set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

createdir () {
    # Cross-platform compatible
    local name="${1:-"c8y-temp"}"
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t "$name")
    echo "$tmpdir"
}

export TEMP_DIR=$(createdir)

IDS=()

cleanup () {
    exit_status=$?
    for i in "${IDS[@]}"; do
        c8y binaries delete --id "$i" --force < /dev/null || true
    done
    rm -rf "$TEMP_DIR"
    exit "$exit_status"
}
trap cleanup EXIT

create_inventory_binary () {
    local name="$1"
    local binary_id=
    # Note: Keep the content small, as the server compresses larger responses (removing the Content-Length header),
    # in which case no download progress bar is shown
    printf 'issue711' > "$TEMP_DIR/$name"
    binary_id=$(c8y binaries create --file "$TEMP_DIR/$name" --select id -o csv < /dev/null)
    IDS+=("$binary_id")
    echo "$binary_id"
}

test01_download_multiple_binaries_with_progress () {
    create_inventory_binary "ci_issue711_1.txt" > /dev/null
    create_inventory_binary "ci_issue711_2.txt" > /dev/null
    create_inventory_binary "ci_issue711_3.txt" > /dev/null

    # Download progress bars are only shown if stderr is a terminal, so force it
    printf '%s\n' "${IDS[@]}" \
    | C8Y_FORCE_STDERR_TTY=true c8y binaries get --outputFileRaw "$TEMP_DIR/output/{filename}" > /dev/null 2> "$TEMP_DIR/stderr.log" || {
        cat "$TEMP_DIR/stderr.log"
        exit 1
    }

    if grep -q "panic:" "$TEMP_DIR/stderr.log"; then
        cat "$TEMP_DIR/stderr.log"
        exit 1
    fi

    test -f "$TEMP_DIR/output/ci_issue711_1.txt"
    test -f "$TEMP_DIR/output/ci_issue711_2.txt"
    test -f "$TEMP_DIR/output/ci_issue711_3.txt"
}

test01_download_multiple_binaries_with_progress
