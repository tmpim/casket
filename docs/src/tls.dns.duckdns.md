# tls.dns.duckdns

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Duck DNS. Credentials must be passed
either via environment variables, or directly in the Casketfile.

Your Duck DNS API token can be found on your account page.

The provider is based on the [libdns/duckdns](https://github.com/libdns/duckdns) module.

## Environment Variables

- `DUCKDNS_TOKEN` - Duck DNS API token

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns duckdns

  # If no environment variables are set
  dns duckdns TOKEN
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns duckdns {
    token YOUR_API_TOKEN
  }
}
```

- **token** is your Duck DNS API token.
