# ext

ext allows your site to have clean URLs by assuming a file extension when the path of the URL does not end with one.

An extension in the URL is detected by checking the last element of the path for a dot (`.`).

## Syntax

``` casketfile
ext extensions...
```

-   **extensions...** is a list of space-separated extensions (including the dot) to try. Extensions will be tried in
    the order listed. At least one extension is required.

## Examples

Suppose you have a file called /contact.html. You could serve that file at /contact by having Casket try .html files.

To try .html, .htm, and .php in order:

``` casketfile
ext .html .htm .php
```
