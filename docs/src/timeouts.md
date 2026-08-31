# timeouts

timeouts configures Casket's HTTP timeouts:

-   **Read:** Maximum duration for reading the entire request, including the body.
-   **Read Header:** The amount of time allowed to read request headers.
-   **Write:** The maximum duration before timing out writes of the response.
-   **Idle:** The maximum time to wait for the next request (when using keep-alive). Default is 5 minutes.

Timeouts are an important way to maintain server connectivity in the face of buggy or malicious clients.

Because timeouts apply to an entire HTTP server which may serve multiple sites defined in your Casketfile, the timeout
values for each site will be reduced to their minimum values (with 0 or none being the lowest) across all sites that
share that server. It's a good idea to keep your timeouts the same or just set them on one site to avoid confusion. A
timeout set on one site will apply to all sites that share that server.

## Syntax

To set all the timeouts to the same value:

``` casketfile
timeouts val
```

-   **val** is a duration value (e.g. 30s, 2m30s, 5m, 1h) that will apply to all timeouts. Set to `0` or `none` to
    disable timeouts that were previously enabled.

You can also configure each timeout individually:

``` casketfile
timeouts {
  read   val
  header val
  write  val
  idle   val
}
```

Valid values are durations, or `0` or `none` to disable the timeout if it was previously enabled.

## Examples

Set all timeouts to 1 minute:

``` casketfile
timeouts 1m
```

Set custom read timeout and write timeouts:

``` casketfile
timeouts {
  read  30s
  write 20s
}
```
