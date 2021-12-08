#!/usr/bin/env bash
set -e

echo Installing linters...

OS_NAME=$(uname)
FORCE_INSTALL=false
BIN_DIR=$(pwd)/.bin/

echo "Creating bin directory: ${BIN_DIR}"
mkdir -p ${BIN_DIR}

if [[ ! -f ${BIN_DIR}/impi ]] ; then
    echo "impi binary does not exist. Fetching and installing..."
    curl -sSfL https://api.github.com/repos/pavius/impi/releases/latest \
      | grep -i "browser_download_url.*impi.*${OS_NAME}" \
      | cut -d : -f 2,3 \
      | tr -d '"' \
      | tr -d '[:space:]' \
      | xargs curl -sSL --output ${BIN_DIR}/impi
    chmod +x ${BIN_DIR}/impi
    echo "impi installed in: ${BIN_DIR}/impi"
fi

if [[ $# -ne 0 && "$1" == "force" ]]
  then
    echo "Force install golangci-lint requested"
    FORCE_INSTALL=true
fi

if [[ $FORCE_INSTALL = true || ! -f ${BIN_DIR}/golangci-lint ]] ; then
    echo "golangci-lint binary does not exist or force install requested. Fetching and installing..."

    # upgrade from v1.24.0 when https://github.com/golangci/golangci-lint/issues/974 is resolved
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${BIN_DIR} v1.24.0
    echo "golangci-lint installed in: ${BIN_DIR}/golangci-lint"
fi

