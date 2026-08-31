# tls.dns.pdns (PowerDNS)

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with PowerDNS. Credentials must be passed
either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/powerdns](https://github.com/libdns/powerdns) module.

## Environment Variables

- `PDNS_API_URL` - the URL of your PowerDNS server
- `PDNS_API_KEY` - the API key for your PowerDNS server
- `PDNS_SERVER_ID` - the server ID for your PowerDNS server (defaults to `localhost`)

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If the environment variable is set
  dns pdns

  # If no environment variables are set
  dns pdns API_URL API_KEY
  
  # If you have a server ID (it defaults to localhost)
  dns pdns API_URL API_KEY SERVER_ID
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns pdns {
    api_url     YOUR_API_URL
    api_key     YOUR_API_KEY
    server_id   YOUR_SERVER_ID # Optional, defaults to `localhost`
  }
}
```

- **api_url** (required) - the URL of your PowerDNS server
- **api_key** (required) - the API key for your PowerDNS server
- **server_id** (optional) - the server ID for your PowerDNS server (defaults to `localhost`)
