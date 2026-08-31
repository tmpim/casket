# tls

tls configures HTTPS connections. **Since HTTPS is [enabled automatically](/automatic-https), this directive should only
be used to deliberately override default settings. Use with care, if at all.**

Casket supports SNI (Server Name Indication), so you can serve multiple HTTPS sites from the same port on your machine.
In addition, Casket implements OCSP stapling for all qualifying certificates. Casket also automatically rotates all TLS
session ticket keys periodically.

The tls directive will ignore sites that are explicitly defined to be http:// or are on port 80. This allows you to use
the tls directive in a server block that is shared with both HTTP and HTTPS sites.

If not all the hostnames are not known when starting the server, you can use the [On-Demand
TLS](/automatic-https#on-demand) feature, which issues certificates during the TLS handshake rather than at startup.
Alternatively, if you just have more subdomains than CA rate limits allow, you can [enable the DNS
challenge](/automatic-https#enabling-the-dns-challenge) and obtain a single [wildcard
certificate](/automatic-https#wildcards).

::: warning
**Casket ships with sane defaults for cipher suites, curves, key types, and protocols.** Their exact selection and
ordering may change at any time with new releases. You probably do not need to change them yourself. Adjust the TLS
configuration at your own risk.
:::

Casket does not disambiguate between different or conflicting TLS configurations with the same hostname as the key. If a
TLS configuration is customized, then any other TLS configuration keyed by the same hostname must match, or at least be
compatible, or it is an error. This includes cipher suites, curve preferences, etc.

## Syntax

``` casketfile
tls off
```

Disables TLS for the site. Not recommended unless you have a good reason. With TLS off, automatic HTTPS is also
disabled, so the default port (2015) will not be changed.

``` casketfile
tls email
```

-   **email** is the email address to use with which to generate a certificate with a trusted CA. By providing an email
    here you will not be prompted when you run Casket.

Although the above syntax is not needed to enable TLS, it allows you to specify the email address used for your CA
account, instead of prompting for one or using another one from a previous run.

To use Casket with your own certificate and key:

``` casketfile
tls cert key
```

-   **cert** is the certificate file. If the certificate is signed by a CA, this certificate file should be a bundle: a
    concatenation of the server's certificate followed by the CA's certificate (root certificate usually not necessary).
-   **key** is the server's private key file which matches the certificate file.

Specifying your own certificate and key disables automatic HTTPS, including the changing of the port and redirecting
HTTP to HTTPS. You will need to do that yourself if you are managing your own certificates.

You can use this directive multiple times to specify multiple certificate and key pairs.

Or to have Casket generate and use an untrusted, self-signed certificate in memory that lasts 7 days (enough for local
development):

``` casketfile
tls self_signed
```

The above syntaxes use Casket's default TLS settings with your own certificate and key or a self-signed certificate that
lasts for 7 days: it intended for local development only.

Advanced users may open a settings block for more control, optionally specifying their own certificate and key:

``` casketfile
tls [cert key] {
  ca        uri
  protocols min max
  ciphers   ciphers...
  curves    curves...
  clients   [request|require|verify_if_given] clientcas...
  load      dir
  ask       url
  key_type  type
  dns       provider
  alpn      protos...
  must_staple
  wildcard
}
```

-   **ca** specifies the ACME-compatible Certificate Authority endpoint to request certificates from.
-   **cert** and **key** are the same as above.
-   **protocols** specifies the minimum and maximum protocol versions to support. See below for valid values. If min and
    max are the same, it need only be specified once.
-   **ciphers** is a list of space-separated ciphers that will be supported, overriding the defaults. If you list any,
    only the ones you specify will be allowed and preferred in the order given. See below for valid values. Customizing
    cipher suites is not allowed with TLS 1.3.
-   **curves** is a list of space-separated curves that will be supported in the given order, overriding the defaults.
    Valid curves are listed below.
-   **clients** is a list of space-separated client root CAs used for verification during TLS client authentication. If
    used, clients will be asked to present their certificate by their browser, which will be verified against this list
    of client certificate authorities. A client will not be allowed to connect if their certificate was not signed by
    one of these root CAs. Note that this setting applies to the entire listener, not just a single site. You may modify
    the strictness of client authentication using one of the keywords before the list of client CAs:
    -   **request** merely asks a client to provide a certificate, but will not fail if none is given or if an invalid
        one is presented.
    -   **require** requires a client certificate, but will not verify it.
    -   **verify_if_given** will not fail if none is presented, but reject all that do not pass verification.
    -   The default, if no flag is set but a CA file found, is to do both: to require client certificates and validate
        them.
-   **load** is a directory from which to load certificates and keys. The entire directory and its subfolders will be
    walked in search of .pem files. Each .pem file must contain the PEM-encoded certificate (chain) and key blocks,
    concatenated together.
-   **ask** enables [On-Demand TLS](/automatic-https#on-demand). On-Demand TLS is NOT recommended if the hostname is
    given in the Casketfile and known at configuration-time. The URL will be queried via GET and should return a 200
    status code if the `domain` form value from the query string is allowed to be given a certificate. Redirects at this
    endpoint are not followed. The URL should only be internally accessible. When using this option, Casket does not
    enforce any extra rate limiting; your endpoint is expected to make wise decisions instead.
-   **key_type** is the type of key to use when generating keys for certificates (only applies to managed or TLS or
    self-signed certificates). Valid values are rsa2048, rsa4096, rsa8192, p256, and p384. Default is currently p256.
-   **dns** is the name of a DNS provider; specifying it enables the [DNS challenge](/automatic-https#dns-challenge)
    (see that link for details). Note that you need to give credentials via environment variables for it to work.
-   **alpn** is a list of space-separated protocols to use for Application-Layer Protocol Negotiation (ALPN). For HTTPS
    servers, HTTP versions can be enabled/disabled with this setting. Default is `h2 http/1.1`.
-   **must_staple** enables
    [Must-Staple](https://blog.mozilla.org/security/2015/11/23/improving-revocation-ocsp-must-staple-and-short-lived-certificates/)
    for managed certificates. Use with care.
-   **wildcard** will obtain and manage a wildcard certificate for this name by replacing the left-most label with `*`,
    as long as managed TLS with the DNS challenge is enabled. Any sites which are configured similarly and have the same
    resulting wildcard name will then share the same, single certificate. This will not work with On-Demand TLS because
    it uses the SNI value for the certificate name.
    ::: warning
    Do not use this feature unless you have many subdomains that would otherwise cause you to hit CA rate limits.
    :::

## Protocols

The following protocols are supported, in descending order of preference:

-   tls1.3 (default max)
-   tls1.2 (default min)
-   tls1.1
-   tls1.0

Note that setting the minimum protocol version lower may allow very old clients to connect, but at the risk of a false
sense of security.

Supported protocols and default protocol versions may be changed at any time.

## Cipher Suites

The following cipher suites are currently supported:

-   ECDHE-ECDSA-AES256-GCM-SHA384
-   ECDHE-RSA-AES256-GCM-SHA384
-   ECDHE-ECDSA-AES128-GCM-SHA256
-   ECDHE-RSA-AES128-GCM-SHA256
-   ECDHE-ECDSA-WITH-CHACHA20-POLY1305
-   ECDHE-RSA-WITH-CHACHA20-POLY1305
-   ECDHE-RSA-AES256-CBC-SHA
-   ECDHE-RSA-AES128-CBC-SHA
-   ECDHE-ECDSA-AES256-CBC-SHA
-   ECDHE-ECDSA-AES128-CBC-SHA
-   RSA-AES256-CBC-SHA
-   RSA-AES128-CBC-SHA
-   ECDHE-RSA-3DES-EDE-CBC-SHA
-   RSA-3DES-EDE-CBC-SHA

::: info
Cipher suites cannot be customized when TLS 1.3 is used.
:::

::: info
The HTTP/2 spec blacklists over 275 cipher suites for security reasons. Unless you know what you're doing, it's best to
accept the default cipher suite settings.
:::

Cipher suites may be added to or removed from Casket at any time. Similarly, the default cipher suites may be changed at
any time.

## Curves

The following curves are supported for EC cipher suites:

-   X25519
-   p256
-   p384
-   p521

## Summary of Features {#summary}

In summary, Casket implements these TLS features for you automatically. It is the only server to do so by default:

-   Session identifiers
-   Session ticket key rotation
-   OCSP stapling
-   Dynamic record sizing
-   Application-layer protocol negotiation
-   Forward secrecy
-   HTTP/2 (for the HTTP server)
-   Certificate management (including [auto-renew](/automatic-https))
-   [Man-In-The-Middle detection](/mitm-detection) (for HTTPS sites)

Most sites will not even use the `tls` directive. It is not needed to use these features.

## Examples

Remember, TLS is enabled by default, and this directive is not usually needed! These examples are for advanced users who
manage certificates manually or need custom settings.

Serve with HTTPS using a certificate and private key located one folder up:

``` casketfile
tls ../cert.pem ../key.pem
```

Obtain certificates during TLS handshakes as needed, after consulting a backend for permission:

``` casketfile
tls {
  ask http://localhost:9095/should-obtain-cert
}
```

Load all certificates and keys from .pem files found in /www/certificates:

``` casketfile
tls {
  load /www/certificates
}
```

Serve a site with a self-signed certificate in memory (untrusted by browsers, but convenient for local development):

``` casketfile
tls self_signed
```
