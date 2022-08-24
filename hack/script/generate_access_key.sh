#!/usr/bin/env bash
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

set -e

# verify required env
: "${V3IO_CONTROLPLANE_URL:?Need to set V3IO_CONTROLPLANE_URL to non-empty value}"
: "${V3IO_CONTROLPLANE_USERNAME:?Need to set V3IO_CONTROLPLANE_USERNAME to non-empty value}"
: "${V3IO_CONTROLPLANE_PASSWORD:?Need to set V3IO_CONTROLPLANE_PASSWORD to non-empty value}"

# requires httpie, jq
echo "Creating control session"
echo '{"data":{"type":"session","attributes":{"plane": "control", "username": "'${V3IO_CONTROLPLANE_USERNAME}'","password": "'${V3IO_CONTROLPLANE_PASSWORD}'"}}}' | http --verify=no ${V3IO_CONTROLPLANE_URL}/api/sessions --session igz_user

echo "Creating Data Access Key"
ACCESS_KEY_ID=$( echo '{"data":{"type":"access_key","attributes":{"plane":"data","label":"v3io-go-test"}}}' | http --verify=no ${V3IO_CONTROLPLANE_URL}/api/self/get_or_create_access_key --session igz_user | jq -r ".data.id")

echo "V3IO_DATAPLANE_ACCESS_KEY=${ACCESS_KEY_ID}"
