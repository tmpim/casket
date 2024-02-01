FROM golang:1.19-bullseye AS builder

WORKDIR /workdir

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOPRIVATE=github.com/tmpim/casket
COPY go.mod go.sum /workdir
RUN go mod download

COPY . /workdir
WORKDIR /workdir/casket

# Required to build with version information - but allow this step to fail (e.g. we're building a PR). Casket will try
# to get the version from the module (this step) first, and then try to get it from `main.version` (goreleaser and
# ldflags). See also:
# - casket/casketmain/run.go#getBuildModule()
# - https://goreleaser.com/cookbooks/using-main.version/
RUN go get "github.com/tmpim/casket@master"; exit 0

ENV CGO_ENABLED=0
# -s: Omit the symbol table and debug information
# -w: Omit the DWARF symbol table
# -X: Include the git tag as the version (goreleaser also uses main.version tag)
RUN go build -ldflags="-s -w -X 'main.version=$(git describe --tags --dirty)'" -o casket .

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