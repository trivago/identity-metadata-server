# syntax=docker/dockerfile:1

FROM google/cloud-sdk:latest

RUN apt-get update && apt-get -qqy install sudo iptables traceroute net-tools

ENTRYPOINT [ "/usr/bin/sleep", "3600" ]
