# errors

errors allows you to set custom error pages and enable error logging.

Without this middleware, error responses (HTTP status \>= 400) are not logged and the client receives a plaintext error
message.

Using an error log, the text of each error will be recorded so you can determine what is going wrong without exposing
those details to the clients. With error pages, you can present custom error messages and instruct your visitor what to
do. When you specify custom error pages, error logging is automatically enabled.

## Syntax

``` casketfile
errors [logfile]
```

-   **logfile** is the path to the error log file to create (or append to), relative to the current working directory.
    See [Log Destination](#destination) for more details about how to specify an output location. Default is `stderr`.

To specify custom error pages, open a block:

``` casketfile
errors [logfile] {
  code     file
  rotate_size     mb
  rotate_age      days
  rotate_keep     count
  rotate_compress
}
```

-   **code** can be an HTTP status code (4xx, 5xx, or `*` for default error page).
-   **file** is the static HTML file of the error page (path is relative to site root).
-   **rotate_size** is the size in megabytes a log file must reach before rolling it.
-   **rotate_age** is how long in days to keep rotated log files.
-   **rotate_keep** is the maximum number of rotated log files to keep; older rotated log files get pruned.
-   **rotate_compress** is the option to compress rotated log files. gzip is the only format supported.

## Log Destination {#destination}

The log destination can be one of a few things:

-   a filename relative to the current working directory
-   `stdout` or `stderr` to write to the console
-   `visible` to write the error (including full stack trace, if applicable) to the response (NOT recommended except for
    local debugging)
-   `syslog` to write to the local system log (except on Windows)
-   `syslog://host[:port]` to write via UDP to a local or remote syslog server
-   `syslog+udp://host[:port]` is the same as above
-   `syslog+tcp://host[:port]` to write via TCP to local or remote syslog server

The default log destination is `stderr`.

## Log Rolling {#rolling}

Logs have the potential to fill the disk. To mitigate this, error logs are rotated ("rolled") automatically according to
this default configuration:

``` casketfile
rotate_size 100 # Rotate a log when it reaches 100 MB
rotate_age  14  # Keep rotated log files for 14 days
rotate_keep 10  # Keep at most 10 rotated log files
rotate_compress # Compress rotated log files in gzip format
```

You can specify these subdirectives to customize log rolling.

## Examples

Log errors to stderr:

``` casketfile
errors
```

Log errors to a custom file in the parent directory:

``` casketfile
errors ../error.log
```

Log errors and serve custom error pages:

``` casketfile
errors {
  404 404.html # Not Found
  500 500.html # Internal Server Error
}
```

Log errors to custom log file and serve custom error pages:

``` casketfile
errors ../error.log {
  404 404.html # Not Found
  500 500.html # Internal Server Error
}
```

Define a default, catch-all error page:

``` casketfile
errors {
  * default_error.html
}
```

Make errors visible to the client (for debugging only):

``` casketfile
errors visible
```

Customize error log rolling:

``` casketfile
errors {
  rotate_size 50  # Rotate after 50 MB
  rotate_age  90  # Keep rotated files for 90 days
  rotate_keep 20  # Keep at most 20 log files
  rotate_compress # Compress rotated log files in gzip format
}
```
