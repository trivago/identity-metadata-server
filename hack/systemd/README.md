# Systemd unit file examples

> [!WARNING]
> The files provided here are examples and need to be adjusted
> Please refer to the documentation of each tool to set the correct values
> for each command line parameter.

## identity-server

Systemd unit to run the identity-server.

## metadata-server

Systemd unit to run the a host-mode metadata-server.

## iptables

Systemd unit to configure iptables for a host-mode metadata-server.  
Please note that this unit expects the iptables rule to _not_ be persisted.

In case your iptables rules keep  piling up, use the script provided in
the [docker version](../../cmd/metadata-iptables-init/) of this step.
