# tls.dns.rfc2136

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using [RFC 2136](https://datatracker.ietf.org/doc/html/rfc2136) dynamic updates.
Credentials must be passed either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/rfc2136](https://github.com/libdns/rfc2136) module.

## Environment Variables

- `RFC2136_NAMESERVER` - the nameserver to send the request to (including port, e.g. `1.2.3.4:53`)
- `RFC2136_TSIG_KEY` - the name of the [TSIG key](https://en.wikipedia.org/wiki/TSIG) used to authenticate the request
- `RFC2136_TSIG_SECRET` - the secret of the TSIG key
- `RFC2136_TSIG_ALGORITHM` - the algorithm of the TSIG key. Must be one of: `hmac-sha1`, `hmac-sha224`, `hmac-sha256`, 
  `hmac-sha384`, `hmac-sha512`. MD5 is not supported.

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns rfc2136

  # If no environment variables are set
  dns rfc2136 NAMESERVER TSIG_ALGORITHM TSIG_KEY TSIG_SECRET
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns rfc2136 {
    nameserver NAMESERVER
    key        TSIG_KEY
    secret     TSIG_SECRET
    algorithm  TSIG_ALGORITHM
  }
}
```

- **nameserver** (required) - the nameserver to send the request to (including port, e.g. `1.2.3.4:53`).
- **key** (required) - the name of the TSIG key used to authenticate the request.
- **secret** (required) - the secret of the TSIG key.
- **algorithm** (required) - the algorithm of the TSIG key. Must be one of: `hmac-sha1`, `hmac-sha224`, `hmac-sha256`,
  `hmac-sha384`, `hmac-sha512`. MD5 is not supported.

## Examples

### Usage

Use the configuration from the environment variables (**recommended**):

``` casketfile
tls {
  dns rfc2136
}
```

Use a specific nameserver and TSIG key:

``` casketfile
tls {
  dns rfc2136 {
    nameserver 1.2.3.4:53
    key        test
    secret     cWnu6Ju9zOki4f7Q+da2KKGo0KOXbCf6Pej6hW3geC4=
    algorithm  hmac-sha256
  }
}
```
