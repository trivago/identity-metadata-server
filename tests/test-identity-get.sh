#!/usr/bin/env bash

CA_ROOT='mtls'
identity='bootstrap'

curl -vvv -f -X GET -H "Content-Type: application/json" \
    --cacert "${CA_ROOT}/cacert.pem" \
    --cert "${CA_ROOT}/generated/${identity}.cert" --key "${CA_ROOT}/generated/${identity}.key" \
    -d '{"audiences": ["test"], "lifetime": "10m"}' \
    'https://localhost:8443/token'
