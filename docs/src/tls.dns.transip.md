# tls.dns.transip

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with TransIP. Credentials must be passed
either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/transip](https://github.com/libdns/transip) module.

## Environment Variables

- `TRANSIP_ACCOUNT_NAME` - the account name for your TransIP account
- `TRANSIP_PRIVATE_KEY_PATH` - the path to the private key file

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns transip

  # If no environment variables are set
  dns transip ACCOUNT_NAME PRIVATE_KEY_PATH
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns transip {
    account_name     YOUR_ACCOUNT_NAME
    private_key_path YOUR_PRIVATE_KEY_PATH
  }
}
```

- **account_name** (required) - the account name for your TransIP account
- **private_key_path** (required) - the path to the private key file
