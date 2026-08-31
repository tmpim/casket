# status

status writes a status code to the response. It does not write a response body.

## Syntax

``` casketfile
status code path
```

-   **code** is the HTTP status code to respond with (must be numeric).
-   **path** is the base request path to match.

If you have a lot of status rewrites to group, share a status code by making a table:

``` casketfile
status code {
  path
}
```

Each line describes a base path which should have that status code written.

## Examples

For a catch-all status:

``` casketfile
status 404 /
```

To hide the existence of a secret folder:

``` casketfile
status 404 /secrets
```

To hide the existence of multiple folders:

``` casketfile
status 404 {
  /hidden
  /secrets
}
```
