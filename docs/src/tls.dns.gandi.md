# tls.dns.gandi

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Gandi. Credentials must be passed
either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/gandi](https://github.com/libdns/gandi) module.

## Environment Variables

- `GANDI_API_KEY` - Gandi API key

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns gandi

  # If no environment variables are set
  dns gandi KEY
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns gandi {
    key YOUR_API_KEY
  }
}
```

- **key** is your Gandi API key.
