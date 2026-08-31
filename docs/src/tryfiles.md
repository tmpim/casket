# tryfiles

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

<NewInCasket />

tryfiles is a directive that allows you to specify a list of files to try to serve from the filesystem defined by the
[root](/root) directive. If none of the files exist, the request is passed on to the next handler in the chain, such as
a [proxy](/proxy) handler.

This directive makes it easier to serve compiled static content (React, Vue, etc.) for single page applications. The
default configuration will first try to serve the requested file, then try to serve an index page (e.g. `index.html`) if
the requested file does not exist.

## Syntax

``` casketfile
# Uses a sensible default configuration, works for most SPAs (see below)
tryfiles
```

This default configuration is equivalent to:

``` casketfile
tryfiles {path} index.html index.htm index.txt default.html default.htm default.txt {
  except /.well-known {cfg.addr.path}/.well-known
  without {cfg.addr.path}
}
```

This default configuration even works if the site block is defined on a subpath, e.g. `example.com/subpath`. 

You can customize the list of files to try:

``` casketfile
tryfiles [paths...] {
  except  paths...
  without prefix
}
```

-   **paths** is a list of space-separated file paths to try in order. [Placeholders](#placeholders) are supported.
-   **except** is a list of space-separated file paths to exclude from the list of files to try.
    [Placeholders](#placeholders) are supported.
-   **without** is a path prefix to strip from the request path before trying to find the files on the filesystem.

## Examples

Serve `index.html` if the requested file does not exist:

``` casketfile
tryfiles
```

Try the requested file, then `index.php` with the query string appended:

``` casketfile
tryfiles {path} /index.php?{query}&p={path}
```
