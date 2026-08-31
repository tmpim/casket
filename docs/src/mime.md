^ http.mime

mime sets the Content-Type in a response based on the file extension in the request.

Normally, Content-Type is detected automatically for static files by sniffing the content, but this is not always
possible. If you encounter responses with the wrong Content-Type or are serving content other than static files, you can
use this middleware to set the right Content-Type.

## Syntax

``` casketfile
mime ext type
```

-   **ext** is the file extension to match, including the dot prefix.
-   **type** is the Content-Type

If you have a lot of MIME types to set, open a block:

``` casketfile
mime {
  ext type
}
```

Each line defines a MIME extension-type pair. You can have as many lines as you need in a mime block.

## Examples

Customize the Content-Type of Flash files:

``` casketfile
mime .swf application/x-shockwave-flash
```

For multiple files:

``` casketfile
mime {
  .swf application/x-shockwave-flash
  .pdf application/pdf
}
```
