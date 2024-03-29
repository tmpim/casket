// Copyright 2015 Light Code Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpserver

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/caddyserver/certmagic"
	"github.com/tmpim/casket/caskettls"
)

func TestRedirPlaintextHost(t *testing.T) {
	for i, testcase := range []struct {
		Host        string // used for the site config
		Port        string
		ListenHost  string
		RequestHost string // if different from Host
	}{
		{
			Host: "foohost",
		},
		{
			Host: "foohost",
			Port: "80",
		},
		{
			Host: "foohost",
			Port: "1234",
		},
		{
			Host:       "foohost",
			ListenHost: "93.184.216.34",
		},
		{
			Host:       "foohost",
			Port:       "1234",
			ListenHost: "93.184.216.34",
		},
		{
			Host: "foohost",
			Port: strconv.Itoa(certmagic.HTTPSPort), // since this is the 'default' HTTPS port, should not be included in Location value
		},
		{
			Host:        "*.example.com",
			RequestHost: "foo.example.com",
		},
		{
			Host:        "*.example.com",
			Port:        "1234",
			RequestHost: "foo.example.com:1234",
		},
	} {
		cfg := redirPlaintextHost(&SiteConfig{
			Addr: Address{
				Host: testcase.Host,
				Port: testcase.Port,
			},
			ListenHost: testcase.ListenHost,
			TLS:        new(caskettls.Config),
		})

		// Check host and port
		if actual, expected := cfg.Addr.Host, testcase.Host; actual != expected {
			t.Errorf("Test %d: Expected redir config to have host %s but got %s", i, expected, actual)
		}
		if actual, expected := cfg.ListenHost, testcase.ListenHost; actual != expected {
			t.Errorf("Test %d: Expected redir config to have bindhost %s but got %s", i, expected, actual)
		}
		if actual, expected := cfg.Addr.Port, strconv.Itoa(certmagic.HTTPPort); actual != expected {
			t.Errorf("Test %d: Expected redir config to have port '%s' but got '%s'", i, expected, actual)
		}

		// Make sure redirect handler is set up properly
		if cfg.middleware == nil || len(cfg.middleware) != 1 {
			t.Fatalf("Test %d: Redir config middleware not set up properly; got: %#v", i, cfg.middleware)
		}

		handler := cfg.middleware[0](nil)

		// Check redirect for correctness, first by inspecting error and status code
		requestHost := testcase.Host // hostname of request might be different than in config (e.g. wildcards)
		if testcase.RequestHost != "" {
			requestHost = testcase.RequestHost
		}
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://"+requestHost+"/bar?q=1", nil)
		if err != nil {
			t.Fatalf("Test %d: %v", i, err)
		}
		status, err := handler.ServeHTTP(rec, req)
		if status != 0 {
			t.Errorf("Test %d: Expected status return to be 0, but was %d", i, status)
		}
		if err != nil {
			t.Errorf("Test %d: Expected returned error to be nil, but was %v", i, err)
		}
		if rec.Code != http.StatusMovedPermanently {
			t.Errorf("Test %d: Expected status %d but got %d", http.StatusMovedPermanently, i, rec.Code)
		}

		// Now check the Location value. It should mirror the hostname and port of the request
		// unless the port is redundant, in which case it should be dropped.
		locationHost, _, err := net.SplitHostPort(requestHost)
		if err != nil {
			locationHost = requestHost
		}
		expectedLoc := fmt.Sprintf("https://%s/bar?q=1", locationHost)
		if testcase.Port != "" && testcase.Port != DefaultHTTPSPort {
			expectedLoc = fmt.Sprintf("https://%s:%s/bar?q=1", locationHost, testcase.Port)
		}
		if got, want := rec.Header().Get("Location"), expectedLoc; got != want {
			t.Errorf("Test %d: Expected Location: '%s' but got '%s'", i, want, got)
		}
	}
}

func TestHostHasOtherPort(t *testing.T) {
	configs := []*SiteConfig{
		{Addr: Address{Host: "example.com", Port: "80"}},
		{Addr: Address{Host: "sub1.example.com", Port: "80"}},
		{Addr: Address{Host: "sub1.example.com", Port: "443"}},
	}

	if hostHasOtherPort(configs, 0, "80") {
		t.Errorf(`Expected hostHasOtherPort(configs, 0, "80") to be false, but got true`)
	}
	if hostHasOtherPort(configs, 0, "443") {
		t.Errorf(`Expected hostHasOtherPort(configs, 0, "443") to be false, but got true`)
	}
	if !hostHasOtherPort(configs, 1, "443") {
		t.Errorf(`Expected hostHasOtherPort(configs, 1, "443") to be true, but got false`)
	}
}

