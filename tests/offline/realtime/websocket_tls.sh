#!/bin/bash
# The websocket TLS behavior of v2.59.0 against the self-signed fake server
# (C8Y_OFFLINE_TLS=1); without TLS there is nothing to verify, so the test is skipped.
#
# usage: websocket_tls.sh <realtime|notification2>
#
# * realtime: connects with the insecure setting off (the websocket never verified the certificate)
# * notification2: honors the insecure setting. The token request is a normal request, which
#   fails verifying the certificate without the setting, so this test only checks that the
#   websocket connects with the setting on.
set -eou pipefail

case "$C8Y_HOST" in
    https://*) ;;
    *)
        echo "skipped, not a TLS server: exit=0"
        exit 0
        ;;
esac

KIND="$1"
device=$(c8y devices create --name "tls-$KIND-$RANDOM" --select id --output csv)

case "$KIND" in
    realtime)
        # insecure setting off for this command only: the websocket still connects
        C8Y_SETTINGS_DEFAULTS_INSECURE=false c8y alarms subscribe --device "$device" --duration 1s </dev/null 2>&1 | head -5
        echo "exit=${PIPESTATUS[0]}"
        ;;
    notification2)
        c8y notification2 subscriptions create --name "tls-sub-$RANDOM" --context mo --apiFilter managedobjects --device "$device" --select name --output csv >/dev/null
        name=$(c8y notification2 subscriptions list --select subscription --output csv | head -1)
        c8y notification2 subscriptions subscribe --name "$name" --duration 1s </dev/null 2>&1 | head -5
        echo "exit=${PIPESTATUS[0]}"
        ;;
    *)
        echo "unknown kind: $KIND" >&2
        exit 2
        ;;
esac
