# Casket

Casket is a fork of [mholt's Caddy web server](https://github.com/caddyserver/caddy) v1.
Its goal is to maintain Caddy's original goal of being a straight forward, simple
to use web server rather than the direction of Caddy v2 which has a focus on
microservices and programtic configurability.

Casket will come with all the features you love about Caddy v1, while also
adding our own touches for convenience and usability.

## Building

To build the main casket executable, the following procedure can be used:
```sh
git clone https://github.com/tmpim/casket
cd casket
go build -o ./build/casket ./casket
# The executable can now be found at ./build/casket
```

Note these development builds will lack version information and will report a version of (devel). You can also instead create a snapshot build using the following procedure:

```sh
go install github.com/goreleaser/goreleaser@latest # Install goreleaser
goreleaser build --snapshot --rm-dist --single-target --id casket # Create a snapshot build
# The executable can now be found at ./build/casket_linux_amd64/casket
```

## Docker

A docker image is provided for the latest version of Casket from `ghcr.io/tmpim/casket`.

Example using docker-compose:
```yaml
version: "3.8"

services:
    casket:
        image: ghcr.io/tmpim/casket:latest # or a specific version like v1.2
        restart: unless-stopped
        ports:
            - "80:80" # HTTP
            - "443:443" # HTTPS
        volumes:
            - ./Casketfile:/config/Casketfile # Pass in your casket config
            - ./static:/www # Pass in your static content
            - casket_data:/data # Create a volume to store persistent data (e.g. certificates)
```