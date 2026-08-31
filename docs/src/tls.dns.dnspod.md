# tls.dns.dnspod

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with DNSPod. Credentials must be passed
either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/dnspod](https://github.com/libdns/dnspod) module.

## Environment Variables

- `DNSPOD_API_KEY` - DNSPod API key

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns dnspod

  # If no environment variables are set
  dns dnspod KEY
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns dnspod {
    key YOUR_API_KEY
  }
}
```

- **key** is your DNSPod API key.
