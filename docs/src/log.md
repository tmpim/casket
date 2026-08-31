# log

log enables request logging. The request log is also known from some vernaculars as an access log.

## Syntax

With no arguments, an access log is written to access.log in the common log format for all requests:

``` casketfile
log
```

Customize the log file location:

``` casketfile
log file
```

-   **file** is the path to the log file to create (or append to), relative to the current working directory. See [Log
    Destination](#destination) for more details about how to specify an output location. Default is access.log.

To restrict this log to certain requests or to change the log format:

``` casketfile
log path file [format]
```

-   **path** is the base request path to match in order to be logged.
-   **file** is the log file to create (or append to), relative to current working directory.
-   **format** is the log format to use; default is Common Log Format.

Large log files are rolled automatically. You can customize log rolling or other things by opening a block:

``` casketfile
log path file [format] {
  rotate_size     mb
  rotate_age      days
  rotate_keep     count
  rotate_compress
  ipmask          ipv4_mask [ipv6_mask]
  except          paths...
}
```

-   **rotate_size** is the size in megabytes a log file must reach before rolling it.
-   **rotate_age** is how long in days to keep rotated log files.
-   **rotate_keep** is the maximum number of rotated log files to keep; older rotated log files get pruned.
-   **rotate_compress** is the option to compress rotated log files. gzip is the only format supported.
-   **ipmask** enables masking IP addresses to comply with corporate or legal restrictions. The first argument is a mask
    for IPv4 addresses, and the second argument is a mask for IPv6 addresses. The IPv6 mask is optional; and if only
    IPv6 is to be masked, the IPv4 mask can be an empty string token.
-   **except** exempts requests by path from being logged. More than one path can be specified per line
    (space-separated), if desired, or this subdirective can be used multiple times.

## Log Format

You can specify a custom log format with any [placeholder](/placeholders) values. Log supports both request and response
placeholders.

Currently there are two predefined formats.

-   **{common}** (default)

    ``` casketfile
    {remote} - {user} [{when}] \"{method} {uri} {proto}\" {status} {size}
    ```

-   **{combined}** - {common} appended with

    ``` casketfile
    \"{>Referer}\" \"{>User-Agent}\"
    ```

## Log Destination {#destination}

The log destination can be one of a few things:

-   a filename relative to the current working directory
-   `stdout` or `stderr` to write to the console
-   `syslog` to write to the local system log (except on Windows)
-   `syslog://host[:port]` to write via UDP to a local or remote syslog server
-   `syslog+udp://host[:port]` is the same as above
-   `syslog+tcp://host[:port]` to write via TCP to local or remote syslog server

## Log Rolling {#rolling}

Logs have the potential to fill the disk. To mitigate this, request logs are rotated ("rolled") automatically according
to this default configuration:

``` casketfile
rotate_size 100 # Rotate a log when it reaches 100 MB
rotate_age  14  # Keep rotated log files for 14 days
rotate_keep 10  # Keep at most 10 rotated log files
rotate_compress # Compress rotated log files in gzip format
```

You can specify these subdirectives to customize log rolling.

## Examples

Log all requests to access.log:

``` casketfile
log
```

Log all requests to stdout:

``` casketfile
log stdout
```

Custom log format:

``` casketfile
log / stdout "{proto} Request: {method} {path}"
```

Predefined format:

``` casketfile
log / stdout "{combined}"
```

With rotation:

``` casketfile
log requests.log {
  rotate_size 50  # Rotate after 50 MB
  rotate_age  90  # Keep rotated files for 90 days
  rotate_keep 20  # Keep at most 20 log files
  rotate_compress # Compress rotated log files in gzip format
}
```

To mask (anonymize) IPv4 addresses and IPv6 addresses down to a couple octets:

``` casketfile
log requests.log {
  ipmask 255.255.0.0 ffff:ffff:ffff:ff00::
}
```
