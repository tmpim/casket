# webdav

WebDAV capabilities with support for path restriction rules and users.

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/webdav/README.md)**

## Examples

### Syntax

``` casketfile
webdav [url] {
  scope       path
  modify      [true|false]
  allow       path
  allow_r     regex
  block       path
  block_r     regex
}
```

All the options are optional. + **url** is the place where you can access the WebDAV interface. Defaults to `/`. +
**scope** is an absolute or relative (to the current working directory of Casket) path that indicates the scope of the
WebDAV. Defaults to `.`. + **modify** indicates if the user has permission to edit/modify the files. Defaults to
`true`. + **allow** and **block** are used to allow or deny access to specific files or directories using their relative
path to the scope. You can use the magic word `dotfiles` to allow or deny the access to every file starting by a dot. +
**allow_r** and **block_r** and variations of the previous options but you are able to use regular expressions with
them. It is highly recommended to use this directive alongside with
[`basicauth`](/basicauth) to protect the WebDAV interface.

``` casketfile
webdav { 
  # You set the global configurations here and 
  # all the users will inherit them. 
  user1: 
  # Here you can set specific settings for the 'user1'. 
  # They will override the global ones for this specific user.
}
```

### Basic

``` casketfile
webdav
```

WebDAV on `/` for the current working directory.

### Custom Scope

``` casketfile
webdav /admin {
  scope /
}
```

WebDAV on `/admin` for the whole file system.

### Denying Rules

``` casketfile
webdav {
  scope /
  block /etc
  block /dev
}
```

WebDAV on `/` for the whole file system, without access to `/etc` and `/dev` directories.

### User Permissions

``` casketfile
basicauth / sam pass
webdav {
  scope /
    
  sam:
  block /var/www
}
```

WebDAV on `/` for the whole file system. The user `sam` can&#39;t access `/var/www` but the others can.
