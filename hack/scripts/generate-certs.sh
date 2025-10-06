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
# [1] Output directory
# [2] Node name
# [3] Node IPs (comma separated)
# [4] Service account name (optional)

OUT_DIR="${1}"
NODE_NAME="${2}"
NODE_IPS_CSV="${3%,[[:space:]]*}"
SERVICE_ACCOUNT="${4:-${NODE_NAME}@${PROJECT_ID}.iam.gserviceaccount.com}"

CERT_BASE_NAME="${NODE_NAME}-client"

# We first need to retrieve the identity-ca root certificate.
# This certificate has to be installed on the node and is required to register
# the client certificate.

if [[ ! -f "${OUT_DIR}/identity-ca.cert" ]]; then
    echo "Retrieving identity-ca certificate..."

    gcloud privateca roots describe 'identity-server-ca' \
        --project="${PROJECT_ID}" \
        --pool="${POOL_NAME}" \
        --location="${LOCATION}" \
        --format 'json' | jq -r '.pemCaCertificates[0]' > "${OUT_DIR}/identity-ca.cert"
else
    echo "Using existing identity-ca certificate..."
fi

# Make sure we have a client.key on disk.

REQUIRE_NEW_CERT='no'
DATE_SUFFIX=$(date +'%Y%m%d')

if [[ -f "${OUT_DIR}/client.key" ]]; then
    echo "Key for ${NODE_NAME} found on disk."
else
    echo "Generating new key for ${NODE_NAME}..."

    openssl ecparam -genkey -name secp384r1 -out "${OUT_DIR}/client-${DATE_SUFFIX}.key"
    ln -sf "${OUT_DIR}/client-${DATE_SUFFIX}.key" "${OUT_DIR}/client.key"

    REQUIRE_NEW_CERT='yes'
fi

# If we found a key, we look for the certificate
# If there is one on disk we assume it is the one fitting our key
# If there is none, we try to download it from GCP. If there's none either,
# we need to generate a new one.

if [[ ${REQUIRE_NEW_CERT} = 'no' ]] && [[ ! -f "${OUT_DIR}/client.cert" ]]; then
    echo "No certificate found on disk. Checking for existing client certificate..."

    CERT_NAME=$(gcloud privateca certificates list \
        --project="${PROJECT_ID}" \
        --issuer-pool="${POOL_NAME}" \
        --location="${LOCATION}" \
        --format='json' |\
        jq -r '.[] | select(.revocationDetails == null) | .name' |\
        grep -E "/${CERT_BASE_NAME}(-[a-f0-9]{8}|)\$" || echo )

    if [[ -n "${CERT_NAME}" ]]; then
        echo "Found. Downloading existing certificate ${CERT_NAME}..."

        gcloud privateca certificates export "${CERT_NAME}" \
            --project="${PROJECT_ID}" \
            --issuer-pool="${POOL_NAME}" \
            --issuer-location="${LOCATION}" \
            --output-file="${OUT_DIR}/client-${DATE_SUFFIX}.cert"

        ln -sf "${OUT_DIR}/client-${DATE_SUFFIX}.cert" "${OUT_DIR}/client.cert"
    else
        echo "Not found."
        REQUIRE_NEW_CERT='yes'
    fi
fi

# Check if ${OUT_DIR}/client.cert is signed by the ${OUT_DIR}/client.key
# If they don't match we create a new CERT with the key we found
if [[ ${REQUIRE_NEW_CERT} = 'no' ]]; then
    CERT_MD5=$(openssl x509 -noout -modulus -in "${OUT_DIR}/client.cert")
    KEY_MD5=$(openssl rsa -noout -modulus -in "${OUT_DIR}/client.key")

    if [[ "${CERT_MD5}" != "${KEY_MD5}" ]]; then
        echo "Certificate is NOT signed by the key. A new certificate is required."
        REQUIRE_NEW_CERT='yes'
    fi
fi

if [[ ${REQUIRE_NEW_CERT} = 'yes' ]]; then
    # Generate a new CSR with the existing key.
    echo "Generating new CSR for ${NODE_NAME}..."
    openssl req -new \
        -nodes \
        -extensions v3_req \
        -key "${OUT_DIR}/client.key" \
        -out "${OUT_DIR}/client-${DATE_SUFFIX}.csr" \
        -subj "/CN=${NODE_NAME}" \
        -addext "subjectAltName=DNS:${NODE_NAME},IP:${NODE_IPS_CSV//,/,IP:},email:${SERVICE_ACCOUNT}" \
        -addext "keyUsage=digitalSignature" \
        -addext "extendedKeyUsage=clientAuth"

    CERT_NAME="${CERT_BASE_NAME}-$(openssl rand -hex 4)"
    echo "Creating client certificate ${CERT_NAME}..."

    gcloud privateca certificates create "${CERT_NAME}" \
        --project="${PROJECT_ID}" \
        --issuer-pool="${POOL_NAME}" \
        --issuer-location="${LOCATION}" \
        --csr="${OUT_DIR}/client-${DATE_SUFFIX}.csr" \
        --validity='P90D' \
        --cert-output-file="${OUT_DIR}/client-${DATE_SUFFIX}.cert"

        ln -sf "${OUT_DIR}/client-${DATE_SUFFIX}.cert" "${OUT_DIR}/client.cert"
fi
