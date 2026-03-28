FROM alpine:latest

# Install bash for our script
RUN apk add --no-cache bash

# Copy the fraglet-entrypoint binary
COPY fraglet-entrypoint /usr/local/bin/fraglet-entrypoint

# Copy all files from files/ directory to root
COPY ./files /

# Make the scripts executable
RUN chmod +x /code/hello-world.sh /code/hello-world-range.sh /code/hello-world-replace.sh

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/fraglet-entrypoint"]
