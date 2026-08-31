# browse

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

browse enables directory browsing within the specified base path. It displays a file listing for directories which don't
have an index file in them. In other server software, this is often called indexing.

By default, file listings are disabled and a request to a directory path (where no index file is present) will result in
a 404 for obscurity reasons.

This middleware may set cookies to preserve UI preferences if the user changes them.

## Syntax

``` casketfile
browse [path [tplfile]] {
  path         directory
  tplfile      file
  servearchive [types...]
  buffer       size
}
```

-   **path** is the base path to match. Any directories in this base path become browsable.
-   **tplfile** is the location of a template file to use.
-   **servearchive** <NewInCasket version="v1.2.10" /> enables serving archives of directories. A button will be added
    to the browse page to download the directory as an archive. By default, all archive types are supported. You can
    specify which types to support by adding them as arguments. Supported types are: `zip`, `tar`, `tar.gz`, `tar.xz`,
    `tar.br`, `tar.bz2`, `tar.lz4`, `tar.sz`, `tar.zst`. ([More info](#archives))
-   **buffer** <NewInCasket version="v1.2.10" /> is the size of the buffer used when generating archives. The default is
    10MB. The size values must be positive integers and are interpreted as bytes unless a unit is given. Valid examples:
    `3500` (3,500 bytes), `500kb` (500 kilobytes), `10mb` (10 megabytes), `1gb` (1 gigabyte).

A default template will be used if no template file is specified. Without any arguments, browsing is enabled on the
entire site (`path=/`).

## Template Format

A template is simply an HTML file with *actions* in it. The actions are parsed and executed to display dynamic content.
This directive supports [Casket's template actions](/template-actions) as well as some additional actions specific to
the browse directive. You may use template actions that render [this struct
type](https://github.com/tmpim/casket/blob/5fd2388ac586cc615be4fa8186ba5e4eaf72a03a/caskethttp/browse/browse.go#L100-L134)
(notice that some helper methods are available).

Here is a very simple example template:

``` html
<html>
    <head>
        <title>{{html .Name}}</title>
    </head>
    <body>
        {{if .CanGoUp}}<a href="..">Up one level</a><br>{{end}}
        <h1>{{.Path}}</h1>
        {{range .Items}}
        <a href="{{html .URL}}">{{html .Name}}</a><br>
        {{end}}
    </body>
</html> 
```

... but the default template is nicer.

Notice that the name and URL are sanitized for safe rendering in a browser. Templates are presumed trusted, so if your
file names are not trusted, be sure they are escaped for use in HTML documents.

## JSON Response

You can ask for a JSON representation instead of a browse page by having **application/json** in your **Accept** header:

::: code-group
``` shell [Request]
curl -H "Accept: application/json" 'localhost:2015/?limit=1'
```
:::

::: code-group
``` json [Response]
[
  {
    "IsDir": true,
    "Name": ".git",
    "Size": 476,
    "URL": ".git/",
    "ModTime": "2015-09-11T03:20:09+03:00",
    "Mode": 2147484141
  }
]
```
:::

The above example demonstrates how to ask for JSON, as well as how to limit the number of entries that we want via a
query called **limit**. To yield the whole listing, omit the limit query.

## Archives

<NewInCasket version="v1.2.10" />

If you enable archive serving with the `servearchive` option, a button will be added to the browse page to download the
directory as an archive. By default, all archive types are supported. You can specify which types to support by adding
them as arguments. All archives are created with a top-level directory named after the directory being archived.

#### Archive Types

<div class="tight-list">

Supported types are: 
- `zip` - [ZIP](https://en.wikipedia.org/wiki/ZIP_(file_format)) with [DEFLATE](https://en.wikipedia.org/wiki/Deflate)
  compression (Level 6)
- `tar` - [tar](https://en.wikipedia.org/wiki/Tar_(computing)) with no compression
- `tar.gz` - tar with [Gzip](https://en.wikipedia.org/wiki/Gzip) compression (Level 6)
- `tar.xz` - tar with
  [LZMA](https://en.wikipedia.org/wiki/Lempel%E2%80%93Ziv%E2%80%93Markov_chain_algorithm#xz_and_7z_formats) compression
- `tar.br` - tar with [Brotli](https://en.wikipedia.org/wiki/Brotli) compression (Quality 3)
- `tar.bz2` - tar with [bzip2](https://en.wikipedia.org/wiki/Bzip2) compression (Level 2)
- `tar.lz4` - tar with [LZ4](https://en.wikipedia.org/wiki/LZ4_(compression_algorithm)) compression (Level 1)
- `tar.sz` - tar with [Snappy](https://en.wikipedia.org/wiki/Snappy_(compression)) compression
- `tar.zst` - tar with [Zstandard](https://en.wikipedia.org/wiki/Zstandard) compression (Level 3)

</div>

#### Performance

The archives are generated on-the-fly, so they are always up-to-date. They are not cached by Casket itself, unless you
have another caching proxy in front. The archive is streamed to the client, so the memory overhead is relatively low,
but you can tune the buffer size with the `buffer` option. However, if the archive format is compressed, each archive
request may consume a lot of CPU time. If you are concerned about this, you should only allow the `tar` format.

#### API

Archives for a directory can be downloaded by sending a GET request to the directory path with the `?archive` query
parameter set to the desired archive type. For example, to download a ZIP archive of the current directory, send a GET
request to: `/path/to/dir?archive=zip`.

## Examples

Allow directory listings in all folders that don't have an index file:

``` casketfile
browse
```

Show photo album contents (in /photos) via a custom template:

``` casketfile
browse /photos ../photo_album.tpl
```
