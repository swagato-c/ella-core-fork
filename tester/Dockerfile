# SPDX-FileCopyrightText: Ella Networks Inc.
# SPDX-License-Identifier: BUSL-1.1
#
# The ella-core-tester image without rockcraft, for running Ella's own
# scenarios (integration/) against pinned Ella Core on a plain Docker host:
#
#   docker build -f tester/Dockerfile -t ella-core-tester:latest .
#
# It carries the same binary and packages as rockcraft-tester.yaml.
FROM golang:1.27 AS build
RUN apt-get update && apt-get install -y --no-install-recommends libpcap-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
COPY sctp/ sctp/
COPY ngap/ ngap/
COPY nrppa/ nrppa/
COPY nas/ nas/
COPY s1ap/ s1ap/
COPY lppa/ lppa/
COPY diameter/ diameter/
COPY per/ per/
RUN go mod download
COPY . .
RUN go build -o /out/core-tester ./cmd/core-tester

FROM ubuntu:26.04
RUN apt-get update && apt-get install -y --no-install-recommends \
      iptables iproute2 iputils-ping libpcap0.8t64 tcpdump ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/core-tester /bin/core-tester
