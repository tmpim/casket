package geoip

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tmpim/casket/caskethttp/httpserver"
)

type testResponseRecorder struct {
	*httpserver.ResponseWriterWrapper
}

func (testResponseRecorder) CloseNotify() <-chan bool { return nil }

func TestReplacers(t *testing.T) {
	reloader, err := newDBReloader("./test-data/GeoLite2-City.mmdb", defaultDebounce, nil)
	if err != nil {
		t.Fatalf("geoip: Can't open database: %v", err)
	}
	defer reloader.Close()

	config := Config{}

	l := GeoIP{
		Next: httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		}),
		Reloader: reloader,
		Config:   config,
	}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.RemoteAddr = "212.50.99.193"
	rr := httpserver.NewResponseRecorder(testResponseRecorder{
		ResponseWriterWrapper: &httpserver.ResponseWriterWrapper{ResponseWriter: httptest.NewRecorder()},
	})

	rr.Replacer = httpserver.NewReplacer(r, rr, "-")

	l.ServeHTTP(rr, r)

	if got, want := rr.Replacer.Replace("{geoip_country_code}"), "CY"; got != want {
		t.Errorf("Expected custom placeholder {geoip_country_code} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_country_name}"), "Cyprus"; got != want {
		t.Errorf("Expected custom placeholder {geoip_country_name} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_country_eu}"), "false"; got != want {
		t.Errorf("Expected custom placeholder {geoip_country_eu} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_city_name}"), "Limassol"; got != want {
		t.Errorf("Expected custom placeholder {geoip_city_name} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_latitude}"), "34.684100"; got != want {
		t.Errorf("Expected custom placeholder {geoip_latitude} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_longitude}"), "33.037900"; got != want {
		t.Errorf("Expected custom placeholder {geoip_longitude} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_time_zone}"), "Asia/Nicosia"; got != want {
		t.Errorf("Expected custom placeholder {geoip_time_zone} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_geohash}"), "swpmrf13wbgg"; got != want {
		t.Errorf("Expected custom placeholder {geoip_geohash} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_city_geoname_id}"), "146384"; got != want {
		t.Errorf("Expected custom placeholder {geoip_city_geoname_id} to be set (%s), but it wasn't; got: %s", want, got)
	}

	if got, want := rr.Replacer.Replace("{geoip_country_geoname_id}"), "146669"; got != want {
		t.Errorf("Expected custom placeholder {geoip_country_geoname_id} to be set (%s), but it wasn't; got: %s", want, got)
	}

	//
	// Verify that a request via the loopback interface address results in
	// the expected placeholder values.
	//
	var loopback_placeholders = [][2]string{
		{"{geoip_country_code}", "**"},
		{"{geoip_country_name}", "Loopback"},
		{"{geoip_city_name}", "Loopback"},
		{"{geoip_country_geoname_id}", "0"},
		{"{geoip_city_geoname_id}", "0"},
		{"{geoip_latitude}", "0.000000"},
		{"{geoip_longitude}", "0.000000"},
		{"{geoip_geohash}", "s00000000000"},
		{"{geoip_time_zone}", ""},
	}

	r = httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.RemoteAddr = "127.0.0.1"
	rr = httpserver.NewResponseRecorder(testResponseRecorder{
		ResponseWriterWrapper: &httpserver.ResponseWriterWrapper{ResponseWriter: httptest.NewRecorder()},
	})

	rr.Replacer = httpserver.NewReplacer(r, rr, "-")

	l.ServeHTTP(rr, r)

	for _, v := range loopback_placeholders {
		if got, want := rr.Replacer.Replace(v[0]), v[1]; got != want {
			t.Errorf("Expected custom placeholder %s to be set (%s), but it wasn't; got: %s", v[0], want, got)
		}
	}

	//
	// Verify that a request via a private address results in the expected
	// placeholder values. Note that the MaxMind DB doesn't include
	// location data for private addresses.
	//
	var private_addr_placeholders = [][2]string{
		{"{geoip_country_code}", "!!"},
		{"{geoip_country_name}", "No Country"},
		{"{geoip_city_name}", "No City"},
		{"{geoip_country_geoname_id}", "0"},
		{"{geoip_city_geoname_id}", "0"},
		{"{geoip_latitude}", "0.000000"},
		{"{geoip_longitude}", "0.000000"},
		{"{geoip_geohash}", "s00000000000"},
		{"{geoip_time_zone}", ""},
	}

	r = httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.RemoteAddr = "192.168.0.1"
	rr = httpserver.NewResponseRecorder(testResponseRecorder{
		ResponseWriterWrapper: &httpserver.ResponseWriterWrapper{ResponseWriter: httptest.NewRecorder()},
	})

	rr.Replacer = httpserver.NewReplacer(r, rr, "-")

	l.ServeHTTP(rr, r)

	if got, want := rr.Replacer.Replace("{geoip_country_code}"), "!!"; got != want {
		t.Errorf("Expected custom placeholder {geoip_country_code} to be set (%s), but it wasn't; got: %s", want, got)
	}

	for _, v := range private_addr_placeholders {
		if got, want := rr.Replacer.Replace(v[0]), v[1]; got != want {
			t.Errorf("Expected custom placeholder %s to be set (%s), but it wasn't; got: %s", v[0], want, got)
		}
	}
}
