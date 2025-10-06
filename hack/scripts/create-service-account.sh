#!/usr/bin/env bash
set -euo pipefail

PROJECT_ID=${PROJECT_ID:-trv-identity-server-testing}
WORKLOAD_IDENTITY_POOL=${WORKLOAD_IDENTITY_POOL:-integration-test}
WORKLOAD_IDENTITY_PROJECT_ID=${WORKLOAD_IDENTITY_PROJECT_ID:-identity-server}
LOCATION=${LOCATION:-europe-west1}

# Arguments:
# [1] Node name
# [2] Service account name (optional)

NODE_NAME="${1}"
SERVICE_ACCOUNT_NAME="${2:-${NODE_NAME}}"

# The service account name can either be an email address or just the name.
# If it's just a name we need to append the project ID.
# If it's an email address we need to extract the project ID.

if [[ "${SERVICE_ACCOUNT_NAME}" == *@* ]]; then
    SERVICE_ACCOUNT="${SERVICE_ACCOUNT_NAME}"
    SERVICE_ACCOUNT_NAME="${SERVICE_ACCOUNT_NAME%%@*}"
    PROJECT_ID=$(echo "${SERVICE_ACCOUNT}" | awk -F'[@.]' '{print $2}')
else
    SERVICE_ACCOUNT="${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
fi

# We need to create the service account if it does not exist yet.

if ! gcloud iam service-accounts describe --project="${PROJECT_ID}" "${SERVICE_ACCOUNT}" &> /dev/null; then
    echo "Creating Service Account: ${SERVICE_ACCOUNT}"
    gcloud iam service-accounts create "${SERVICE_ACCOUNT_NAME}" \
        --project="${PROJECT_ID}" \
        --description="Service account generated for ${NODE_NAME}"
else
    echo "Service Account already exists: ${SERVICE_ACCOUNT}"
fi

# Bind the service account to the workload identity pool.

WORKLOAD_IDENTITY_PROJECT_NUMBER=$(gcloud projects describe "${WORKLOAD_IDENTITY_PROJECT_ID}" --format 'value(projectNumber)')
PRINCIPAL="principal://iam.googleapis.com/projects/${WORKLOAD_IDENTITY_PROJECT_NUMBER}/locations/global/workloadIdentityPools/${WORKLOAD_IDENTITY_POOL}/subject/${NODE_NAME}"

if ! gcloud iam service-accounts get-iam-policy "${SERVICE_ACCOUNT}" | grep "${PRINCIPAL}" &> /dev/null; then
    echo "Creating IAM policy binding for ${SERVICE_ACCOUNT}"
    gcloud iam service-accounts add-iam-policy-binding "${SERVICE_ACCOUNT}" \
        --project="${PROJECT_ID}" \
        --role='roles/iam.workloadIdentityUser' \
        --member="${PRINCIPAL}"
else
    echo "IAM policy binding already exists for ${SERVICE_ACCOUNT}"
fi
