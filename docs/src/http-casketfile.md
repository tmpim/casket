# The HTTP Casketfile

This page documents how the HTTP server uses the Casketfile. If you haven't already, take the [Casketfile
tutorial](/tutorial/casketfile) or read up on the [Casketfile syntax](/casketfile) first.

## Site Addresses {#addresses}

The HTTP server uses site addresses for [labels](/casketfile#structure). Addresses are specified in the form
`scheme://host:port/path`, where all but one are optional.

The host portion is usually localhost or the domain name. The default port is 2015 (unless the site qualifies for
[automatic HTTPS](/automatic-https), in which case it's changed to 443 for you). The scheme portion is another way to
specify a port. Valid schemes are "http" or "https" which represent, respectively, ports 80 and 443. If both a scheme
and port are specified, the port takes precedence. For example (this table assumes automatic HTTPS is applied where it
qualifies):

``` casketfile
:2015                    # Host: (any), Port: 2015
localhost                # Host: localhost; Port: 2015
localhost:8080           # Host: localhost; Port: 8080
example.com              # Host: example.com; Ports: 80->443
http://example.com       # Host: example.com; Port: 80
https://example.com      # Host: example.com; Ports: 80->443
http://example.com:1234  # Host: example.com; Port: 1234
https://example.com:80   # Error! HTTPS on port 80
*                        # Hosts: *; Port: 2015
*.example.com            # Hosts: *.example.com; Port: 443
*.*.example.com          # Hosts: *.*.example.com; Port: 2015
example.com/foo/         # Host: example.com; Ports: 80, 443; Path: /foo/
/foo/                    # Host: (any), Port: 2015, Path: /foo/
```

Site addresses must be unique; all configuration for a site must be grouped in the same definition.

Wildcard characters `*` can be used in a hostname. A wildcard must take the place of an entire domain label:
`*.example.com` is valid but `foo*.example.com` is not. Hostnames may have more than one wildcard label, but they must
be the left-most labels. (Note that to obtain a wildcard certificate, only one wildcard label is allowed; this special
case allows automatic HTTPS to stay on.) A site defined as `*` matches only one-label names like "localhost" but not
"example.com". To match all hosts, leave the hostname empty.

## Path Matching

Some directives accept an argument that specifies a *base path* to match. A base path is a prefix. If the URL starts
with the base path, it will be a match. For example, a base path of `/foo` will match requests to `/foo`, `/foo.html`,
`/foobar`, and `/foo/bar.html`. If you want to limit a base path to match a specific directory only, then suffix it with
a forward slash like `/foo/`, which will *not* match `/foobar`.

## Directives

Most directives invoke a layer of middleware. Middleware is a small layer in the application that handles HTTP requests
and does one thing really well. Middleware are chained together (pre-compiled, if you will) at startup. Only middleware
handlers which are invoked from the Casketfile will be chained in, so small Casketfiles are very fast and efficient.

The syntax of arguments varies from directive to directive. Some have required arguments, others don't. Consult the
documentation of each directive for their signatures.

For directives which are registered as plugins, the documentation pages will show the directive name prefixed with the
server type, for example, "http.realip" or "dns.dnssec". When using them in the Casketfile, drop the prefix ("http.").
The prefix is just to assure unique naming, but is not used in the Casketfile.

## Placeholders

In some cases, directives will accept [placeholders](/placeholders) (replaceable values). These are words that are
enclosed by curly braces `{ }` and interpreted by the HTTP server at request-time. For example, `{query}` or
`{>Referer}`. Think of them like variables. These placeholders have no relationship to the environment variables you can
use in Casketfiles, except we borrowed the syntax for familiarity.
