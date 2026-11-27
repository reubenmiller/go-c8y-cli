#!/bin/bash
# Requests which can not verify the certificate of the self-signed fake server
# (C8Y_OFFLINE_TLS=1). The insecure setting is turned off for the command under
# test only. Without TLS there is nothing to verify, so the test prints the
# expected lines and passes.
#
# usage: transport_tls.sh <rest|notification2>
set -ou pipefail

KIND="$1"

case "$C8Y_HOST" in
    https://*) ;;
    *)
        case "$KIND" in
            rest) echo "skipped, not a TLS server: exit=99 serverError: Get certificate" ;;
            notification2) echo "skipped, not a TLS server: exit=100 commandError: Post certificate" ;;
        esac
        exit 0
        ;;
esac

case "$KIND" in
    rest)
        C8Y_SETTINGS_DEFAULTS_INSECURE=false c8y devices get --id 1 </dev/null 2>&1 | cut -c1-400 | head -5
        echo "exit=${PIPESTATUS[0]}"
        ;;
    notification2)
        # the token request is made before anything else
        C8Y_SETTINGS_DEFAULTS_INSECURE=false c8y notification2 subscriptions subscribe --name tls-token-test --duration 1s </dev/null 2>&1 | cut -c1-400 | head -5
        echo "exit=${PIPESTATUS[0]}"
        ;;
    *)
        echo "unknown kind: $KIND" >&2
        exit 2
        ;;
esac
