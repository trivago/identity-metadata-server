#!/usr/bin/env bash

# Create users and groups
groupadd metadata-server

usermod -a -G metadata-server metadata-server

# Create directories
mkdir -p '/opt/metadata-server'
mkdir -p '/opt/metadata-server/private'
mkdir -p '/opt/metadata-server/public'

# Set permissions
chown -R metadata-server:metadata-server '/opt/metadata-server'

chmod -R 750 '/opt/metadata-server'
chmod -R 600 '/opt/metadata-server/private'
chmod -R 644 '/opt/metadata-server/public'

# Make root certificate available to the system
# cp /opt/metadata-server/public/identity-server-ca.crt /usr/local/share/ca-certificates/identity-server-ca.crt
# update-ca-certificates
