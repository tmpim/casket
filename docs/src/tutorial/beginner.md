# Beginner Tutorial

This tutorial will help you install, run, and configure Casket for your first time. It assumes you have never used a web
server before! (If you have, do the [quick start](/tutorial/).) Although Casket is very easy to use, it is still
expected that you are already familiar with using your machine:

-   Extracting, moving, and renaming files
-   Managing users and file permissions
-   Using the terminal or command line
-   Configuring your firewall

With these prerequisites, you're ready to go.

<br />

#### Topics

[[toc]]

## Download

Download Casket from [the download page](https://github.com/tmpim/casket/releases). Builds are available for most common
architectures. You can also [build it from source](https://github.com/tmpim/casket?tab=readme-ov-file#building) if you 
prefer.

## Install

The file you downloaded is a compressed archive. You'll want to extract the Casket binary (executable file).

#### Windows

1. Right-click the .zip file and choose "Extract All". Choose any folder to extract into, just don't lose track of it.
   You can delete it when we're done.
2. Move the executable to any folder that's easy to get at. For example, `C:\Casket`.

#### macOS

1. Double-click the .zip file to extract it, or run this command:
    ``` shell
    unzip casket*.zip casket
    ```
2. Next, we will move the Casket binary into a folder where we can execute it easily. Any \$PATH location will do:
    ``` shell
    mv ./casket /usr/local/bin
    ```
    If you get permission denied errors, you'll need to run with sudo.

#### Linux

1. Run this command to extract the archive:
    ``` shell
    tar -xzf casket*.tar.gz casket
    ```
2. Next, we will move the Casket binary into a folder where we can execute it easily. Any \$PATH location will do:
    ``` shell
    mv ./casket /usr/local/bin
    ```
    If you get permission denied errors, you'll need to run with sudo.

## Run

By default, Casket will use the current directory (the directory it is being executed from, *not* the folder where the
binary lives) as the root of the site. This makes it easy to work on sites locally!

Using the terminal or command line, change into the folder where your site is:

``` shell
cd path/to/my/site
```

And run Casket:

``` shell
casket
```

Load <http://localhost:2015> in your browser. If you see a 404 error, then Casket is working but your site is missing an
index file.

You can quit Casket by pressing Ctrl+C. It will terminate as gracefully as possible.

## Configure

Your site is already fit for production! But it's not ideal, because we happen to running on `localhost` (your home
computer):

1.  The site is being served on port 2015, not 80 (the standard HTTP port).
2.  The site isn't protected with HTTPS.

It's easy to fix both of these simply by telling Casket the name of the site to serve. By "name", we mean a domain name.
We'll use `example.com` here, but you use your real domain. This next part is only going to work if your computer is
accessible from the wider Internet on ports 80 and 443, and your domain name points to the computer you're on. If not,
or you don't have a real domain name, then use `localhost` as your domain name.

The name of the site is also called the **host** or **hostname**. One way to specify the host is with a command line
option:

``` shell
casket -host example.com
```

The first time you run Casket with a real hostname (not `localhost`), you'll be asked to enter your email address. This
is because Casket needs to verify you own the domain and to store the certificate safely on disk for you.

After submitting your email address, do you see an error like `permission denied`? That's because Casket is trying to
bind to ports 80 and 443 for a *real* site, but doing this requires root or Administrator privileges:

#### Windows

Right-click cmd.exe and click "Run as Administrator". Then run Casket again:

``` shell
C:\Casket\casket.exe -host example.com
```

#### macOS

Use sudo to run Casket as root:

``` shell
sudo casket -host example.com
```

#### Linux

On a real server, you would consider using [setcap](https://linux.die.net/man/8/setcap) like this:

``` shell
sudo setcap cap_net_bind_service=+ep $(which casket)
casket -host example.com
```

But for your own computer, running Casket with sudo probably fine:

``` shell
sudo casket -host example.com
```

If you got your permissions right and run Casket again, you'll see:

``` casketfile
Activating privacy features... done.
https://example.com
http://example.com
```

Using a real domain name triggers Casket's privacy features, which operate on ports 80 and 443. If you're just using
`localhost` as your hostname, Casket continues to serve on port 2015 unless you change it with the `-port` option.

The [command line interface](/cli) is great for quick Casket configurations. But what if you want to reuse the same
config every time? It's easy with the Casketfile.

The **Casketfile** is a text file that tells Casket how to serve. It usually lives next to your site. Let's make one.
Create a file called `Casketfile` (no extension), and put one line in it (use your actual domain name, or localhost):

``` casketfile
example.com
```

Casket will automatically find that file when you start it:

``` shell
casket
```

This works because the first line of a Casketfile is always the address (or name) of the site to serve.

If the Casketfile is ever in another directory than the current one, you can tell Casket where to get the Casketfile:

``` shell
casket -conf ../path/to/Casketfile
```

You almost know enough to be dangerous. Next, learn [how to wield the Casketfile](/tutorial/casketfile). You'll love how
easy it is to write.

Or [go to docs index](/).
