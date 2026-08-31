# templates

templates allows you to add template actions to your pages. Templates can be a convenient way to render the current
timestamp, request URL, visitor's IP address, and more. You also get basic control statements like `if` and `range`. See
[Template Actions](/template-actions) for instructions on how to use templates.

Some common uses of templates are to include the content of other files, show the current date or time, and hide or show
certain parts of a page depending on the request path, cookies, or headers.

Note that [custom error pages](/errors) do not get executed as templates, even if this directive is enabled. Error pages
are served by a separate middleware.

Templates can come from static files or be loaded by other middleware. For example, you can [proxy](/proxy) to a backend
that outputs a template which Casket then executes with this directive.

## Syntax

``` casketfile
templates [path [extensions...]]
```

-   **path** is the path to match before templates will be invoked
-   **extensions...** is a list of space-separated file extensions that will have templates

To specify certain extensions, a path must also be provided. The default path is / and the default extensions that will
be executed as templates are .html, .htm, .tpl, .tmpl, and .txt.

For more options, open a block:

``` casketfile
templates {
  path    basepath
  ext     extensions...
  between open_delim close_delim
}
```

::: v-pre
-   **path** is the base path to match for templates to be invoked.
-   **ext** is a list of space-separated file extensions that are executed as templates.
-   **between** specifies the open and close delimiter for templates. Default is `{{` and `}}`.
:::

## Template Format

See [Template Actions](/template-actions).

## Example Template File

``` html
<!DOCTYPE html>
<html>
    <head>
        <title>Example Templated File</title>
    </head>
    <body>
        {{.Include "/includes/header.html"}}
        <p>
            Welcome {{.IP}}! You're visiting {{.URI}}.
        </p>
        {{.Include "/includes/footer.html"}}
    </body>
</html>
```

## Examples

Enable templates for all .html, .htm, .tpl, .tmpl, and .txt files:

``` casketfile
templates
```

Templates for the same file extensions but only under /portfolio:

``` casketfile
templates /portfolio
```

Enable templates only on .html and .txt files in /portfolio:

``` casketfile
templates /portfolio .html .txt
```
