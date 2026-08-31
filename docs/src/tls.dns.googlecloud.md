# tls.dns.googlecloud

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Google Cloud DNS. Credentials must be
passed either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/googlecloud](https://github.com/libdns/googlecloud) module.

## Environment Variables

- `GCE_PROJECT` - Google Cloud project ID
- `GCE_SERVICE_ACCOUNT` - Path to the service account JSON file

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns googlecloud

  # If GCE_SERVICE_ACCOUNT is set in the environment
  dns googlecloud PROJECT_ID
  
  # If no environment variables are set
  dns googlecloud PROJECT_ID SERVICE_ACCOUNT_PATH
}
```

- **PROJECT_ID** is your Google Cloud project ID.
- **SERVICE_ACCOUNT_PATH** is the path to the service account JSON file.

<br />

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns googlecloud {
    project         YOUR_PROJECT_ID
    service_account PATH_TO_SERVICE_ACCOUNT_JSON
  }
}
```

- **project** is your Google Cloud project ID.
- **service_account** is the path to the service account JSON file.
 
