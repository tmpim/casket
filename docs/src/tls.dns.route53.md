# tls.dns.route53

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

::: danger Migration notice
As of Casket 1.4.0, the code used for the `tls.dns.route53` directive has been updated to use the `libdns/route53` 
module. We have attempted to preserve the functionality of the previous implementation, but due to limited resources,
we have not been able to test it. If you encounter any issues, please report them to [the issue
tracker](https://github.com/tmpim/casket/issues/new) and we will do our best to accommodate your use case.
:::

Allows you to obtain certificates using DNS records for domains managed with Amazon Route 53. Credentials must be
passed either via environment variables, or directly in the Casketfile.

The provider is based on the [libdns/route53](https://github.com/libdns/route53) module.

## Environment Variables

- `AWS_REGION` - the AWS region to send the request to (e.g. `us-west-2`)
- `AWS_ACCESS_KEY_ID` - the AWS access key ID
- `AWS_SECRET_ACCESS_KEY` - the AWS secret access key
- `AWS_MAX_RETRIES` - (*optional*) the maximum number of retries for failed requests. Defaults to `5`.
- `AWS_PROFILE` - (*optional*) the AWS profile to use. If set, this will override the `AWS_ACCESS_KEY_ID` and
  `AWS_SECRET_ACCESS_KEY` environment variables.

## Syntax

Shorthand configuration syntax:

``` casketfile
tls {
  # If all the environment variables are set
  dns route53
}
```

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns route53 {
    region            AWS_REGION
    access_key_id     AWS_ACCESS_KEY_ID
    secret_access_key AWS_SECRET_ACCESS_KEY
    max_retries       AWS_MAX_RETRIES
    aws_profile       AWS_PROFILE
  }
}
```

- **region** (required) - the AWS region to send the request to (e.g. `us-west-2`).
- **access_key_id** (required) - the AWS access key ID.
- **secret_access_key** (required) - the AWS secret access key.
- **max_retries** (optional) - the maximum number of retries for failed requests. Defaults to `5`.
- **aws_profile** (optional) - the AWS profile to use. If set, this will override the `AWS_ACCESS_KEY_ID` and
  `AWS_SECRET_ACCESS_KEY` environment variables.
