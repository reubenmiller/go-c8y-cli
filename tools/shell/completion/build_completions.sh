#!/bin/bash
set -euo pipefail

ROOT_DIR="$( pwd )"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

OUTPUT_DIR=${OUTPUT_DIR:-"$SCRIPT_DIR"}
mkdir -p "$OUTPUT_DIR"
pushd "$OUTPUT_DIR"

echo "Cloning addons repo"
if [[ ! -d .go-c8y-cli ]]; then
    git clone --depth 1 https://github.com/reubenmiller/go-c8y-cli-addons.git .go-c8y-cli 2>/dev/null
else
    git -C .go-c8y-cli pull --ff-only
fi

mkdir -p "bash"
mkdir -p "zsh"
mkdir -p "fish"

# Use go source code as the binary may not be built yet (i.e. in CI/CD)
echo "Creating completions: $OUTPUT_DIR"
go run $ROOT_DIR/cmd/c8y/main.go completion bash > "bash/c8y"
go run $ROOT_DIR/cmd/c8y/main.go completion zsh > "zsh/_c8y"
go run $ROOT_DIR/cmd/c8y/main.go completion fish > "fish/c8y.fish"

popd
