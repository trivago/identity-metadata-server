# syntax=docker/dockerfile:1

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -qqy iptables gawk bind9-host

COPY cmd/metadata-iptables-init/iptables.sh /usr/local/bin/metadata-iptables-init.sh

ENTRYPOINT ["/usr/local/bin/metadata-iptables-init.sh"]
