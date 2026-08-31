# basicauth

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

basicauth implements HTTP Basic Authentication. Basic Authentication can be used to protect directories and files with a
username and password. Note that basic auth is *not secure* over plain HTTP. Use discretion when deciding what to
protect with HTTP Basic Authentication.

When a user requests a resource that is protected, the browser will prompt the user for a username and password if they
have not already supplied one. If the proper credentials are present in the Authorization header, the server will grant
access to the resource and set the `{user}` [placeholder](/placeholders) to the value of the username. If the header is
missing or the credentials are incorrect, the server will respond with HTTP 401 Unauthorized.

This directive allows use of .htpasswd files by prefixing the password argument with `htpasswd=` and the path to the
.htpasswd file to use. 

::: warning
Support for .htpasswd is for legacy sites only and may be removed in the future; do not use .htpasswd with new sites.
:::

Note that basicauth does not protect `OPTIONS` requests. This is so that clients can know whether to prompt for
authentication in cross-origin contexts. We emphasize that servers violate RFC 2616 if they respond to OPTIONS with a
resource retrieval (i.e. if they respond with secret content to OPTIONS requests).

## Syntax

``` casketfile
basicauth path username password {
  exclude [paths...]
  cookie [{
    ttl [duration]
    refresh [duration]
    name [name]
  }]
}
```

-   **path** is the file or directory to protect
-   **username** is the username
-   **password** is the password
-   **exclude** <NewInCasket version="v1.2.10" /> is a list of paths to exclude from protection. This is useful if you
    want to protect a directory but allow access to a specific file within it. Paths are relative to the protected path.
    For example, if you want to protect `/secret` but allow access to `/secret/allowed.txt`, you would use `exclude
    allowed.txt`.
-   **cookie** <NewInCasket version="v1.5.0" /> allows storing the basicauth header in a cookie, to allow the user to
    stay authenticated longer than their current browser session. See [Cookies](#cookies) below for more information.
    -   **ttl** is the Max-Age of the cookie to set. Note that some browsers have their own maximum, typically 400 days
        but may be less depending on the user's configuration. Syntax [here](<https://pkg.go.dev/time#ParseDuration>). 
        Default is `24h`.
    -   **refresh** is the threshold, if reached, that the cookie will be refreshed and reset to `ttl`. 
        Set to `0` to disable refreshing. Syntax [here](<https://pkg.go.dev/time#ParseDuration>). 
        Default is half of `ttl`.
    -   **name** is the name of the cookie. An additional cookie with the `_ts` suffix will also be set. 
        Default is `casket_basicauth`.

This syntax is convenient for protecting a single file or base path/directory with the default realm "Restricted". To
protect multiple resources or to specify a realm, use the following variation:

``` casketfile
basicauth username password {
  realm name
  resources...
  exclude [paths...]
  cookie [{
    ttl [duration]
    refresh [duration]
    name [name]
  }]
}
```

-   **username** is the username.
-   **password** is the password.
-   **realm** identifies the protection partition; it is optional and cannot be repeated. Realms are used to specify the
    space in which the protection applies. This can be convenient for user agents that are configured to remember
    authentication details (which is most browsers).
-   **resources** is a list of files/directories to protect, one per line.
-   **exclude** <NewInCasket version="v1.2.10" /> is a list of paths to exclude from protection. This is useful if you
    want to protect a directory but allow access to a specific file within it. Paths are relative to the protected path.
    For example, if you want to protect `/secret` but allow access to `/secret/allowed.txt`, you would use `exclude
    allowed.txt`.
-   **cookie** <NewInCasket version="v1.5.0" /> allows storing the basicauth header in a cookie, to allow the user to
    stay authenticated longer than their current browser session. See [Cookies](#cookies) below for more information.
    -   **ttl** is the Max-Age of the cookie to set. Note that some browsers have their own maximum, typically 400 days
        but may be less depending on the user's configuration. Syntax [here](<https://pkg.go.dev/time#ParseDuration>). 
        Default is `24h`.
    -   **refresh** is the threshold, if reached, that the cookie will be refreshed and reset to `ttl`. 
        Set to `0` to disable refreshing. Syntax [here](<https://pkg.go.dev/time#ParseDuration>). 
        Default is half of `ttl`.
    -   **name** is the name of the cookie. An additional cookie with the `_ts` suffix will also be set. 
        Default is `casket_basicauth`.

## Examples

Protect all files in /secret so only Bob can access them with the password "hiccup":

``` casketfile
basicauth /secret Bob hiccup
```

Protect multiple files and directories in the realm "Mary Lou's documents" so Mary Lou has access with her password
"milkshakes":

``` casketfile
basicauth "Mary Lou" milkshakes {
  realm "Mary Lou's documents"
  /notes-for-mary-lou.txt
  /marylou-files
  /another-file.txt
}
```

Protect all files in the root (`/`) except for `/robots.txt` and everything in `/public`:

``` casketfile
basicauth / Bob hiccup {
  exclude /robots.txt /public
}  
```

Store the authentication in a cookie for 24 hours (and refreshes after 12 days):

``` casketfile
basicauth / Bob hiccup {
  cookie
}
```

Store the authentication in a cookie for 365 days (and refreshes after 182.5 days):

``` casketfile
basicauth / Bob hiccup {
  cookie {
    ttl 8760h
  }
}
```

## Cookies

<NewInCasket version="v1.5.0" />

By adding the `cookie` parameter or block in the `basicauth` directive, you can opt-in to storing a cookie when the user
authenticates with HTTP Basic Authentication. After a successful authentication, the `casket_basicauth` header will be
set to the same value as the `Authentication` header (in **plaintext**), and an additional `casket_basicauth_ts` header
will be stored with the current timestamp. Future requests will read this cookie and pass it through to the 
`Authentication` (if it was not already set), allowing the user to remain authenticated for longer than their browser
session would normally allow for HTTP Basic Authentication. The modified `Authentication` header is passed upstream to 
proxies.

When the `refresh` threshold is reached and another request comes in, the cookie's Max-Age will be reset to the `ttl` 
parameter. If `refresh` is set to 0, the cookie will only be set once, and the user will have to re-authenticate when it
expires.

If the server's `basicauth` configuration changes, for example changing a username or password, then the authentication
via cookie will no longer succeed. If authentication fails, the cookie will be removed on the next request.

::: danger
The cookie is stored on the user's device in plaintext. This option further reduces the already low security of HTTP 
Basic Authentication. If a user's browser or device is compromised, the cookies can be extracted and the password 
retrieved. Furthermore, Casket does not enforce the TTL or Refresh values, so a savvy user could tamper with them to
keep their session alive for longer than you intend. This feature is intended to be used in closed environments with 
HTTPS. If you need security, consider proxying to alternative authentication providers.
:::

