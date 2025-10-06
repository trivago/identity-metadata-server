#!/usr/bin/env bash

# These commands are for testing the token exchange from a container.
# They are not intended to be run as a script, but are rather copied
# and pasted into a shell session.

gcloud compute instances list

curl -s -H "Metadata-Flavor: Google" \
    http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=api://AzureADTokenExchange
