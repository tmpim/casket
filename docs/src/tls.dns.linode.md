# tls.dns.linode

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Linode. Credentials must be passed
either via environment variables, or directly in the Casketfile.

Instructions on how to generate an API token can be found in the [Linode API
documentation](https://www.linode.com/docs/products/tools/api/guides/manage-api-tokens/#create-an-api-token).

The provider is based on the [libdns/linode](https://github.com/libdns/linode) module.

## Environment Variables

- `LINODE_AUTH_TOKEN` - Linode v4 API token

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns linode

  # If no environment variables are set
  dns linode TOKEN
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns linode {
    token YOUR_API_TOKEN
  }
}
```

- **token** is your Linode v4 API token.
