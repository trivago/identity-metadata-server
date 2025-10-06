# Example scripts

## create-service-account

Creates a service account for a machine to bind to.

## generate-certs

Generates a client-certificate that is used to identify a machine. 
This requires a workload-identity pool and certificate authority
being set up in Google Cloud.

## unregister-client

Revokes a client-certificate generated through [generate-certs],

## setup-identity-server

Creates the expected directory structure, user and groups for the
identity-server. This structure is also used by the [sytemd example](../systemd/identity-server.service).

## setup-metadata-server

Creates the expected directory structure, user and groups for the
metadata-server. This structure is also used by the [sytemd example](../systemd/metadata-server.service).
