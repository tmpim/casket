# request_id

request_id generates a UUID which can be referenced via the `{request_id}` [placeholder](/placeholders), for use in many
other directives including [`header`](/header) and [`log`](/log)

When request_id is used, the `{request_id}` [placeholder](/placeholders) will have a value; otherwise it will be empty.

## Syntax

``` casketfile
request_id [header_field]
```

-   **header_field** is an optional header field name from which to read an existing request ID. If this field name is
    set and the field contains a valid UUID, it will be used and a new one will not be created.

    ::: warning
    For the sake of your site's security, do NOT abuse this feature (i.e. don't use it to track user sessions).
    :::

## Examples

Set the request ID in a response header:

``` casketfile
request_id 
header /  Casket-Request-Id {request_id}
```

Read the request ID from a request header called X-Request-ID, if present:

``` casketfile
request_id X-Request-ID
```
