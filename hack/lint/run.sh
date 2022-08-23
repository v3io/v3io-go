# Copyright 2019 Iguazio
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
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
