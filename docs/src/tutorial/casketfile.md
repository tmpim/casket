# Casketfile Primer

This tutorial will show you how easy it is to configure Casket with the Casketfile.

The Casketfile is a text file that configures how Casket runs.

**The first line of the Casketfile is *always* the address of the site to serve.** For example:

``` casketfile
localhost:8080
```

When you save that in a file called Casketfile, Casket will automatically find it when you start it:

```sh
casket
```

If the Casketfile is in a different location or has a different name, tell Casket where it is:

``` shell
casket -conf ../path/to/Casketfile
```

The lines following a site address start with a directive. Directives are [keywords that Casket recognizes](/). For
example, [gzip](/gzip) is an HTTP directive:

``` casketfile
localhost:8080
gzip
```

Directives might have one or more arguments after them:

``` casketfile
localhost:8080
gzip
log ../access.log
```

Some directives require more configuration than can fit on one line. For those directives, you can open a *block* and
set more parameters. The open curly brace must be at the end of a line:

``` casketfile
localhost:8080
gzip
log ../access.log
markdown /blog {
  css /blog.css
  js  /scripts.js
}
```

If the directive block is left empty, you should omit the curly braces entirely.

Arguments that contain whitespace must be enclosed in quotes `"`.

The Casketfile can also have comments starting with the `#` character:

``` casketfile
# Comments can start a line
foobar # or go at the end
```

To configure multiple sites with a single Casketfile, you **must** use curly braces around each one to separate their
configurations:

``` casketfile
mysite.com {
  root /www/mysite.com
}
sub.mysite.com {
  root /www/sub.mysite.com
  gzip
  log ../access.log
}
```

As with directives, the opening curly brace must be at the end of the same line. The closing curly brace must be on its
own line. **All directives must appear inside a site's definition.**

For sites which share the same configuration, specify multiple addresses:

``` casketfile
localhost:8080, https://site.com, http://mysite.com {
  ...
}
```

Site addresses can also be defined under a specific path or have wildcards in place of individual domain labels from the
left side:

``` casketfile
example.com/static, *.example.com {
  ...
}
```

Note that using a path in your site address will route requests by longest matching prefix. If your base path is a
directory, you may wish to suffix the path with a forward slash `/`.

Use of environment variables is allowed in addresses and arguments. They must be enclosed in curly braces, and you can
use either Unix or Windows variable formats:

``` casketfile
localhost:{$PORT}
root {%SITE_ROOT%}
```

Either syntax works on any platform. A single environment variable does not expand out into multiple arguments/values.

There is **no inheritence or scripting** in the Casketfile and **you may not specify the same site address more than
once**. Yes, sometimes that means you copy+paste a few repeated lines. If you have many repeated lines, you can use the
[import](/import) directive to reduce repetition.

Alrighty, that should be more than enough to make you literate in the [Casket docs](/). Go forth and conquer!
