# limits

limits sets limits on HTTP requests that are accepted. The request header and body sizes can be limited.

The request header size can be capped at some number of bytes. Due to the way the Go libraries work, one limit applies
to all sites on the same listener (if there are multiple configured, the minimum limit is selected).

Request body sizes can also be limited. When the number of bytes read from a request body exceeds the limit, reading
will terminate and an error will be sent to the client. (Technically, it depends on each middleware that reads a request
body to handle the error properly, but standard Casket directives should respond with HTTP 413.) By default, there is no
size limit.

The size values must be positive integers and are interpreted as bytes unless a unit is given. Valid examples: `3500`
(3,500 bytes), `500kb` (500 kilobytes), `10mb` (10 megabytes), `1gb` (1 gigabyte).

## Syntax

``` casketfile
limits size
```

-   **size** is the maximum size for both headers and bodies (individually) for the whole site.

For more control, you can open a block:

``` casketfile
limits {
  header size
  body   [path] size
}
```

-   **header** restricts the size of the request header.
-   **body** restricts the size of a request body, optionally only for a base path.

## Examples

Limit all request headers and bodies to 7.5 kilobytes:

``` casketfile
limits 7500
```

Limit only request bodies within /upload to 50 megabytes:

``` casketfile
limits {
  body /upload 50mb
}
```

Various limits:

``` casketfile
limits {
  header 100KB
  body   /upload 100MB
  body   /profile 25KB
  body   /api 10KB
}
```
