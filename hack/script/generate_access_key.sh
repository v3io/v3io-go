#!/usr/bin/env bash

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
