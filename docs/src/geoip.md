# geoip

The geoip directive adds the ability to determine a client's geolocation based on their IP address using a
[MaxMind database](https://www.maxmind.com/en/geoip2-services-and-databases).

## Syntax

``` casketfile
geoip <database_path>
```

* **database_path**: The path to a MaxMind database file (e.g., `GeoLite2-City.mmdb`).

## Placeholders

The following placeholders are available:

```
  geoip_country_code - Country ISO code, example CY for Cyprus
  geoip_country_geoname_id - GeoNameID of the city, example 146669
  geoip_latitude - Latitude, example 34.684100
  geoip_longitude - Longitude, example 33.037900
  geoip_time_zone - Time zone, example Asia/Nicosia
  geoip_country_eu - Return 'true' if country in Europen Union
  geoip_country_name - Full country name
  geoip_city_name - City name
  geoip_city_geoname_id - GeoNameID of the city, example 146384
  geoip_geohash - Geohash of latitude and longitude
```

## Missing geolocation data

If there is no geolocation data for an IP address most of the placeholders listed above will be empty. The exceptions 
are `geoip_country_code`, `geoip_country_name`, and `geoip_city_name`. If the request originated over the system 
loopback interface (e.g., 127.0.0.1) those vars will be set to `**`, `Loopback`, and `Loopback` respectively. For any 
other address, including private addresses such as 192.168.0.1, the values will be `!!`, `No Country`, and `No City`
respectively.

## Auto-reloading and auto-updating

The geoip directive will automatically reload `.mmdb` files if they are modified or overwritten. Files are reloaded with
a 500ms debounce, and multiple `geoip` directives with the same path will share the reference to the file in-memory.

With this feature, you can use any tooling to automatically update your MaxMind database, such as the 
[maxmind/geoipupdate](https://github.com/maxmind/geoipupdate/) tool. For example, with Docker:

``` yml
services:
  geoipupdate:
    image: ghcr.io/maxmind/geoipupdate:latest
    container_name: geoip-updater
    restart: unless-stopped
    environment:
      GEOIPUPDATE_ACCOUNT_ID: ${GEOIPUPDATE_ACCOUNT_ID}
      GEOIPUPDATE_LICENSE_KEY: ${GEOIPUPDATE_LICENSE_KEY}
      GEOIPUPDATE_EDITION_IDS: GeoLite2-City
      GEOIPUPDATE_FREQUENCY: 168
      GEOIPUPDATE_DB_DIR: /usr/share/GeoIP
    volumes:
      - geoip-data:/usr/share/GeoIP

volumes:
  geoip-data:
```

``` casketfile
geoip /path/to/docker/volume/GeoLite2-City.mmdb
```

## Examples

### Set database path and return country code header

``` casketfile
geoip /path/to/db/GeoLite2-City.mmdb
header Country-Code {geoip_country_code}
```

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

### Include geolocation info in the access log

``` casketfile
log / {$HOME}/log/access.log "{when_iso} {status} {method} {latency_ms} ms {size} bytes {geoip_country_code} {remote} {host} {proto} \"{uri}\" \"{>User-Agent}\""
```
