#!/usr/bin/env sh

set -e
set -u

REPO_ROOT=$(dirname "$0")/../..
echo $REPO_ROOT

for mod in $(go run scripts/go-workspace/main.go list-submodules --path "${REPO_ROOT}/go.work"); do
    cd "${mod}"
    echo "Running go mod tidy in ${mod}"
    go mod tidy || true
    cd - > /dev/null
done

cd "${REPO_ROOT}"
echo "running go mod download"
go mod download

echo "running go work sync"
go work sync
