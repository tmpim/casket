# header

header can manipulate response headers.

Note that if you wish to remove response headers from a proxied backend, you must do so in the [proxy](/proxy)
directive.

## Syntax

``` casketfile
header path name value
```

-   **path** is the base path to match.
-   **name** is the name of the field. Prefix with a hyphen (`-`) to remove the header or a plus (`+`) to append instead
    of overwrite.
-   **value** is the value for the field. Dynamic values can also be inserted using [placeholders](/placeholders).

This directive can be used multiple times, or you can group multiple custom header fields for the same path:

``` casketfile
header path {
  name value
}
```

## Examples

Custom header for all pages:

``` casketfile
header / X-Custom-Header "My value"
```

Strip the "Hidden" field from header:

``` casketfile
header / -Hidden
```

Multiple custom headers for a specific path, while removing the Server field:

``` casketfile
header /api {
  Access-Control-Allow-Origin  *
  Access-Control-Allow-Methods "GET, POST, OPTIONS"
  -Server
}
```

Add some security headers to all pages:

``` casketfile
header / {
  # Enable HTTP Strict Transport Security (HSTS) to force clients to always
  # connect via HTTPS (do not use if only testing)
  Strict-Transport-Security "max-age=31536000;"
  # Enable cross-site filter (XSS) and tell browser to block detected attacks
  X-XSS-Protection "1; mode=block"
  # Prevent some browsers from MIME-sniffing a response away from the declared Content-Type
  X-Content-Type-Options "nosniff"
  # Disallow the site to be rendered within a frame (clickjacking protection)
  X-Frame-Options "DENY"
}
```
