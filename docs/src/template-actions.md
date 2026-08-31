# Template Actions

Casket's template features make it possible to add some dynamic content to your static website and include pages to help
reduce duplication. Templates are supported by a few different directives such as [templates](/templates),
[browse](/browse), and [markdown](/markdown).

::: info
Template actions only work on directives that support them. Check the documentation for your directive to see if and how
they can be used.
:::

Casket templates use the syntax defined in Go's [text/template](http://golang.org/pkg/text/template/) package.
Understanding text/template will help you take full advantage of templates, but for those of us who aren't programmers,
here are some simplified instructions. Casket also adds some extra helpful functions for templating web pages
specifically, which we've documented here.

## Basic Syntax

::: v-pre
Template actions are enclosed between `{{` and `}}` markers. Template words are case-sensitive.
:::

## Common Functions

#### Include another file:

```
{{.Include "path/to/file.html"}}  // no arguments
{{.Include "path/to/file.html" "arg1" 2 "value 3"}}  // with arguments
```

#### Get argument within included file:

```
{{index .Args 0}} // 0 is first argument, and so on
```

#### Include and render a Markdown file: (not needed in [markdown](/markdown) middleware)

```
{{.Markdown "path/to/file.md"}}
```

#### Show current timestamp: ([format](https://golang.org/pkg/time/#pkg-constants))

```
{{.Now "Monday, 2 Jan 2006"}}
```

#### Cookie value:

```
{{.Cookie "name"}}
```

#### Header field value:

```
{{.Header "name"}}
```

#### Visitor's IP:

```
{{.IP}}
```

#### Server's IP:

```
{{.ServerIP}}
```

#### Visitor's hostname (requires DNS lookup, probably slow):

```
{{.Hostname}}
```

#### Request URI:

```
{{.URI}}
```

#### Host portion of request:

```
{{.Host}}
```

#### Port portion of request:

```
{{.Port}}
```

#### Request method:

```
{{.Method}}
```

#### Whether request path matches another path:

```
{{.PathMatches "/v1/some/path"}}
```

#### A part of the URL:

```
{{.URL.RawQuery}}
```

#### TLS version:

```
{{.TLSVersion}}
```

#### Environment variables:

```
{{.Env.ENV_VAR_NAME}}
```

#### Truncate value to certain length: (from beginning, or from end)

```
{{.Truncate "value" 3}}  // "val"
```

```
{{.Truncate "value" -3}} // "lue"
```

#### String replacement:

```
{{.Replace "haystack" "needle" "replacement"}}
```

#### Current date/time object: (useful in date-related functions)

```
{{.NowDate}}
```

#### Get extension from file path:

```
{{.Ext "path/filename.ext"}}
```

#### Strip extension from file path:

```
{{.StripExt "filename.ext"}}
```

#### Strip HTML, leaving only the plain text:

```
{{.StripHTML "Shows <b>only</b> text content"}}
```

#### Lower-case a string:

```
{{.ToLower "Makes Me ONLY lowercase"}}
```

#### Upper-case a string:

```
{{.ToUpper "Makes me only UPPERCASE"}}
```

#### Split a string by separator:

```
{{.Split "123-456-7890" "-"}}
```

#### Convert list of values to a slice (array):

```
{{.Slice "a" "b" "c"}}
```

#### Map keys to values: (useful in advanced cases, with subtemplates, etc.)

```
{{.Map "key1" "value1" "key2" "value2"}}
```

#### List files in a directory:

```
{{.Files "sub/directory"}}
{{.Context.Files "sub/directory"}} // prefix with .Context in markdown templates
```

#### Whether [HTTPS interception](/mitm-detection) is likely:

```
{{.IsMITM}}
```

#### Random string of random length between min and max length:

```
{{.RandomString 100 10000}}
```

## Built-in Sanitization Functions {#sanitization}

These functions are built into text/template but you may find them helpful.

#### Make HTML-safe (escape special characters):

```
{{html "Makes it <i>safe</i> to render as HTML"}}
```

#### Make JavaScript-safe:

```
{{js "Makes content safe for use in JS"}}
```

#### URL-escape (query-encode):

```
{{urlquery "Makes safe for URL query strings"}}
```

## Control Statements

#### If:

```
{{if .PathMatches "/v1/secret/sauce"}}
    Only for secret sauce pages
{{end}}
```

#### If-else:

```
{{if .PathMatches "/v1/secret/sauce"}}
    Only for secret sauce pages
{{else}}
    No secret sauce for you
{{end}}
```

#### If-elseif-else:

```
{{if .PathMatches "/v1/secret/sauce"}}
    Only for secret sauce pages
{{else if eq .URL "/v1/banana.html"}}
    You're on the banana page
{{else}}
    No bananas or secret sauce
{{end}}
```

#### Range: (iterate data; this example dumps request headers)

```
{{range $field, $val := .Req.Header}}
    {{$field}}: {{$val}}
{{end}}
```

#### Server-side comments:

```
{{/* This isn't sent to the client */}}
```

## Comparison Functions

Useful in "if" statements, you can use comparison functions:

-   `eq` Equal
-   `ne` Not equal
-   `lt` Less than
-   `le` Less than or equal
-   `gt` Greater than
-   `ge` Greater than or equal

Or these logic functions:

-   `not` Reverses the if condition
-   `or` Returns first non-empty argument or the last argument
-   `and` Returns first empty argument or the last argument

## Further Reading

These are just a few examples of what you can do. If you need even more template power, read the description of the
[text/template](http://golang.org/pkg/text/template/) package which goes into much more detail.
