# geoip

geoip is a Casket plugin that allow to determine user Geolocation by IP address using MaxMind database.

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/geoip/README.md)**

## Examples

### Proxy pass headers to backend

``` casketfile
localhost
geoip /path/to/db/GeoLite2-City.mmdb
proxy / localhost:3000 {
  header_upstream Country-Name {geoip_country_name}
  header_upstream Country-Code {geoip_country_code}
  header_upstream Country-Eu {geoip_country_eu}
  header_upstream City-Name {geoip_city_name}
  header_upstream Latitude {geoip_latitude}
  header_upstream Longitude {geoip_longitude}
  header_upstream Time-Zone {geoip_time_zone}
}
```

This will pass Country-Name/Country-Code.. etc. headers to you backend and you must get it like
r.Header.Get(&#34;Country-Name&#34;) or other way related to you backend language/framework
