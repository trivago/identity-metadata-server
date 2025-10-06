# syntax=docker/dockerfile:1

# ------------------------------------------------------------------------------
#  Compress the binary with UPX
# ------------------------------------------------------------------------------

FROM alpine:latest AS upx

ARG CMD

# Set automatically by the buildx builder, but not by the docker builder
ARG TARGETARCH
ENV TARGETARCH=${TARGETARCH:-"amd64"}

ARG UPX_VERSION=4.2.1
ENV UPX_PATH=upx-${UPX_VERSION}-${TARGETARCH}_linux/upx

WORKDIR /workspace

RUN apk add curl
RUN curl -fsSL \
      "https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-${TARGETARCH}_linux.tar.xz" \
    | tar -xJf - ${UPX_PATH} \
 && chmod oga+x ${UPX_PATH}

COPY bin/$CMD-linux-$TARGETARCH ./$CMD
RUN ./${UPX_PATH} -1 -k $CMD

# ------------------------------------------------------------------------------
#  Runtime image
# ------------------------------------------------------------------------------

FROM cgr.dev/chainguard/static:latest

ARG CMD

WORKDIR /
USER nobody

COPY --chown=nobody --from=upx /workspace/${CMD} /entrypoint

ENTRYPOINT [ "/entrypoint" ]
