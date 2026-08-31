# realip

This plugin allows you to see the actual client IP from `X-Forwarded-For` headers if you are running behind a CDN or
Proxy. It will make it so logs and other downstream directives will see the actual client IP, *not* the proxy&#39;s.
Implements security measures so that X-Forwarded-For cannot be spoofed from unauthorized IP ranges.

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/realip/README.md)**

## Examples

### Accept X-Forwarded-For from known IPs only

``` casketfile
realip {
  from 1.2.3.4/32
  from 2.3.4.5/32
}
```

### Cloudflare Preset

``` casketfile
realip cloudflare
```

Automatically adds cloudflare ips to allowed set.
