# tls.dns.ovh

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with OVH. Credentials must be passed either
via environment variables, or directly in the Casketfile.

To use this provider, you need to create [script credentials](https://github.com/ovh/go-ovh#supported-apis) in your
region account and specify the following API rights:

#### For multiple domains

```shell
GET /domain/zone/*/record
POST /domain/zone/*/record
GET /domain/zone/*/record/*
PUT /domain/zone/*/record/*
DELETE /domain/zone/*/record/*
GET /domain/zone/*/soa
POST /domain/zone/*/refresh
```

#### For a single domain or delegation

```shell
GET /domain/zone/yourdomain.com/record
POST /domain/zone/yourdomain.com/record
GET /domain/zone/yourdomain.com/record/*
PUT /domain/zone/yourdomain.com/record/*
DELETE /domain/zone/yourdomain.com/record/*
GET /domain/zone/yourdomain.com/soa
POST /domain/zone/yourdomain.com/refresh
```

The provider is based on the [libdns/ovh](https://github.com/libdns/ovh) module.

## Environment Variables

- `OVH_ENDPOINT` - OVH API endpoint
- `OVH_APPLICATION_KEY` - OVH application key
- `OVH_APPLICATION_SECRET` - OVH application secret
- `OVH_CONSUMER_KEY` - OVH consumer key

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns ovh

  # If no environment variables are set
  dns ovh ENDPOINT APPLICATION_KEY APPLICATION_SECRET CONSUMER_KEY
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns ovh {
    endpoint           YOUR_ENDPOINT
    application_key    YOUR_APPLICATION_KEY
    application_secret YOUR_APPLICATION_SECRET
    consumer_key       YOUR_CONSUMER_KEY
  }
}
```

- **endpoint** is the OVH API endpoint.
- **application_key** is the OVH application key.
- **application_secret** is the OVH application secret.
- **consumer_key** is the OVH consumer key.
