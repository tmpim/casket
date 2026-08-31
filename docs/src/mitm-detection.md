# Detecting HTTPS Interception

Casket has the ability to detect certain Man-in-the-Middle (MITM) attacks on HTTPS connections that may otherwise be
invisible to the browser and the end user. This means Casket can determine whether it is "likely" or "unlikely" that a
TLS proxy is actively intercepting an HTTPS connection.

All incoming HTTPS connections are automatically checked for tampering using techniques described by Durumeric,
Halderman, et. al. in their [NDSS '17 paper](https://jhalderm.com/pub/papers/interception-ndss17.pdf). The results of
the inspection are exposed in various ways throughout Casket so you can choose how to handle a suspected MITM attack on
your clients. (Keep in mind that many TLS proxies take the form of "benevolent" antivirus or firewall products.)

TLS connections that are being intercepted are NOT secure, despite software vendor advertisements to the contrary. The
degree to which you respond to a suspected MITM attack is up to you and depends on the nature of your site, your
audience, and political circumstances. You might respond to suspected HTTPS interception in any of these ways (in order
of increasing ostentatiousness):

-   [Log](/log) that it happened using the `{mitm}` [placeholder](/placeholders) and a custom log format.
-   If proxying to an upstream application, [add a request header](/proxy) using header_upstream in the `proxy`
    directive with a value that has `{mitm}`.
-   If you're using [templates](/templates), display a warning to the user on your website using the 
    <code v-pre>{{.IsMITM}}</code> [template action](/template-actions).
-   Perform an [internal rewrite](/rewrite) of the URI to a dedicated warning, error, or informational page.
-   [Redirect](/redir) the user to another page or site.
-   Close the connection immediately with an empty response (coming soon, maybe).

Please read this entire page before implementing any measures that might be considered extreme.

## Disclaimer

The Casket authors, maintainers, and contributors will make a good-faith attempt to keep this feature working correctly
with commonly-used versions of mainstream browsers, but cannot guarantee perfect accuracy. This feature relies on
hard-coded heuristics that attempt to identify browsers from the TLS handshake. Browser and OS updates may render the
heuristics obsolete at any time. The Casket developers are not responsible for any damages, costs, miscommunications, or
misunderstandings, or other consequences that may result from using this feature. Use with wisdom and at your own risk.

## Supported Clients

Casket is programmed to protect recent versions of Chrome, Firefox, IE/Edge, and Safari. Bleeding-edge development
versions of these browsers may not yet be recognized (let us know if they're not!). We also experimentally attempt to
recognize and support the Tor browser.

## False Positives

Casket may, on occasion, incorrectly flag a connection as "likely" intercepted even if it is not. This usually happens
when clients spoof their User-Agent string. For the best possible protection, we recommend that users do not change
their User-Agent header and that site owners keep Casket updated.

It is also possible that there is a browser/platform combination that is not yet considered. To report a false positive,
please [file an issue](https://github.com/tmpim/casket/issues/new) with your real, unmodified User-Agent string,
browser version, OS/platform details, the raw ClientHello bytes, and any other relevant build information. You must also
be certain that your connection was made on a trusted network that is NOT being firewalled or proxied and that all OS
"security" products are completely disabled on your machine and the local network. (You *must* convince us that the
connection was actually secure, and we have to be able to reproduce your report.)

## False Negatives

When HTTPS interception is occurring but Casket is not able to detect it (an "unlikely" classification), there could be
a number of reasons. From least to most ominous:

-   It could simply be that Casket's detection heuristics are not comprehensive enough. Please [file an
    issue](https://github.com/tmpim/casket/issues/new) as long as you can provide enough information to catch the
    interception, including the ClientHello bytes and as many details about the MITM software that are known.
-   The client is not sending a recognized User-Agent header. Except for a few limited exceptions, Casket only checks
    for MITM with major browsers.
-   The TLS proxy is preserving the original properties of the TLS handshake between the client and the server. This
    scenario is not the worst false negative because at least the browser will be able to show a warning if the TLS
    connection is weak.
-   A TLS proxy is modifying or stripping the User-Agent header, perhaps in an attempt to hide. However, any TLS proxy
    that modifies HTTP requests runs the risk of breaking HTTP, which would expose its presence (similar to how gravity
    exposes black holes).

Casket's MITM detection features work mainly because TLS proxies are implemented carelessly, documented poorly, and
updated sparingly.
