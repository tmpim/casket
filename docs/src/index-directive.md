# index

index sets the list of file names that are used as "index" files. When a directory path is requested instead of a
specific file, the directory is checked for existing index files. The first matching file name is served.

The default index files are, in this order:

1.  index.html
2.  index.htm
3.  index.txt
4.  default.html
5.  default.htm
6.  default.txt

Using this directive overwrites the list, it does not append to it.

## Syntax

``` casketfile
index filenames...
```

-   **filenames...** is a list of space-separated filenames to try as indexes. At least one name is required.

## Examples

Use only goaway.png and easteregg.html as index files:

``` casketfile
index goaway.png easteregg.html
```
