FROM alpine:3

COPY casket /usr/bin/

RUN apk --no-cache add tzdata ca-certificates && update-ca-certificates

# Create empty directories for:
# /config: where the casketfile other configuration files will be stored
# /data: where the persistent data will be stored (certificates, etc)
# /www: the default web root
# NOTE: it is your responsibility (the user) to create bind mounts or volumes.
RUN mkdir /config \
    && mkdir /data \
    && mkdir /www

# Set the casket path to store certificates in /data
ENV CASKETPATH=/data

# Set working directory to /config so that casket will load /config/Casketfile
WORKDIR /config

ENTRYPOINT ["casket", "-agree", "-root", "/www"]