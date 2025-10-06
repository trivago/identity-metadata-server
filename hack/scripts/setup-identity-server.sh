#!/usr/bin/env bash

# Create users and groups
groupadd identity-server
useradd -M -r -s /usr/bin/false -g identity-server identity-server

# Create directories
mkdir -p '/opt/identity-server'
mkdir -p '/opt/identity-server/private'
mkdir -p '/opt/identity-server/public'

# Set permissions
chown -R identity-server:identity-server '/opt/identity-server'

chmod -R 750 '/opt/identity-server'
chmod -R 600 '/opt/identity-server/private'
chmod -R 644 '/opt/identity-server/public'

# Make root certificate available to the system
# cp /opt/metadata-server/public/identity-server-ca.crt /usr/local/share/ca-certificates/identity-server-ca.crt
# update-ca-certificates
