#!/bin/sh

# Exit immediately if a command exits with a non-zero status
# Treat unset variables as an error
# Make pipelines return non-zero status if any command fails

REPO_ROOT=$(dirname "$0")/../..

cd "${REPO_ROOT}"
echo "running go work sync"
go work sync
cd - > /dev/null

for mod in $(go run scripts/go-workspace/main.go list-submodules --path "${REPO_ROOT}/go.work"); do
    cd "${mod}"
    echo "Running go mod tidy in ${mod}"
    go mod tidy || true
    cd - > /dev/null
done

cd "${REPO_ROOT}"
echo "running go mod download"
go mod download
