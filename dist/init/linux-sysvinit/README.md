SysVinit conf for Casket
=======================

Usage
-----

* Download the appropriate Casket binary in `/usr/local/bin/casket` or use `curl https://getcasket.com | bash`.
* Save the SysVinit config file in `/etc/init.d/casket`.
* Ensure that the folder `/etc/casket` exists and that the folder `/etc/ssl/casket` is owned by `www-data`.
* Create a Casketfile in `/etc/casket/Casketfile`
* Now you can use `service casket start|stop|restart|reload|status` as `root`.

Init script manipulation
-----

The init script supports configuration via the following files:
* `/etc/default/casket` ( Debian based https://www.debian.org/doc/manuals/debian-reference/ch03.en.html#_the_default_parameter_for_each_init_script )
* `/etc/sysconfig/casket` ( CentOS based https://www.centos.org/docs/5/html/5.2/Deployment_Guide/s1-sysconfig-files.html )

The following variables can be changed:
* DAEMON: path to the casket binary file (default: `/usr/local/bin/casket`)
* DAEMONUSER: user used to run casket (default: `www-data`)
* PIDFILE: path to the pidfile (default: `/var/run/$NAME.pid`)
* LOGFILE: path to the log file for casket daemon (not for access logs) (default: `/var/log/$NAME.log`)
* CONFIGFILE: path to the casket configuration file (default: `/etc/casket/Casketfile`)
* CASKETPATH: path for SSL certificates managed by casket (default: `/etc/ssl/casket`)
* ULIMIT: open files limit (default: `8192`)
