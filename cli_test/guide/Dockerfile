FROM alpine:latest

ARG FRAGLET_VERSION=v0.4.0

RUN apk add --no-cache curl bash

RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        BINARY="fraglet-entrypoint-linux-amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then \
        BINARY="fraglet-entrypoint-linux-arm64"; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi && \
    curl -L -o /usr/local/bin/fraglet-entrypoint \
      "https://github.com/ofthemachine/fraglet/releases/download/${FRAGLET_VERSION}/${BINARY}" && \
    chmod +x /usr/local/bin/fraglet-entrypoint

COPY ./files /

RUN chmod +x /code/noop.sh

ENTRYPOINT ["/usr/local/bin/fraglet-entrypoint"]
