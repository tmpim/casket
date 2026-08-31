# tls.dns.namedotcom

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Name.com. Credentials must be passed either
via environment variables, or directly in the Casketfile.

Your Name.com API token can be found on your [API Token Management](https://www.name.com/account/settings/api) page.

The provider is based on the [libdns/namedotcom](https://github.com/libdns/namedotcom) module.

## Environment Variables

- `NAMECOM_API_USER` - Name.com username
- `NAMECOM_API_TOKEN` - Name.com API token
- `NAMECOM_API_SERVER` - (*optional*) URL of the API server to use, defaults to `https://api.name.com`. The sandbox
  environment can be used by setting this to `https://api.dev.name.com`.

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns namedotcom

  # If no environment variables are set
  dns namedotcom USERNAME TOKEN
  
  # If you want to specify the server
  dns namedotcom USERNAME TOKEN SERVER
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns namedotcom {
    username YOUR_USERNAME
    token    YOUR_API_TOKEN
    
    # Specify to use the sandbox environment
    server https://api.dev.name.com
  }
}
```

- **username** is your Name.com username.
- **api_token** is your Name.com API token.
- **server** is the URL of the API server to use, defaults to `https://api.name.com`. The sandbox environment can be used
  by setting this to `https://api.dev.name.com`.
