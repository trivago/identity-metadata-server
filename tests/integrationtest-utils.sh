#!/usr/bin/env bash


# Required. Path to output of gen-certs.sh
# CA_ROOT='mtls'
# Optional. Identity to use for client registration.
# CURL_IDENTITY='bootstrap'
# Optional. JSON content to send with request.
# CURL_CONTENT_JSON='{"identity": "block", "addresses": ["127.0.0.1"], "certificate": "-"}'
#
# Example:
# curl_expect 403 GET https://localhost:8443/client
#
# Expect 403 to be returned from doing a GET request to the given endpoint

curl_expect() {
    expected_status="${1}"
    method="${2}"
    url="${3}"

    args=('-sL' '-vvv' '-w' '%{http_code}\n' '-X' "${method}" "--resolve" "identity-server:8443:127.0.0.1")

    # If URL starts with https, pass the CA_ROOT.
    # If it's not https, we don't need to pass an identity either.
    if [[ "${url}" == https* ]]; then
        if [[ ! -f "${CA_ROOT}/cacert.pem" ]]; then
            echo "Error: Certificate file not found at ${CA_ROOT}/cacert.pem"
            exit 1
        fi

        args+=('--cacert' "${CA_ROOT}/cacert.pem")

        # if CURL_IDENTITY is set, add client cert and key
        if [[ -n "${CURL_IDENTITY:-}" ]]; then
            if [[ ! -f "${CA_ROOT}/client/${CURL_IDENTITY}.cert" ]]; then
                echo "Error: Client certificate file not found at ${CA_ROOT}/client/${CURL_IDENTITY}.cert"
                exit 1
            fi

            if [[ ! -f "${CA_ROOT}/client/${CURL_IDENTITY}.key" ]]; then
                echo "Error: Client private key file not found at ${CA_ROOT}/client/${CURL_IDENTITY}.key"
                exit 1
            fi

            args+=('--cert' "${CA_ROOT}/client/${CURL_IDENTITY}.cert" '--key' "${CA_ROOT}/client/${CURL_IDENTITY}.key")
        fi
    fi

    # If CURL_CONTENT_JSON is set, add it to the request alongside the Content-Type header
    if [[ -n "${CURL_CONTENT_JSON:-}" ]]; then
        args+=('-H' 'Content-Type: application/json' '-d' "${CURL_CONTENT_JSON}")
    elif [[ -n "${CURL_CONTENT_RAW:-}" ]]; then
        CURL_CONTENT_TYPE=${CURL_CONTENT_TYPE:-'application/octet-stream'}
        args+=('-H' "Content-Type: ${CURL_CONTENT_TYPE}" '-d' "${CURL_CONTENT_RAW}")
    fi

    # Stderr will be visible, stdout will be stored in the variable
    result=$( (curl "${args[@]}" "${url}" || echo $?) | tail -1)

    if [[ ${expected_status} == 'x'* ]]; then
        if [[ ${result} != "${expected_status:1}" ]]; then
            echo "Error: Expected ${expected_status:1}, got ${result}"
            exit 1
        fi
    elif [[ "${result}" != *"${expected_status}"* ]]; then
        echo "Error: Expected ${expected_status}, got ${result}"
        exit 1
    fi

    echo "Success: ${result}"
}
