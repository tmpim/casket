# on

on executes a command when a specified event is triggered. This can be useful for preparing to serve a site by running a
script or starting a background process like php-fpm when the server starts, or for stopping php-fpm when the server
exits.

Each command that is executed is blocking, unless you suffix the command with a space and `&`, which will cause the
command to be run in the background. (Do not do this when the server is exiting, or the command may not finish before
its parent process exits.) The output and error of the command go to stdout and stderr, respectively. There is no stdin.

A command will only be executed once for each time it appears in the Casketfile. In other words, even if this directive
is shared by more than one host, a command will only execute once per appearance in the Casketfile.

Note that commands scheduled for the shutdown event will not execute if Casket is force-terminated, for example, by
using a "Force Quit" feature provided by your operating system. However, a typical SIGINT (Ctrl+C) will allow the
shutdown commands to execute.

## Syntax

``` casketfile
on event command
```

-   **event** is the name of the event on which to execute the command (see list below)
-   **command** is the command to execute; it may be followed by arguments

## Events

Commands can execute on the following events:

-   **startup** - The server instance is first starting up
-   **shutdown** - The server instance is shutting down (not restarting)
-   **certrenew** - A managed certificate is renewed

## Examples

Start php-fpm before the server starts listening:

``` casketfile
on startup /etc/init.d/php-fpm start
```

Stop php-fpm when the server quits:

``` casketfile
on shutdown /etc/init.d/php-fpm stop
```

On Windows, you might need to use quotes when the command path contains spaces:

``` casketfile
on startup "\"C:\Program Files\PHP\v7.0\php-cgi.exe\" -b 127.0.0.1:9123" &
```
