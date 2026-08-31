# gzip

gzip enables gzip compression if the client supports it. By default, responses are not gzipped. If enabled, the default
settings will ensure that images, videos, and archives (already compressed) are not gzipped.

Note that, even without the gzip directive, Casket will serve .gz (gzip) or .br (brotli) compressed files if they
already exist on disk and the client supports that encoding.

## Syntax

``` casketfile
gzip
```

The plain gzip config is good enough for most things, but you can gain more control if needed:

``` casketfile
gzip {
    ext        extensions...
    not        paths
    level      compression_level
    min_length min_bytes
}
```

-   **extensions...** is a space-separated list of file extensions to compress. Supports wildcard `*` to match all
    extensions.
-   **paths** is a space-separated list of paths in which *not* to compress.
-   **compression_level** is a number from 1 (best speed) to 9 (best compression). Default is 6.
-   **min_bytes** is the minimum number of bytes in a response needed before compression will happen. Default is no
    minimum length.

## Examples

Enable gzip compression:

``` casketfile
gzip
```

Enable very fast but minimal compression except in the /images and /videos folders (note, however, that images and
videos will not be gzipped anyway):

``` casketfile
gzip {
  level 1
  not   /images /videos
}
```
