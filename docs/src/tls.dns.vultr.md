# tls.dns.vultr

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Vultr. Credentials must be passed either
via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/vultr](https://github.com/libdns/vultr) module.

## Environment Variables

- `VULTR_API_KEY` - Vultr API key

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns vultr

  # If no environment variables are set
  dns vultr KEY
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns vultr {
    key KEY
  }
}
```

- **key** is your Vultr API key.
