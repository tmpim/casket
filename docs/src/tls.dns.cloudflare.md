# tls.dns.cloudflare

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

::: danger Migration notice
As of Casket 1.4.0, the Cloudflare provider now **only** supports [Scoped 
Tokens](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/). "Global API Keys" are no longer 
supported. The token must have the following permissions:
- **Zone / Zone / Read** - Required to get the Zone ID
- **Zone / DNS / Edit** - Required to edit the DNS record for the ACME DNS-01 challenge
:::

Allows you to obtain certificates using DNS records for domains managed with Cloudflare.

A [Scoped Token](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/) with permissions for the 
zone(s) you want to manage must be provided via environment variables, or directly in the Casketfile. The token must
have the following permissions:
- **Zone / Zone / Read** - Required to get the Zone ID
- **Zone / DNS / Edit** - Required to edit the DNS record for the ACME DNS-01 challenge

## Environment Variables

- `CLOUDFLARE_ZONE_API_TOKEN` - Cloudflare API token

::: tip Note
For compatibility with previous versions of Casket, the follow environment variables are also checked, but it is
recommended to use `CLOUDFLARE_ZONE_API_TOKEN`:

<div class="tight-list">

- `CLOUDFLARE_DNS_API_TOKEN`
- `CF_ZONE_API_TOKEN`
- `CF_DNS_API_TOKEN`

</div>
:::

## Syntax

``` casketfile
tls {
  # If the environment variable is set
  dns cloudflare

  # If no environment variables are set
  dns cloudflare TOKEN
}
```

- **token** is your Cloudflare API token.

<NewInCasket version="v1.4.0" /> Alternative block configuration syntax:

``` casketfile
tls {
  dns cloudflare {
    token YOUR_API_TOKEN
  }
}
```

## Examples

### Usage

Use the email and API token from the `CLOUDFLARE_ZONE_API_TOKEN` environment variable (**recommended**): 

``` casketfile
tls {
  dns cloudflare
}
```

<NewInCasket version="v1.4.0" />

Use the API token directly in the configuration:

``` casketfile
tls {
  dns cloudflare YOUR_API_TOKEN
}
```
