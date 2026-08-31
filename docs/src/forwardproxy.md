# forwardproxy

The forwardproxy directive enables Casket to act as a forward proxy, with support for HTTP/2.0 and HTTP/1.1 requests.
HTTP/2.0 will usually improve performance due to multiplexing. Forward proxy plugin includes common features like Access
Control Lists and authentication, as well as some unique features to assist with security and privacy. Default
configuration of forward proxy is compliant with existing HTTP standards, but some features force plugin to exhibit
non-standard but non-breaking behavior to preserve privacy. Probing resistance—one of the signature features of this
plugin—attempts to hide the fact that your web server is also a forward proxy, helping the proxy to stay under the
radar.

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/forwardproxy/README.md)**

## Examples

### Default Forward Proxy

``` casketfile
forwardproxy
```

Spins up default unauthenticated forwardproxy. Use with care, since other people might start (ab-)using your proxy.

### Authenticated Forward Proxy

``` casketfile
forwardproxy {
  basicauth casketuser1 0NtCL2JPJBgPPMmlPcJ
  basicauth casketuser2 秘密
}
```

Spins up authenticated forwardproxy.

### Probe Resistant Forward Proxy

``` casketfile
forwardproxy {
  basicauth casketuser1 0NtCL2JPJBgPPMmlPcJ
  probe_resistance hiddenlink-u13PJVFur3.com
}
```

Enable (experimental) probe resistance, which attempts to hide the fact that the site is a forwardproxy. Proxy will no
longer respond with &#34;407 Proxy Authentication Required&#34; if credentials are incorrect or absent, and will attempt
to mimic generic forwardproxy-less Casket server in other regards. Optionally, you can specify hiddenlink, which, when
visited, will prompt 407 response. Authentication is required.

### Generate and serve PAC file from memory

``` casketfile
forwardproxy {
  serve_pac proxyautoconfig.pac
}
```

Generate in memory and serve Proxy Auto-Config(see <https://en.wikipedia.org/wiki/Proxy_auto-config>) file on given
path. If no path is provided, PAC file will be served at /proxy.pac

### Hide user IP address from visited websites

``` casketfile
forwardproxy {
  hide_ip
}
```

Forward proxy will stop adding user&#39;s IP to &#34;Forwarded:&#34; header.

### Probe resistant Reverse Proxy

``` casketfile
forwardproxy {
  basicauth user1 0NtCL2JPJBgPPMmlPcJ
  probe_resistance
  upstream         https://user:password@someplace.com
}
```

upstream directive creates a reverse proxy(which simply relays *all* requests to given upstream without rewriting them),
and it was implemented in forwardproxy with a single purpose: to allow users to take advantage of probing resistance.
forwardproxy&#39;s Reverse Proxy implementation is simplistic, and if you need a powerful reverse proxy, but do not need
proving resistance, use Casket&#39;s standard proxy directive.
