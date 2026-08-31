# ratelimit

ratelimit is used to limit the request processing rate based on client&#39;s IP address/request header. Excessive
requests will be terminated with an error 429 (Too Many Requests) and `X-RateLimit-RetryAfter` header will be returned.

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/rate-limit/README.md)**

## Examples

### For single resource:

``` casketfile
ratelimit methods path rate burst unit
```

`methods` are the request methods it will match (comma separately); `path` is the file or directory to apply rate limit;
`rate` is the limited request in every time unit (r/s, r/m, r/h, r/d, r/w) (e.g. 1); `burst` is the maximum burst size
client can exceed; burst &gt;= rate (e.g. 2); `unit` is the time interval (currently support: `second`, `minute`,
`hour`, `day`, `week`).

### For multiple resources:

``` casketfile
ratelimit methods rate burst unit {
  whitelist CIDR,CIDR
  limit_by_header xxx
  status xxx,xxx
  resources
}
```

`whitelist` is the keyword for whitelisting your trusted ips (comma separately).
[CIDR](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing) is the IP range you don&#39;t want to perform rate
limit, `whitelist` is a general rule, it won&#39;t target for specific resources; `limit_by_header` is the keyword for
matching the request header. Like `whitelist`, it&#39;s also a general rule (normally you shouldn&#39;t apply this rule
unless the default &#39;limit by ip&#39; is not what you want and you want to &#39;limit by request header&#39;);
`status` is the keyword for matching the [response status code](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes)
(comma separately). If this rule is triggered, all subsequent requests from that client will be blocked regardless of
which status code is returned or which resource is requested (this won&#39;t block resources not defined in
ratelimit&#39;s config); `resources` is a list of files/directories to apply rate limit, one per line. Note: If you
don&#39;t want to apply rate limit on some special resources, add `^` in front of the path.

### Examples

* Limit clients to 2 requests per second (bursts of 3) to any methods and any resources under /r:
    ``` casketfile
    ratelimit * /r 2 3 second
    ```

* Don't perform rate limit if requests come from 1.2.3.4 or 192.168.1.0/30(192.168.1.0 \~ 192.168.1.3), for the listed
  paths, limit clients to 2 requests per minute (bursts of 2) if the request method is GET or POST and always ignore
  /dist/app.js:
    ``` casketfile
    ratelimit 2 2 minute {
      whitelist 1.2.3.4/32,192.168.1.0/30
      status *
      /foo.html
      /api
      ^/dist/app.js
    }
    ```
