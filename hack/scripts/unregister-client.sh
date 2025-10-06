#!/usr/bin/env bash
set -euo pipefail

# All these variables can be set as environment variables.
# If not set, the default values will be used.

PROJECT_ID=${PROJECT_ID:-trv-identity-server-testing}
POOL_NAME=${POOL_NAME:-integration-test-ca-pool}
LOCATION=${LOCATION:-europe-west1}

# This script has to be executed on the ansible host machine.
# The user calling this script has to have access to all required certificates and keys.
# All files will be placed into the directory specified as the first argument.
#
# Arguments:
# [1] Node name

NODE_NAME="${1:?Node name is required}"
CERT_BASE_NAME="${NODE_NAME}-client"

echo "Getting existing client certificate..."

CERT_NAME=$(gcloud privateca certificates list \
    --project="${PROJECT_ID}" \
    --issuer-pool="${POOL_NAME}" \
    --location="${LOCATION}" \
    --format='json' |\
    jq -r '.[] | select(.revocationDetails == null) | .name' |\
    grep -E "/${CERT_BASE_NAME}(-[a-f0-9]{8}|)\$" || echo )

if [[ -n "${CERT_NAME}" ]]; then
    echo "Found. Revoking certificate ${CERT_NAME}..."

    gcloud privateca certificates revoke \
        --certificate "${CERT_NAME}" \
        --project="${PROJECT_ID}" \
        --issuer-pool="${POOL_NAME}" \
        --issuer-location="${LOCATION}"
else
    echo "No client certificate found."
    exit 1
fi
