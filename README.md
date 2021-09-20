# Casket

Casket is a fork of [mholt's Caddy web server](https://github.com/caddyserver/caddy) v1.
Its goal is to maintain Caddy's original goal of being a straight forward, simple
to use web server rather than the direction of Caddy v2 which has a focus on
microservices and programtic configurability.

Casket will come with all the features you love about Caddy v1, while also
adding our own touches for convenience and usability.

## Building

To build the main casket executable, the following procedure can be used:
```
git clone https://github.com/tmpim/casket
cd casket
go build -o ./build/casket ./casket
# The executable can now be found at ./build/casket
```

Note these development builds will lack version information and will report a version of (devel)
