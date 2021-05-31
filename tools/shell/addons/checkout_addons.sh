#!/bin/bash

ADDONS=.go-c8y-cli

if [[ ! -d "$ADDONS" ]]; then
    git clone --depth 1 https://github.com/reubenmiller/go-c8y-cli-addons.git "$ADDONS" 2>/dev/null
else
    git -C "$ADDONS" pull --ff-only
fi
