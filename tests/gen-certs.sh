#!/usr/bin/env bash
set -euo pipefail

GITHUB_WORKSPACE="${GITHUB_WORKSPACE:-$(pwd)}"

CA_ROOT="${CA_ROOT:-${GITHUB_WORKSPACE}/mtls}"
cd "${GITHUB_WORKSPACE}"

mkdir -p "${CA_ROOT}/private"
mkdir -p "${CA_ROOT}/generated"

# This key is used for the custom root CA
openssl genrsa -out "${CA_ROOT}/private/cakey.pem" 4096

openssl req -config "tests/openssl-ca.cnf" \
    -new -x509 -nodes \
    -subj "/CN=localhost" \
    -addext 'subjectAltName=DNS:localhost' \
    -key "${CA_ROOT}/private/cakey.pem" \
    -out "${CA_ROOT}/generated/cacert.pem"

# Generate a ccertificate
generate_cert() {
    identity="${1}"
    email="${2:-${identity}@trv-identity-server-testing.iam.gserviceaccount.com}"

    openssl genrsa -out "${CA_ROOT}/generated/${identity}.key" 4096

    echo "> wrote ${CA_ROOT}/generated/${identity}.key"

    tmp_conf=$(mktemp)
    cp 'tests/openssl-certs.cnf' "${tmp_conf}"
    echo "email = ${email}" >> "${tmp_conf}"

    openssl req -config "${tmp_conf}" \
        -new -nodes \
        -subj "/CN=localhost/emailAddress=${email}" \
        -key "${CA_ROOT}/generated/${identity}.key" \
        -out "${CA_ROOT}/generated/${identity}.csr"

    echo "> wrote ${CA_ROOT}/generated/${identity}.csr"

    openssl x509 -req \
        -CAcreateserial \
        -CA "${CA_ROOT}/generated/cacert.pem" \
        -CAkey "${CA_ROOT}/private/cakey.pem" \
        -in "${CA_ROOT}/generated/${identity}.csr" \
        -out "${CA_ROOT}/generated/${identity}.cert" \
        -extfile "${tmp_conf}" \
        -extensions v3_ext

    echo "> wrote ${CA_ROOT}/generated/${identity}.cert"

    rm "${tmp_conf}"
}

echo
echo "Generate self-signed client certificate"
generate_cert "client-selfsigned"

