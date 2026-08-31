# tls.dns.godaddy

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with GoDaddy. Credentials must be passed
either via environment variables, or directly in the Casketfile.

Your GoDaddy API key and secret can be found on your [API Access](https://developer.godaddy.com/keys) page.

The provider is based on the [libdns/godaddy](https://github.com/libdns/godaddy) module.

## Environment Variables

- `GODADDY_API_KEY` - GoDaddy API key
- `GODADDY_API_SECRET` - GoDaddy API secret

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns godaddy

  # If no environment variables are set
  dns godaddy KEY SECRET
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns godaddy {
    key    YOUR_API_KEY
    secret YOUR_API_SECRET
  }
}
```

- **key** is your GoDaddy API key.
- **secret** is your GoDaddy API secret.
