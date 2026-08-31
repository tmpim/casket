# root

root simply specifies the root of the site. This is very useful, in fact required, if the root (/) directory of the
website is not the same as where Casket is being executed from.

The default root is the current working directory. A relative root path is relative to the current working directory.

## Syntax

``` casketfile
root path
```

-   **path** is the directory to set as the site root

## Examples

Serve the website out of Jake's public_html folder instead of the current working directory:

``` casketfile
root /home/jake/public_html
```
