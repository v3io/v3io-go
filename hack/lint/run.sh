#!/usr/bin/env bash
set -e

# NOTE: RUN THAT from repo root

if [[ -z "${BIN_DIR}" ]]; then
  BIN_DIR=$(pwd)/.bin
fi

echo Verifying imports...


${BIN_DIR}/impi \
  --local github.com/v3io/v3io-go \
  --scheme stdLocalThirdParty \
  --skip pkg/dataplane/schemas/ \
  ./...

echo "Linting @"$(pwd)"..."
${BIN_DIR}/golangci-lint run -v
echo Done.
