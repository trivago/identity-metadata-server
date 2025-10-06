#!/usr/bin/env bash

CA_ROOT='mtls'
cert=$(cat ${CA_ROOT}/generated/bootstrap.cert | base64 -w0)
identity='bootstrap'

curl -vvv -f -X POST -H "Content-Type: application/json" \
    --cacert "${CA_ROOT}/cacert.pem" \
    --cert "${CA_ROOT}/generated/${identity}.cert" --key "${CA_ROOT}/generated/${identity}.key"  \
    -d "{\"identity\": \"test\", \"addresses\": [\"127.0.0.1\", \"::1\"], \"certificate\": \"${cert}\"}" \
    'https://127.0.0.1:8443/identity'
