FROM golang:1.17-bullseye AS builder

WORKDIR /workdir

COPY go.mod go.sum /workdir
RUN GOPROXY=https://proxy.golang.org,direct go mod download

COPY . /workdir
WORKDIR /workdir

RUN cd casket && go mod init casket \
    && GOPROXY=https://proxy.golang.org,direct GOPRIVATE=github.com/tmpim/casket go get "github.com/tmpim/casket@master" \
    && go mod tidy && go build -o casket

FROM alpine:3

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

COPY --from=builder /workdir/casket/casket /usr/bin/casket

ENTRYPOINT ["/usr/bin/casket", "-agree", "-root", "/www"]
