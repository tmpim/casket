# cors

Supports Cross Origin Resource Sharing headers

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/cors/README.md)**

## Examples

### Simple usage

``` casketfile
cors
```

Allows all origins access to all resources

### Only allow certain origin domains

``` casketfile
cors / http://mytrusteddomain.tld http://myotherdomain.com
```

Only allow cross-origin requests from a few specific domains

### Full config

``` casketfile
cors / {
  origin            http://allowedsite.com
  origin            http://anothersite.org https://anothersite.org
  methods           POST,PUT
  allow_credentials false
  max_age           3600
  allowed_headers   X-Custom-Header,X-Foobar
  exposed_headers   X-Something-Special,SomethingElse
}
```

Shows examples of all available options
