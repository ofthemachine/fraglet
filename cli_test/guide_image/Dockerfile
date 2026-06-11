FROM alpine:latest

RUN apk add --no-cache bash

COPY fraglet-entrypoint /usr/local/bin/fraglet-entrypoint

COPY ./files /

RUN chmod +x /code/noop.sh

ENTRYPOINT ["/usr/local/bin/fraglet-entrypoint"]
