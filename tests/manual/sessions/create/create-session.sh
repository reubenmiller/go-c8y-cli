#!/usr/bin/env bash
set -e
if [ -n "$CI" ]; then
    set -x
fi

# store the current session variables, before clearing the session (to simulate creating a new session)
BACKUP_C8Y_HOST="$C8Y_HOST"
BACKUP_C8Y_TENANT="$C8Y_TENANT"
BACKUP_C8Y_USER="$C8Y_USER"
BACKUP_C8Y_PASSWORD="$C8Y_PASSWORD"

if [ -z "$C8Y_PASSWORD" ]; then
    echo "This test requires the 'C8Y_PASSWORD' env variable to be set!"
    exit 1
fi

echo "Clearing existing session" >&2
eval "$(c8y sessions clear --shell bash)" ||:

TMPDIR=$(mktemp -d)
cleanup () {
    rm -rf "$TMPDIR"
}
trap cleanup EXIT
export C8Y_SESSION_HOME="$TMPDIR"

#
# Test case 1: Create a session and reference it via the --session global flag
#
c8y sessions create \
    --mode dev \
    --username "$BACKUP_C8Y_USER" \
    --host "$BACKUP_C8Y_HOST" \
    --tenant "$BACKUP_C8Y_TENANT" \
    --password "$BACKUP_C8Y_PASSWORD" \
    --name subtenant
c8y devices list -p 1 -n --session "$TMPDIR/subtenant.json"

#
# Test case 2: Create a session without storing the password and use a combination
# of --session and --sessionPassword
#
c8y sessions create \
    --mode dev \
    --username "$BACKUP_C8Y_USER" \
    --host "$BACKUP_C8Y_HOST" \
    --tenant "$BACKUP_C8Y_TENANT" \
    --password "$BACKUP_C8Y_PASSWORD" \
    --name subtenant \
    --noStorage

echo "Checking session credentials using custom session password" >&2
c8y devices list -p 1 -n --session "$TMPDIR/subtenant.json" --sessionPassword "$BACKUP_C8Y_PASSWORD"

echo "Checking resolution of session using just the name" >&2
c8y devices list -p 1 -n --session "subtenant.json" --sessionPassword "$BACKUP_C8Y_PASSWORD"