func TestMakePlaintextRedirects(t *testing.T) {
	configs := []*SiteConfig{
		// Happy path = standard redirect from 80 to 443
		{Addr: Address{Host: "example.com"}, TLS: &caskettls.Config{Managed: true, Enabled: true}},

		// Host on port 80 already defined; don't change it (no redirect)
		{Addr: Address{Host: "sub1.example.com", Port: "80", Scheme: "http"}, TLS: new(caskettls.Config)},
		{Addr: Address{Host: "sub1.example.com"}, TLS: &caskettls.Config{Managed: true, Enabled: true}},

		// Redirect from port 80 to port 5000 in this case
		{Addr: Address{Host: "sub2.example.com", Port: "5000"}, TLS: &caskettls.Config{Managed: true, Enabled: true}},

		// Can redirect from 80 to either 443 or 5001, but choose 443
		{Addr: Address{Host: "sub3.example.com", Port: "443"}, TLS: &caskettls.Config{Managed: true, Enabled: true}},
		{Addr: Address{Host: "sub3.example.com", Port: "5001", Scheme: "https"}, TLS: &caskettls.Config{Managed: true, Enabled: true}},
	}

	result := makePlaintextRedirects(configs)
	expectedRedirCount := 3

	if len(result) != len(configs)+expectedRedirCount {
		t.Errorf("Expected %d redirect(s) to be added, but got %d",
			expectedRedirCount, len(result)-len(configs))
	}
}

func TestEnableAutoHTTPS(t *testing.T) {
	configs := []*SiteConfig{
		{Addr: Address{Host: "example.com"}, TLS: &caskettls.Config{Managed: true, Manager: &certmagic.Config{}}},
		{}, // not managed - no changes!
	}

	if err := enableAutoHTTPS(configs, false); err != nil {
		log.Println("[ERROR]  enableAutoHTTPS failed: ", err)
	}

	if !configs[0].TLS.Enabled {
		t.Errorf("Expected config 0 to have TLS.Enabled == true, but it was false")
	}
	if configs[0].Addr.Scheme != "https" {
		t.Errorf("Expected config 0 to have Addr.Scheme == \"https\", but it was \"%s\"",
			configs[0].Addr.Scheme)
	}
	if configs[1].TLS != nil && configs[1].TLS.Enabled {
		t.Errorf("Expected config 1 to have TLS.Enabled == false, but it was true")
	}
}

func TestMarkQualifiedForAutoHTTPS(t *testing.T) {
	// TODO: caskettls.TestQualifiesForManagedTLS and this test share nearly the same config list...
	configs := []*SiteConfig{
		{Addr: Address{Host: ""}, TLS: newManagedConfig()},
		{Addr: Address{Host: "localhost"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "123.44.3.21"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "example.com"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "example.com"}, TLS: &caskettls.Config{Manual: true}},
		{Addr: Address{Host: "example.com"}, TLS: &caskettls.Config{ACMEEmail: "off"}},
		{Addr: Address{Host: "example.com"}, TLS: &caskettls.Config{ACMEEmail: "foo@bar.com", Manager: &certmagic.Config{}}},
		{Addr: Address{Host: "example.com", Scheme: "http"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "example.com", Port: "80"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "example.com", Port: "1234"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "example.com", Scheme: "https"}, TLS: newManagedConfig()},
		{Addr: Address{Host: "example.com", Port: "80", Scheme: "https"}, TLS: newManagedConfig()},
	}
	expectedManagedCount := 4

	markQualifiedForAutoHTTPS(configs)

	count := 0
	for _, cfg := range configs {
		if cfg.TLS.Managed {
			count++
		}
	}

	if count != expectedManagedCount {
		t.Errorf("Expected %d managed configs, but got %d", expectedManagedCount, count)
	}
}

func newManagedConfig() *caskettls.Config {
	return &caskettls.Config{Manager: &certmagic.Config{}}
}
