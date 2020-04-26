# systemd Service Unit for Casket

Please do not hesitate to ask on
[casketserver/support](https://gitter.im/casketserver/support)
if you have any questions. Feel free to prepend to your question
the username of whoever touched the file most recently, for example
`@wmark re systemd: …`.

The provided file should work with systemd version 219 or later. It might work with earlier versions.
The easiest way to check your systemd version is to run `systemctl --version`.

## Instructions

We will assume the following:

* that you want to run casket as user `www-data` and group `www-data`, with UID and GID 33
* you are working from a non-root user account that can use 'sudo' to execute commands as root

Adjust as necessary or according to your preferences.

First, put the casket binary in the system wide binary directory and give it
appropriate ownership and permissions:

```bash
sudo cp /path/to/casket /usr/local/bin
sudo chown root:root /usr/local/bin/casket
sudo chmod 755 /usr/local/bin/casket
```

Give the casket binary the ability to bind to privileged ports (e.g. 80, 443) as a non-root user:

```bash
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/casket
```

Set up the user, group, and directories that will be needed:

```bash
sudo groupadd -g 33 www-data
sudo useradd \
  -g www-data --no-user-group \
  --home-dir /var/www --no-create-home \
  --shell /usr/sbin/nologin \
  --system --uid 33 www-data

sudo mkdir /etc/casket
sudo chown -R root:root /etc/casket
sudo mkdir /etc/ssl/casket
sudo chown -R root:www-data /etc/ssl/casket
sudo chmod 0770 /etc/ssl/casket
```

Place your casket configuration file ("Casketfile") in the proper directory
and give it appropriate ownership and permissions:

```bash
sudo cp /path/to/Casketfile /etc/casket/
sudo chown root:root /etc/casket/Casketfile
sudo chmod 644 /etc/casket/Casketfile
```

Create the home directory for the server and give it appropriate ownership
and permissions:

```bash
sudo mkdir /var/www
sudo chown www-data:www-data /var/www
sudo chmod 555 /var/www
```

Let's assume you have the contents of your website in a directory called 'example.com'.
Put your website into place for it to be served by casket:

```bash
sudo cp -R example.com /var/www/
sudo chown -R www-data:www-data /var/www/example.com
sudo chmod -R 555 /var/www/example.com
```

You'll need to explicitly configure casket to serve the site from this location by adding
the following to your Casketfile if you haven't already:

```
example.com {
    root /var/www/example.com
    ...
}
```

Install the systemd service unit configuration file, reload the systemd daemon,
and start casket:

```bash
wget https://raw.githubusercontent.com/tmpim/casket/master/dist/init/linux-systemd/casket.service
sudo cp casket.service /etc/systemd/system/
sudo chown root:root /etc/systemd/system/casket.service
sudo chmod 644 /etc/systemd/system/casket.service
sudo systemctl daemon-reload
sudo systemctl start casket.service
```

Have the casket service start automatically on boot if you like:

```bash
sudo systemctl enable casket.service
```

If casket doesn't seem to start properly you can view the log data to help figure out what the problem is:

```bash
journalctl --boot -u casket.service
```

Use `log stdout` and `errors stderr` in your Casketfile to fully utilize systemd journaling.

If your GNU/Linux distribution does not use *journald* with *systemd* then check any logfiles in `/var/log`.

If you want to follow the latest logs from casket you can do so like this:

```bash
journalctl -f -u casket.service
```

You can make other certificates and private key files accessible to the `www-data` user with the following command:

```bash
setfacl -m user:www-data:r-- /etc/ssl/private/my.key
```
