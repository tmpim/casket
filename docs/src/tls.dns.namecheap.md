# tls.dns.namecheap

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with NameCheap. Credentials must be passed
either via environment variables, or directly in the Casketfile.

Instructions on how to obtain your NameCheap API key can be found [here](https://www.namecheap.com/support/api/intro/).

The provider is based on the [libdns/namecheap](https://github.com/libdns/namecheap) module.

## Environment Variables

- `NAMECHEAP_API_USER` - NameCheap username
- `NAMECHEAP_API_KEY` - NameCheap API key
- `NAMECHEAP_API_SANDBOX` - (*optional*) Set to `true` to enable the sandbox environment

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns namecheap

  # If no environment variables are set
  dns namecheap USER KEY
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns namecheap {
    username YOUR_USERNAME
    api_key  YOUR_API_KEY
    
    # Specify to enable the sandbox environment
    sandbox
  }
}
```

- **username** is your NameCheap username.
- **api_key** is your NameCheap API key.
- **sandbox** is an optional parameter to enable the sandbox environment.
