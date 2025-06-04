#!/usr/bin/env bash
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$SCRIPT_DIR"

FAIL_COUNT=0

# Create dummy input file used in some tests
INPUT_FILE=./input.txt
c8y inventory list -p 3 --select id -o csv > "$INPUT_FILE"

fail() {
	printf '\n\033[31m FAIL %s \033[0m\n' "$1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

pass() {
	printf '\n\033[32m PASS %s \033[0m\n' "$1"
}

echo "
------------------------------------
# Normal command
------------------------------------
"
if c8y devices list -p 1 --name asdfasdfasdf --select id -o csv; then
    pass
else
    fail
fi


echo "
------------------------------------
# Empty does not return any characters
------------------------------------
"
if [ -z "$(c8y devices list -p 1 --name asdfasdfasdf --select id -o csv)" ]; then
    pass
else
    fail
fi


echo "
------------------------------------
# Pipe to another command
------------------------------------
"
if c8y devices list -p 1 --select id -o csv | wc -c | awk '{print $1}' ; then
    pass
else
    fail
fi

echo "
------------------------------------
# Cat to while loop (requires -n)
------------------------------------
"
# shellcheck disable=SC2002
OUTPUT=$(
    cat "$INPUT_FILE" | while read -r line; do
        current_id="$line"
        c8y devices get -n --id "$current_id" --select id --output csv
    done
)

if [ "$OUTPUT" = "$(cat "$INPUT_FILE" )" ]; then
    pass
else
    fail
fi

echo "
------------------------------------
# While loop
------------------------------------
"
COUNT=0

while read -r line
do
    {
        COUNT=$((COUNT + 1))
        current_id="$line"
        c8y devices get --id "$current_id" --select id,name,type,lastUpdated --output csv >/dev/null
        c8y devices get --id "$current_id" --select creationTime -o csv >/dev/null 2>&1
    } </dev/null
done <"$INPUT_FILE"

if [ "$COUNT" -eq 3 ]; then
    pass "(count=$COUNT)"
else
    fail "(got=$COUNT, want=3)"
fi


echo "
------------------------------------
# Empty pipe
------------------------------------
"
if c8y devices list --name "somethingThatDoesNotExist123456789" | c8y devices get; then
    pass
else
    fail
fi


echo "
------------------------------------
# normal pipe
------------------------------------
"
LINE_COUNT=$(c8y devices list -p 1 | c8y devices get | wc -l | awk '{print $1}')
if [ "$LINE_COUNT" -eq 1 ]; then
    pass
else
    fail
fi

echo "
------------------------------------
# FIFO (github) with newline on fifo
------------------------------------
"
run_command() {
    result="$1"
    c8y devices list -p 1 > "$result"
}

RESULT_FILE=./fifo_result
rm -f dummy_fifo
mkfifo dummy_fifo
run_command "$RESULT_FILE" <dummy_fifo &
echo "" > dummy_fifo
sleep 1
LINE_COUNT=$(wc -l "$RESULT_FILE" | awk '{print $1}')
if [ "$LINE_COUNT" -eq 1 ]; then
    pass
else
    fail
fi
rm -f dummy_fifo
rm -f "$RESULT_FILE"

echo "
------------------------------------
# FIFO (github) with unterminated newline on fifo
------------------------------------
"
run_command() {
    result="$1"
    c8y devices list -p 1 > "$result"
}

RESULT_FILE=./fifo_result
rm -f dummy_fifo
mkfifo dummy_fifo
run_command "$RESULT_FILE" <dummy_fifo &
echo -n "" > dummy_fifo
sleep 1
LINE_COUNT=$(wc -l "$RESULT_FILE" | awk '{print $1}')
if [ "$LINE_COUNT" -eq 1 ]; then
    pass
else
    fail
fi
rm -f dummy_fifo
rm -f "$RESULT_FILE"
echo

popd >/dev/null ||:

if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
fi
