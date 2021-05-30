#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

OUTPUT_DIR=${OUTPUT_DIR:-"$SCRIPT_DIR"}
mkdir -p "$OUTPUT_DIR"
pushd "$OUTPUT_DIR"

echo "Cloning addons repo"
if [[ -d .go-c8y-cli ]]; then
    rm -Rf .go-c8y-cli
fi
git clone --depth 1 https://github.com/reubenmiller/go-c8y-cli-addons.git .go-c8y-cli 2>/dev/null

mkdir -p "bash"
mkdir -p "zsh"
mkdir -p "fish"

echo "Creating completions: $OUTPUT_DIR"
c8y completion bash > "bash/c8y"
c8y completion zsh > "zsh/_c8y"
c8y completion fish > "fish/c8y.fish"

popd
