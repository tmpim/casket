# tls.dns.acmedns

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

::: danger Migration notice
As of Casket 1.4.0, the code used for the `tls.dns.acmedns` directive has been updated to use the `libdns/acmedns`
module. We have attempted to preserve the functionality of the previous implementation, but due to limited resources,
we have not been able to test it. If you encounter any issues, please report them to [the issue
tracker](https://github.com/tmpim/casket/issues/new) and we will do our best to accommodate your use case.
:::

Allows you to obtain certificates using DNS records for any domain registrar, by using an 
[ACME-DNS](https://github.com/joohoi/acme-dns) server. This allows you to solve ACME DNS-01 challenges on any provider,
even if they do not have their own API or a Casket DNS provider module, and without having to store your account
credentials or API keys on your web server. ACME-DNS credentials must be passed either via environment variables, or 
directly in the Casketfile.

The provider is based on the [libdns/acmedns](https://github.com/libdns/acmedns) module.

::: warning
The steps below are based on the documentation for the Caddy v2 [acmedns](https://github.com/caddy-dns/acmedns)
provider. It has not actually been tested with Casket yet. Please report any issues to the [issue
tracker](https://github.com/tmpim/casket/issues/new).
:::

## Prerequisites

- A domain name you control, for example `your-domain.example.com`. You'll need to be able to create a CNAME record with
  the name `_acme-challenge.your-domain.example.com`.
- Access to an ACME-DNS server. For testing purposes, you can use the public server at <https://auth.acme-dns.io>.
  However, self-hosting is highly encouraged. To learn how to self-host an ACME-DNS server, refer to the
  [ACME-DNS documentation](https://github.com/joohoi/acme-dns#self-hosted).

## Setup

1. [Register an account](https://github.com/joohoi/acme-dns#register-endpoint) on your ACME-DNS server by making a POST
   request to the `/register` endpoint. For example:

   ``` shell
    curl -X POST https://auth.acme-dns.io/register
    ```
   
    The server will respond with a JSON object containing your account credentials, similar to this one:

    ``` json
    {
      "username": "5d26a340-2e1d-4b6b-af1a-4aab569897b7",
      "password": "_r2gFOVtrF9I82_l6ZXTfPaxCgldqJSWaTmd4BS9",
      "fulldomain": "37c51280-79ca-435f-af32-c775eb67e2ab.auth.acme-dns.io",
      "subdomain": "37c51280-79ca-435f-af32-c775eb67e2ab",
      "allowfrom": []
    }
    ```

2. Create a CNAME record for the `_acme-challenge` subdomain of your domain, pointing to the `fulldomain` value from the
   previous step. For example, if your `fulldomain` is `37c51280-79ca-435f-af32-c775eb67e2ab.auth.acme-dns.io`, you
   would create a CNAME record for `_acme-challenge.your-domain.example.com` pointing to
   `37c51280-79ca-435f-af32-c775eb67e2ab.auth.acme-dns.io`:

    ```
    _acme-challenge.your-domain.example.com. 300 IN CNAME 37c51280-79ca-435f-af32-c775eb67e2ab.auth.acme-dns.io.
    ```
   
3. Use the credentials obtained in step 1 to configure the `acmedns` plugin (block configuration syntax), or save
   them to a file and set the `ACME_DNS_STORAGE_PATH` environment variable (shorthand configuration syntax).

## Environment Variables

- `ACME_DNS_API_BASE` - the base URL of your ACME-DNS server
- `ACME_DNS_STORAGE_PATH` - the path to the JSON account data file

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns acmedns
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns acmedns {
    server_url url
    stoarge    path
    username   username
    password   password
    subdomain  subdomain
  }
}
```

- **server_url** (string) - the base URL of your ACME-DNS server
- **storage** (string) - the path to the JSON account data file
- **username** (string) - the username to use for the ACME-DNS server
- **password** (string) - the password to use for the ACME-DNS server
- **subdomain** (string) - the subdomain to use for the ACME-DNS server
