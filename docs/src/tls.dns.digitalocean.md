# tls.dns.digitalocean

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with DigitalOcean. Credentials must be passed
either via environment variables, or directly in the Casketfile.

You can generate a personal access token from the [Applications &
API](https://cloud.digitalocean.com/account/api/tokens) section of the DigitalOcean control panel.

The provider is based on the [libdns/digitalocean](https://github.com/libdns/digitalocean) module.

## Environment Variables

- `DO_AUTH_TOKEN` - DigitalOcean API token

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns digitalocean

  # If no environment variables are set
  dns digitalocean TOKEN
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns digitalocean {
    token YOUR_API_TOKEN
  }
}
```

- **token** is your DigitalOcean API token.
