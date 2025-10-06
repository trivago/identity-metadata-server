#!/bin/bash

# The following environment variables can be set:
#
# - PORT: The port to use for the metadata server (default is 80).
# - METADATA_SERVER: The hostname of the metadata server to resolve.
# - METADATA_HOST_IP: The exact IP address to use for the metadata server.
# - HOST_IP_MATCH: A regex to match the host IP during auto-detection
#
# If `METADATA_SERVER` is set, it will resolve the hostname to an IP address.
# If `METADATA_HOST_IP` is set, it will use that IP address directly.
# The `METADATA_SERVER` variable takes precedence over `METADATA_HOST_IP`.
#
# If none of these are set, the script will default to use auto-detection.
# If a bond0 interface exists with a "10.x.x.x" IP, it will be used.
# Otherwise, it will fall back to the first IP address found on the host
# that matches "10.x.x.x" (sorted).
# If HOST_IP_MATCH is set, it will use that regex for matching the IP.
#
# Please note that 10.x.x.x is chosen as a common IP range of an internal
# network. Your network might use a different range.
# Bond0 is an interface that is commonly used for bonding multiple network
# interfaces and might not apply to every use case.

if [[ ! -z ${METADATA_SERVER:-} ]]; then
    echo "Resolving ${METADATA_SERVER}"
    METADATA_HOST_IP="$(host "${METADATA_SERVER}" | awk '/has address/ { print $4 ; exit }')"

elif [[ ! -z ${METADATA_HOST_IP:-} ]]; then
    echo "Using provided Metadata-server IP"

else
    echo "Metadata-server IP is auto-detected"

    # Allow `HOST_IP_MATCH` to be set from the environment, otherwise default to
    # a regex that matches 10.x.x.x IPs.
    IP_MATCH='10\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}'
    HOST_IP_MATCH="${HOST_IP_MATCH:-$IP_MATCH}"

    # If we have a bond0 interface, we use its IP as the host IP.
    BOND_IP="$(ip address show bond0 2> /dev/null | grep -Eo "inet ${IP_MATCH}" | awk '{print $2}')"

    # We're using the "smallest" IP as the fallback host IP
    FIRST_IP="$(hostname -I | grep -Eo "${HOST_IP_MATCH}" | sort -n | head -n 1)"

    # The host IP is either the bond0 IP or the first IP found
    METADATA_HOST_IP="${BOND_IP:-${FIRST_IP}}"
fi

echo "Using Metadata-server IP: ${METADATA_HOST_IP}"

echo "Removing old rules for metadata.google.internal:"
iptables -t nat -S PREROUTING | grep '169.254.169.254'

# Make sure the index is _reversed_ otherwise the deletion will fail after the first one
iptables -t nat -L PREROUTING --line-numbers \
  | grep '169.254.169.254' \
  | sort -rn \
  | awk '{print $1}' \
  | xargs -r -I% iptables -t nat -D PREROUTING %

echo "Adding new rule for metadata.google.internal"
iptables -w -t nat -I PREROUTING 1 -m comment -m tcp \
  --comment "metadata.google.internal" \
  -d 169.254.169.254 -p tcp --dport 80 \
  -j DNAT --to-destination "${METADATA_HOST_IP}:${PORT:-80}"

echo "iptables now"
iptables -t nat -S
