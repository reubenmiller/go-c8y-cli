#!/bin/bash
#
# Upload inventory binaries with c8y api using curl style form fields (--form/-F)
# and check the uploaded contents
#
set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false
export C8Y_SETTINGS_CI=true
export C8Y_SETTINGS_DEFAULTS_CACHE=false

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
        c8y binaries delete --id "$i" || true
    done
    rm -rf "$TEMP_DIR"
    exit "$exit_status"
}

trap cleanup EXIT

# It uploads a binary using a file and custom form fields
test_upload_file_with_form_fields () {
    file1="$TEMP_DIR/form_upload.txt"
    printf "Line 1\nLine 2\n" > "$file1"

    binary_id=$(
        c8y api POST /inventory/binaries \
            -F "file=@$file1" \
            -F 'object={"name":"form_upload_custom.txt","type":"text/plain","c8y_FormUploadTest":{}}' \
            --select id -o csv
    )
    [ -n "$binary_id" ]
    IDS+=("$binary_id")

    # The object form field is used for the managed object
    name=$(c8y inventory get --id "$binary_id" --select name -o csv)
    [ "$name" = "form_upload_custom.txt" ]
    c8y inventory get --id "$binary_id" --select c8y_FormUploadTest -o json | grep -q c8y_FormUploadTest

    # The file contents are uploaded unchanged
    c8y binaries get --id "$binary_id" --outputFileRaw "$TEMP_DIR/downloaded1" > /dev/null
    cmp --silent "$file1" "$TEMP_DIR/downloaded1"
}

# It sends the file with every piped request
test_upload_file_for_each_piped_request () {
    file2="$TEMP_DIR/form_upload_piped.txt"
    printf "Line 3\nLine 4\n" > "$file2"

    piped_ids=()
    while IFS= read -r id; do
        [ -n "$id" ] && piped_ids+=("$id")
    done < <(
        printf "/inventory/binaries\n/inventory/binaries\n" |
            c8y api -F "file=@$file2" -F 'object={"name":"form_upload_piped.txt","type":"text/plain"}' --select id -o csv
    )
    IDS+=("${piped_ids[@]}")
    [ "${#piped_ids[@]}" -eq 2 ]

    for i in "${!piped_ids[@]}"; do
        c8y binaries get --id "${piped_ids[$i]}" --outputFileRaw "$TEMP_DIR/downloaded_piped_$i" > /dev/null
        cmp --silent "$file2" "$TEMP_DIR/downloaded_piped_$i"
    done
}

test_upload_file_with_form_fields
test_upload_file_for_each_piped_request
