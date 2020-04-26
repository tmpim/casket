Upstart conf for Casket
=====================

Usage
-----

Usage in this blogpost: [Running Casket Server as a service with Upstart](https://denbeke.be/blog/servers/running-casket-server-as-a-service/).
Short recap:

* Download Casket in `/usr/local/bin/casket` and execute `sudo setcap cap_net_bind_service=+ep /usr/local/bin/casket`.
* Save the appropriate upstart config file in `/etc/init/casket.conf`.
* Ensure that the folder `/etc/casket` exists and that the subfolder .casket is owned by `www-data`.
* Create a Casketfile in `/etc/casket/Casketfile`.
* Now you can use `sudo service casket start|stop|restart`.
