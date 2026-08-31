package forwardproxy

import (
	"net/http"
	"testing"
)

/*
test port blocking working
test blacklist allowed
test blacklist refused with correct status

*/

func TestWhitelistAllowing(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketTestTarget.addr, resource, casketForwardProxyWhiteListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if err = responseExpected(response, casketTestTarget.contents[resource]); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestWhitelistBlocking(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketHTTPTestTarget.addr, resource, casketForwardProxyWhiteListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("google.com:6451", resource, casketForwardProxyWhiteListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}
}

func TestLocalhostDefaultForbidden(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("localhost:6451", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("127.0.0.1:808", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("[::1]:8080", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}
}

func TestLocalNetworksDefaultForbidden(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("10.0.0.0:80", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("127.222.34.1:443", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("172.16.0.1:8080", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("192.168.192.168:888", resource, casketForwardProxyNoBlacklistOverride.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}
}

func TestBlacklistBlocking(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(blacklistedDomain, resource, casketForwardProxyBlackListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(blacklistedIPv4, resource, casketForwardProxyBlackListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}

	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy("["+blacklistedIPv6+"]:80", resource, casketForwardProxyBlackListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if response.StatusCode != http.StatusForbidden {
				t.Fatal("Expected response \"403 Forbidden\", got:", response.StatusCode)
			}
		}
	}
}

func TestBlacklistAllowing(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketTestTarget.addr, resource, casketForwardProxyBlackListing.addr, httpProxyVer,
				"", useTls)
			if err != nil {
				t.Fatal(err)
			} else if err = responseExpected(response, casketTestTarget.contents[resource]); err != nil {
				t.Fatal(err)
			}
		}
	}
}
