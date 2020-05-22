#!/bin/bash

set -eo pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

set -x
go run github.com/go-bindata/go-bindata/go-bindata \
    -nocompress \
    -nometadata \
    -pkg "assets" \
    -prefix "${SCRIPT_ROOT}" \
    -o "${SCRIPT_ROOT}/pkg/controller/assets/bindata.go" \
    ${SCRIPT_ROOT}/manifests/...
