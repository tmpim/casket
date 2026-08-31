# rewrite

rewrite performs internal URL rewriting. This allows the client to request one resource but actually be served another
without an HTTP redirect. Rewrites are invisible to the client.

There are simple rewrites (fast) and complex rewrites (slower), but they're powerful enough to accommodate most dynamic
back-end applications.

## Syntax

``` casketfile
rewrite [not] from to...
```

-   **not** if specified, will invert the pattern and matching logic
-   **from** is a regular expression to match
-   **to** is a space-separated list of destination paths to rewrite to (the resource to respond with); it will try each
    destination in sequence until the first one that exists on disk in the site root

Advanced users may open a block and make a complex rewrite rule:

``` casketfile
rewrite [basepath] {
  regexp pattern
  ext    extensions...
  if     a cond b
  if_op  [and|or]
  to     destinations...
}
```

-   **basepath** is the base path to match before rewriting with regular expression. Default is /.
-   **regexp** (shorthand: **r**) will match the path with the given regular expression **pattern**.
    ::: warning
    Extremely high-load servers should avoid using regular expressions.
    :::
-   **extensions...** is a space-separated list of file extensions (prepended with a dot) to include or ignore. Prefix
    an extension with `!` to exclude an extension. The forward slash `/` symbol matches paths without file extensions.
-   **if** specifies a rewrite condition. Multiple ifs are AND-ed together. **a** and **b** are any string and may use
    [request placeholders](/placeholders). **cond** is the condition, with possible values explained below.
-   **if_op** specifies how the ifs are evaluated; the default is `and`.
-   **destinations...** is one or more space-separated paths to rewrite to, with support for [request
    placeholders](/placeholders) as well as numbered regular expression captures such as {1}, {2}, etc. Rewrite will
    check each destination in order and rewrite to the first destination that exists. Each one is checked as a file or,
    if ends with /, as a directory. The last destination will act as default if no other destination exists.

## "if" conditions

The if keyword is a powerful way to describe your rule. It takes the format `a cond b`, where the values `a` and `b` are
separated by `cond`, a condition. The condition can be any of these:

-   `is` = a equals b
-   `not` = a does NOT equal b
-   `has` = a has b as a substring (b is a substring of a)
-   `not_has` = b is NOT a substring of a
-   `starts_with` = b is a prefix of a
-   `not_starts_with` = b is NOT a prefix of a
-   `ends_with` = b is a suffix of a
-   `not_ends_with` = b is NOT a suffix of a
-   `match` = a matches b, where b is a regular expression
-   `not_match` = a does NOT match b, where b is a regular expression

Note: As a general rule, you can negate any condition `cond` by prefixing it with `not_`.

## Examples

Rewrite everything to /index.php. (`rewrite / /index.php` will only match /)

``` casketfile
rewrite .* /index.php
```

When requests come in for /mobile, actually serve /mobile/index.

``` casketfile
rewrite /mobile /mobile/index
```

If the file is not favicon.ico and it is not a valid file or directory, serve the maintenance page if present, or
finally, rewrite to index.php.

``` casketfile
rewrite {
  if {file} not favicon.ico
  to {path} {path}/ /maintenance.html /index.php
}
```

If user agent includes "mobile" and path is not a valid file/directory, rewrite to the mobile index page.

``` casketfile
rewrite {
  if {>User-agent} has mobile
  to {path} {path}/ /mobile/index.php
}
```

Rewrite /app to /index with a query string. `{1}` is the matched group `(.*)`.

``` casketfile
rewrite /app {
  r  (.*)
  to /index?path={1}
}
```

Rewrite requests for /app/example to /index.php?category=example.

``` casketfile
rewrite /app {
  r  ^/(\w+)/?$
  to /index?category={1}
}
```
