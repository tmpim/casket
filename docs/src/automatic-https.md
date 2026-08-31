# Automatic HTTPS

Casket [automatically enables HTTPS](https://github.com/mholt/certmagic) for all your sites, given that some reasonable
criteria are met:

<div class="tight-list">

-   The hostname:
    -   is not empty
    -   is not localhost
    -   is not an IP address
    -   has no more than 1 wildcard (`*`)
    -   wildcard must be left-most label
-   The port is not explicitly 80
-   The scheme is not explicitly http
-   TLS is not turned off in site's definition
-   Certificates and keys are not provided by you
-   Casket is able to bind to ports 80 and 443 (unless you use the DNS challenge)

</div>

Casket will also redirect all HTTP requests to their HTTPS equivalent if the plaintext variant of the hostname is not
defined in the Casketfile.

All pertinent assets are fully managed, including renewals—no action is required by you. Here's a 28-second
[video](https://www.youtube.com/watch?v=nk4EWHvvZtI) showing how it works:

<iframe width="560" height="315" src="https://www.youtube-nocookie.com/embed/nk4EWHvvZtI?si=5VbBfmMmRbWASZ2N" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Things to Know / FAQ {#faq}

In order to fully enjoy this flagship feature, please read the following.

### Ports 80 and 443 must be externally open {#ports}

By default, Casket will bind to ports 80 and 443 to serve HTTPS and redirect HTTP to HTTPS. This usually requires
privilege escalation. On Linux systems, you can give Casket permission to bind to port 80 and 443 without being root
using [setcap](http://man7.org/linux/man-pages/man8/setcap.8.html), like so: `setcap cap_net_bind_service=+ep casket`.
Don't forget to configure all relevant firewalls to allow Casket to use these ports for incoming and outgoing
connections! Casket must have claim on at least one of these ports to obtain certificates unless you enable the DNS
challenge OR forward ports 80 and 443 to different ports internally (in which case you can change the HTTP and HTTPS
ports using [CLI flags](/cli)). Although technically only one of these ports is absolutely required, unavailability of
one port may result in temporarily-degraded certificate management results unless the associated ACME challenge is
disabled, but should not ultimately affect uptime.

### The .casket folder {#dot-casket}

Casket will create a folder in your home directory called `.casket`. It uses this to store and manage cryptographic
assets required to serve your site privately over HTTPS. If there is no home folder, the .casket folder is created in
the current working directory unless `CASKETPATH` is set. The home folder is learned from the environment (`$HOME` or
`%HOMEPATH%`). Multiple Casket instances can use or mount the `acme` subfolder as a disk and Casket will automatically
share the certificates and coordinate maintenance between them.

::: warning
Your sites' certificates and private keys are stored here. Take care to back up and protect this folder.
:::

### Testing, developing, and advanced setups {#testing}

To test or experiment with your Casket configuration, make sure you use the `-ca` flag to change the ACME endpoint to a
staging or development URL, otherwise you are likely to hit rate limits which can block your access to HTTPS for up to a
week. This is especially common when using process managers or containers. Casket's default CA is [Let's
Encrypt](https://letsencrypt.org), which has [a staging endpoint](https://letsencrypt.org/docs/staging-environment/)
that is not subject to the same rate limits.

### Behind a load balancer or proxy {#proxying}

If Casket is behind other infrastructure like a load balancer, it may have trouble obtaining certificates. It is your
responsibility, then, to ensure SSL certificates are obtained and properly set up on all machines. In most cases like
these, we recommend using the DNS challenge (described below) to obtain certificates. It's not used by default, but it's
very easy to configure. If you have a fleet of Casket instances, they will automatically coordinate certificate
management as long as they share (mount) the same \$CASKETPATH/acme folder.

### Sharing certificates between multiple Casket instances {#fleet}

As of version 0.10.12, Casket supports using automatic HTTPS in a fleet/cluster configuration. As of version 0.11.2,
this is done via clustering plugins. For example, Casket can join a cluster by using the file system, Amazon S3, Consul,
and others through these plugins. Simply build Casket with your desired clustering plugin and set the
`CASKET_CLUSTERING` environment variable. See the docs for your clustering plugin to learn how to configure it. Also
[see which storage backends Casket supports](https://github.com/mholt/certmagic/wiki/Storage-Implementations) for TLS
assets.

### Wildcard certificates {#wildcards}

Casket can obtain and manage wildcard certificates when it is configured to serve a site with a qualifying wildcard
name. A site name qualifies for a wildcard if only its left-most domain label is a wildcard. For example,
`*.example.com` qualifies, but these do not: `sub.*.example.com`, `foo*.example.com`, `*bar.example.com`,
`*.*.example.com`, etc. To get a wildcard, you simply need to enable the DNS challenge (described below; it's very
easy). We recommend using wildcards *only* when you have so many subdomains that you would encounter CA rate limits
trying to obtain certificates for them all. If you have many subdomains configured differently in your Casketfile, you
can also force a wildcard for them by using the `wildcard` subdirective of the [tls](/tls) directive.

### Transparency reports {#transparency}

When Casket obtains a certificate from a CA that publishes certificate transparency logs, it is requisite that your
domain name and/or IP address will be included in those logs, as they are not considered private information. (Let's
Encrypt is one such CA.) This is a good thing; certificate transparency reports [keep the CAs
accountable](https://googleonlinesecurity.blogspot.com/2015/10/sustaining-digital-certificate-security.html).

## DNS Challenge

There are multiple challenge types by which Casket can obtain certificates. Casket uses all but one of them without any
configuration; another one—the DNS challenge—requires configuration and can be used when the other challenge types would
fail. [Casket supports many DNS providers.](https://github.com/tmpim/dnsproviders)

To use the DNS challenge, you must do three things. **1)** Download Casket with your provider plugged in. **2)** Tell
Casket which provider to use; this is done in the Casketfile. **3)** Give Casket credentials to access your account to
solve the challenge; this is done with environment variables.

### Enabling the DNS Challenge

In your Casketfile, you will use the [tls](/tls) directive with the dns keyword like so:

``` casketfile
tls {
  dns provider
}
```

Replace "provider" with the name of your DNS provider (in the table below). You will also need to specify the account
credentials for your provider, either via environment variables or in the Casketfile. See the provider's documentation
for details.

| Provider         | Name to use in Casketfile             |
|------------------|---------------------------------------|
| ACME-DNS         | [acmedns](/tls.dns.acmedns)           |
| Azure DNS        | [azure](/tls.dns.azure)               |
| Cloudflare       | [cloudflare](/tls.dns.cloudflare)     |
| DigitalOcean     | [digitalocean](/tls.dns.digitalocean) |
| DNSPod           | [dnspod](/tls.dns.dnspod)             |
| Duck DNS         | [duckdns](/tls.dns.duckdns)           |
| Gandi            | [gandi](/tls.dns.gandi)               |
| GoDaddy          | [godaddy](/tls.dns.godaddy)           |
| Google Cloud DNS | [googlecloud](/tls.dns.googlecloud)   |
| Linode           | [linode](/tls.dns.linode)             |
| Namecheap        | [namecheap](/tls.dns.namecheap)       |
| Name.com         | [namedotcom](/tls.dns.namedotcom)     |
| OVH              | [ovh](/tls.dns.ovh)                   |
| PowerDNS         | [pdns](/tls.dns.pdns)                 |
| RFC 2136         | [rfc2136](/tls.dns.rfc2136)           |
| Route53 by AWS   | [route53](/tls.dns.route53)           |
| TransIP          | [transip](/tls.dns.transip)           |
| Vultr            | [vultr](/tls.dns.vultr)               |

When you configure the DNS challenge, Casket will use that challenge type exclusively. Note that some providers may be
slow to apply changes (on the order of minutes).

## On-Demand TLS {#on-demand}

Casket pilots a new technology called On-Demand TLS. This means Casket can obtain a certificate for your site during the
first TLS handshake for a hostname that does not yet have a certificate.

To enable on-demand TLS, use the `tls` directive with either `max_certs` or `ask`. For example, your Casketfile might
look like this:

``` casketfile
*.example.com
proxy / localhost:4001 localhost:4002
tls {
  max_certs 10
}
```

This Casketfile will proxy all requests on subdomains to `example.com` to your backends for the first 10 unique
hostnames. This means you can dynamically provision new DNS records, and they will just start working with HTTPS. Unlike
wildcard certificates, on-demand certificates are not limited to subdomains.

You could also use `ask` to query a local backend whether a hostname should be allowed to get a certificate:

``` casketfile
tls {
  ask http://localhost:9005/allowed
}
```

If the HTTP request returns 200, a certificate will be requested, on-demand, from the CA. At least one of `max_certs` or
`ask` is required for On-Demand TLS.

::: warning Mitigating abuse
This feature is intentionally rate-limited. The *max_certs* property of the [tls](/tls) directive sets a hard limit on
how many new certificates are issued this way, so that even over a long period of time, attackers cannot issue unlimited
certificates and fill up your disk space. If this rate limiting as a safeguard is unacceptable in your case, you can
alternatively use the *ask* subdirective to specify an internal URL from which Casket can ask if a certain hostname is
authorized for a certificate (Casket's built-in rate limits do not apply when using *ask*!)
:::

On-Demand TLS is a special kind of managed TLS, so all the requirements above still apply except for the one about not
providing your own certificate and key: you may supplement on-demand TLS with your own certificates. And like regular
managed TLS, HTTP will be redirected to HTTPS.

::: warning Future support
This feature relies on the CA issuing certificates without delay. If instantaneous issuance becomes uncommon among ACME
CAs, we may discontinue this feature in Casket.
:::

Once a certificate is obtained on-demand, it is stored on disk just like any other managed certificates. It is also
stored in memory for quick retrieval during future handshakes.

On-Demand TLS is subject to these rate limits:

-   At most one certificate challenge happens at a time.
-   After 10 certificates are successfully obtained, new certificate challenges will not happen until 10 minutes after
    the last successful challenge.
-   A name that fails a challenge will not be allowed to be attempted again for 5 minutes.

Note that rate limits are reset when the process exits. Using the `-log` flag is recommended, since all certificate
challenges are logged. **Note that these built-in rate limits do not apply if an *ask* URL is specified in the
configuration.** When using *ask*, your endpoint must return a 200 status code for the name in order for a certificate
validation to be performed.

::: warning Testing your configuration
If your CA has a staging endpoint, it is *strongly* recommended that you use it to test your configuration. Let's
Encrypt is Casket's default CA, and [it has a staging endpoint](https://letsencrypt.org/docs/staging-environment/) that
is not subject to tight rate limits.
:::

The rest of this page explains more details about automatic HTTPS, but it is not required knowledge for using Casket.

## Obtaining Certificates

To serve a site over HTTPS, a valid SSL certificate is required from a trusted certificate authority (CA). When Casket
starts, it obtains certificates for eligible sites from [Let's Encrypt](https://letsencrypt.org). The following process
is nearly entirely *automatic* and *on by default*.

If necessary, Casket creates an account on the CA's server with (or without) your email address. Casket may have to
prompt you for an email address if it is not able to find one from the Casketfile, in the command line flags, or on disk
from a previous run. Use the `-agree` flag along with providing an email address to ensure you are not prompted to agree
to terms, when using in automated environments. But that should only be needed the first time automatic HTTPS is used.

Once the formalities are taken care of, Casket generates a private key and a Certificate Signing Request (CSR) for each
site. The private keys never leave the server and are safely stored on your file system.

Casket establishes a link with the CA's server. A brief cryptographic transaction takes place to prove that Casket
really is serving the sites it says it is. Once the CA server verifies this, it sends the certificate for that site over
the wire to Casket, which tucks it neatly away in the .casket folder.

This process usually takes a few seconds per domain, so once a certificate has been obtained for a site, it is simply
loaded from disk and reused the next time Casket is run. In other words, this delayed startup is a one-time event. If an
existing certificate needs to be renewed, Casket takes care of it right away.

Casket synchronizes the obtaining of certificates between multiple instances as long as they share the same .casket/acme
folder on disk. This means multiple instances requiring the same certificate will not both request one from the CA, and
they will share the same copy from disk.

## Renewing Certificates

Certificates are only valid for a limited time, so Casket checks each certificate on a regular basis and automatically
renews certificates that expire soon (30 days). If renewal fails, Casket will keep trying.

Once Casket gets the new certificate, it swaps out the old certificate with the new one. This replacement incurs zero
downtime.

As with obtaining certificates, Casket coordinates renewals when used in a cluster, as long as the instances share the
same .casket/acme folder. Only one instance will actually perform the renewal, then the others will reload the updated
certificate.

## Revoking Certificates

Casket does not automatically revoke a certificate, but you can do this with the `-revoke` option, specifying the domain
name. This is only necessary if your site's private key or the certificate authority was compromised. Upon revocation,
Casket deletes the certificate file from disk to prevent it from being used at next run.

## OCSP Stapling

Casket staples OCSP information of all certificates containing an OCSP link to protect the privacy of your sites'
clients and reduce stress on OCSP servers. The cached OCSP status is checked on a regular basis, and if there is a
change, the server will staple the new response.

When new OCSP responses are obtained, Casket persists the staple to disk so that it can weather long OCSP responder
outages. Like certificates, persisted OCSP responses are fully maintained within the .casket folder.

## HTTP Strict Transport Security {#hsts}

HTTP Strict Transport Security (HSTS) is a web security policy mechanism which helps to protect websites against
protocol downgrade attacks and cookie hijacking. Enabling HSTS declares that web browsers should only interact with the
server using a secure HTTPS connection, and never via an insecure HTTP connection. The policy specifies a period of time
during which the server must be accessed in a secure fashion.

Casket does not enable HSTS by default, because if you wish to use the domain without HTTPS, HSTS having been enabled
and remembered by a browser means the browser will not allow connections to your server. HSTS should only be enabled in
a production environment when you know that you won't want to disable HTTPS in the future. It can easily be enabled by
adding the header below to your Casketfile.

``` casketfile
header / Strict-Transport-Security "max-age=31536000;"
```
