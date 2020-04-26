launchd service for macOS
=========================

This is a working sample file for a *launchd* service on Mac, which should be placed here:

```bash
/Library/LaunchDaemons/com.casketserver.web.plist
```

To create the proper directories as used in the example file:

```bash
sudo mkdir -p /etc/casket /etc/ssl/casket /var/log/casket /usr/local/bin /var/tmp /srv/www/localhost
sudo touch /etc/casket/Casketfile
sudo chown root:wheel /usr/local/bin/casket /Library/LaunchDaemons/
sudo chown _www:_www /etc/casket /etc/ssl/casket /var/log/casket
sudo chmod 0750 /etc/ssl/casket
```

Create a simple web page and Casketfile

```bash
sudo bash -c 'echo "Hello, World!" > /srv/www/localhost/index.html'
sudo bash -c 'echo "http://localhost {
    root /srv/www/localhost
}" >> /etc/casket/Casketfile'
```

Start and Stop the Casket launchd service using the following commands:

```bash
launchctl load /Library/LaunchDaemons/com.casketserver.web.plist
launchctl unload /Library/LaunchDaemons/com.casketserver.web.plist
```

To start on every boot use the `-w` flag (to write):

```bash
launchctl load -w /Library/LaunchDaemons/com.casketserver.web.plist
```

To start the service now:

```bash
launchctl start -w /Library/LaunchDaemons/com.casketserver.web.plist
```

More information can be found in this blogpost: [Running Casket as a service on macOS X server](https://denbeke.be/blog/software/running-casket-as-a-service-on-macos-os-x-server/)
