#!/bin/bash

echo "Installing tooling"
command -v yq || go install github.com/mikefarah/yq/v4@latest
command -v goimports || go install golang.org/x/tools/cmd/goimports@latest
command -v task || go install github.com/go-task/task/v3/cmd/task@latest
command -v goreleaser || go install github.com/goreleaser/goreleaser@latest
