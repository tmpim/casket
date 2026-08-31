# fastcgi

fastcgi proxies requests to a FastCGI server. Even though the most common use for this directive is to serve PHP sites,
it is by default a generic FastCGI proxy. This directive may be used multiple times with different request paths.

## Syntax

``` casketfile
fastcgi path endpoint [preset] {
  root     directory
  ext      extension
  split    splitval
  index    indexfile
  env      key value
  except   ignored_paths...
  upstream endpoint
  connect_timeout duration
  read_timeout    duration
  send_timeout    duration
}
```

-   **path** is the base path to match before the request will be forwarded.
-   **endpoint** is the address or Unix socket of the FastCGI server.
-   **preset** is an optional preset name (see below). You do not need to repeat a preset's individual settings when
    using a preset.
-   **root** specifies the root directory used by the FastCGI server if different from the root directory of the virtual
    host. Useful if the FastCGI server is on a different server, chroot-jailed, and/or containerized.
-   **ext** specifies the extension which, if the request URL has it, would proxy the request to FastCGI.
-   **split** specifies how to split the URL; the split value becomes the end of the first part and anything in the URL
    after it becomes part of the PATH_INFO CGI variable.
-   **index** specifies the default file to try if a file is not specified by the URL.
-   **env** sets an environment variable named *key* with the given *value*; the **env** property can be used multiple
    times and values may use [request placeholders](/placeholders).
-   **except** is a list of space-separated request paths to be excepted from fastcgi processing, even if it matches the
    base path.
-   **upstream** specifies an additional backend to use. Basic load balancing will be performed. This can be specified
    multiple times.
-   **connect_timeout** is the time allowed for connecting to the backend. Must be a duration value (e.g. "10s").
    Default 60s.
-   **read_timeout** is the time allowed to read a response from the backend. Must be a duration value. Default 60s.
-   **send_timeout** is the time allowed to send a request to the backend. Must be a duration value. Default 60s.

For HTTPS connections, the following environment variables are set, in a manner compliant with httpd's mod_ssl: `HTTPS`,
`SSL_PROTOCOL`, and `SSL_CIPHER`.

## Presets

A preset is shorthand for a certain FastCGI configuration. These presets are available:

-   **php** is shorthand for:

    ``` casketfile
    ext   .php
    split .php
    index index.php
    ```

**You do not need to specify the individual configuration settings for a preset.** However, you can overwrite its
individual settings if needed by declaring them manually.

## Examples

Proxy all requests to a FastCGI responder listening at 127.0.0.1:9000:

``` casketfile
fastcgi / 127.0.0.1:9000
```

Forward all requests in /blog to a PHP site (like WordPress) being served with php-fpm:

``` casketfile
fastcgi /blog/ 127.0.0.1:9000 php
```

With custom FastCGI configuration:

``` casketfile
fastcgi / 127.0.0.1:9001 {
  split .html
}
```

With PHP preset, but overriding the ext property:

``` casketfile
fastcgi / 127.0.0.1:9001 php {
  ext .html
}
```

With PHP preset, but the FastCGI server is running in a container based on an [official Docker
image](https://hub.docker.com/_/php/) (with container port 9000 published to 127.0.0.1:9001):

``` casketfile
fastcgi / 127.0.0.1:9001 php {
  root /var/www/html
}
```
