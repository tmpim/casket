// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package forwardproxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tmpim/casket/caskethttp/forwardproxy/httpclient"
	_ "github.com/tmpim/casket/caskethttp/header"
	_ "github.com/tmpim/casket/caskethttp/httpserver"
	_ "github.com/tmpim/casket/caskethttp/redirect"
	_ "github.com/tmpim/casket/caskethttp/root"
	"golang.org/x/net/http2"
)

func dial(proxyAddr, httpProxyVer string, useTls bool) (net.Conn, error) {
	if useTls {
		return tls.Dial("tcp", proxyAddr, &tls.Config{InsecureSkipVerify: true,
			NextProtos: []string{httpVersionToAlpn[httpProxyVer]}})
	} else {
		return net.Dial("tcp", proxyAddr)
	}
}

func getViaProxy(targetHost, resource, proxyAddr, httpProxyVer, proxyCredentials string, useTls bool) (*http.Response, error) {
	proxyConn, err := dial(proxyAddr, httpProxyVer, useTls)
	if err != nil {
		return nil, err
	}
	return getResourceViaProxyConn(proxyConn, targetHost, resource, httpProxyVer, proxyCredentials)
}

// if connect is not successful - that response is returned, otherwise the requested resource
func connectAndGetViaProxy(targetHost, resource, proxyAddr, httpTargetVer, proxyCredentials, httpProxyVer string, useTls bool) (*http.Response, error) {
	proxyConn, err := dial(proxyAddr, httpProxyVer, useTls)
	if err != nil {
		return nil, err
	}

	req := http.Request{Header: make(http.Header)}
	if len(proxyCredentials) > 0 {
		req.Header.Set("Proxy-Authorization", proxyCredentials)
	}
	req.Host = targetHost
	req.URL, err = url.Parse("https://" + req.Host)
	if err != nil {
		return nil, err
	}
	req.RequestURI = req.Host
	req.Method = "CONNECT"
	req.Proto = httpProxyVer

	var resp *http.Response
	switch httpProxyVer {
	case "HTTP/2.0":
		req.ProtoMajor = 2
		req.ProtoMinor = 0
		pr, pw := io.Pipe()
		req.Body = ioutil.NopCloser(pr)
		t := http2.Transport{}
		clientConn, err := t.NewClientConn(proxyConn)
		if err != nil {
			return nil, err
		}
		resp, err = clientConn.RoundTrip(&req)
		if err != nil {
			return resp, err
		}
		proxyConn = httpclient.NewHttp2Conn(proxyConn, pw, resp.Body)
	case "HTTP/1.1":
		req.ProtoMajor = 1
		req.ProtoMinor = 1
		req.Write(proxyConn)
		resp, err = http.ReadResponse(bufio.NewReader(proxyConn), &req)
		if err != nil {
			return resp, err
		}
	default:
		panic("proxy ver: " + httpProxyVer)
	}

	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, err
	}

	return getResourceViaProxyConn(proxyConn, targetHost, resource, httpTargetVer, proxyCredentials)
}

func getResourceViaProxyConn(proxyConn net.Conn, targetHost, resource, httpTargetVer, proxyCredentials string) (*http.Response, error) {
	var err error

	req := http.Request{Header: make(http.Header)}
	if len(proxyCredentials) > 0 {
		req.Header.Set("Proxy-Authorization", proxyCredentials)
	}
	req.Host = targetHost
	req.URL, err = url.Parse("http://" + req.Host + resource)
	if err != nil {
		return nil, err
	}
	req.RequestURI = req.Host + resource
	req.Method = "GET"
	req.Proto = httpTargetVer

	switch httpTargetVer {
	case "HTTP/2.0":
		req.ProtoMajor = 2
		req.ProtoMinor = 0
		t := http2.Transport{AllowHTTP: true}
		clientConn, err := t.NewClientConn(proxyConn)
		if err != nil {
			return nil, err
		}
		return clientConn.RoundTrip(&req)
	case "HTTP/1.1":
		req.ProtoMajor = 1
		req.ProtoMinor = 1
		t := http.Transport{Dial: func(network, addr string) (net.Conn, error) {
			return proxyConn, nil
		}}
		return t.RoundTrip(&req)
	default:
		panic("proxy ver: " + httpTargetVer)
	}
}

// If response is expected: returns nil.
func responseExpected(res *http.Response, expectedResponse []byte) error {
	responseLen := len(expectedResponse) + 2 // 2 extra bytes is enough to detected that expectedResponse is longer
	response := make([]byte, responseLen)
	var nTotal int
	for {
		n, err := res.Body.Read(response[nTotal:])
		nTotal += n
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if nTotal == responseLen {
			return fmt.Errorf("nTotal == responseLen, but haven't seen io.EOF. Expected response: %s\nGot: %s\n",
				expectedResponse, response)
		}
	}
	response = response[:nTotal]
	if len(expectedResponse) != len(response) {
		return fmt.Errorf("Expected length: %d. Got thus far: %d. Expected response: %s\nGot: %s\n",
			len(expectedResponse), len(response), expectedResponse, response)
	}
	for i := range response {
		if response[i] != expectedResponse[i] {
			return fmt.Errorf("Response mismatch at character #%d. Expected response: %s\nGot: %s\n",
				i, expectedResponse, response)
		}
	}
	return nil
}

func TestPassthrough(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 2 * time.Second}
	resp, err := client.Get("https://" + casketForwardProxy.addr)
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketForwardProxy.contents[""]); err != nil {
		t.Fatal(err)
	}

	resp, err = client.Get("https://" + casketForwardProxy.addr + "/pic.png")
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketForwardProxy.contents["/pic.png"]); err != nil {
		t.Fatal(err)
	}

	resp, err = client.Get("https://" + casketForwardProxy.addr + "/idontexist")
	if err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected: 404 StatusNotFound, got %d. Response: %#v\n", resp.StatusCode, resp)
	}
}

func TestGETNoAuth(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketHTTPTestTarget.addr, resource, casketForwardProxy.addr, httpProxyVer, credentialsEmpty, useTls)
			if err != nil {
				t.Fatal(err)
			} else if err = responseExpected(response, casketHTTPTestTarget.contents[resource]); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGETAuthCorrect(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketHTTPTestTarget.addr, resource, casketForwardProxyAuth.addr, httpProxyVer, credentialsCorrect, useTls)
			if err != nil {
				t.Fatal(err)
			} else if err = responseExpected(response, casketHTTPTestTarget.contents[resource]); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGETAuthWrong(t *testing.T) {
	useTls := true
	for _, wrongCreds := range credentialsWrong {
		for _, httpProxyVer := range testHttpProxyVersions {
			for _, resource := range testResources {
				response, err := getViaProxy(casketHTTPTestTarget.addr, resource, casketForwardProxyAuth.addr, httpProxyVer, wrongCreds, useTls)
				if err != nil {
					t.Fatal(err)
				}
				if response.StatusCode != http.StatusProxyAuthRequired {
					t.Fatalf("Expected response: 407 StatusProxyAuthRequired, Got: %d %s\n",
						response.StatusCode, response.Status)
				}
			}
		}
	}
}

func TestProxySelfGet(t *testing.T) {
	useTls := true
	// GETNoAuth to self
	for _, httpTargetVer := range testHttpTargetVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketForwardProxy.addr, resource, casketForwardProxy.addr, httpTargetVer, credentialsEmpty, useTls)
			if err != nil {
				t.Fatal(err)
			} else if err = responseExpected(response, casketForwardProxy.contents[resource]); err != nil {
				t.Fatal(err)
			}
		}
	}

	// GETAuthCorrect to self
	for _, httpTargetVer := range testHttpTargetVersions {
		for _, resource := range testResources {
			response, err := getViaProxy(casketForwardProxyAuth.addr, resource, casketForwardProxyAuth.addr, httpTargetVer, credentialsCorrect, useTls)
			if err != nil {
				t.Fatal(err)
			} else if err = responseExpected(response, casketForwardProxyAuth.contents[resource]); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// TODO: self TestProxySelfConnect.
// It requires tls-in-tls, which tests are not currently set up for.
// Low priority since this is a functionality issue, not security, and it would be easily caught in the wild.

func TestConnectNoAuth(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, httpTargetVer := range testHttpTargetVersions {
			for _, resource := range testResources {
				response, err := connectAndGetViaProxy(casketTestTarget.addr, resource, casketForwardProxy.addr, httpTargetVer, credentialsEmpty, httpProxyVer, useTls)
				if err != nil {
					t.Fatal(err)
				} else if err = responseExpected(response, casketTestTarget.contents[resource]); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestConnectAuthCorrect(t *testing.T) {
	useTls := true
	for _, httpProxyVer := range testHttpProxyVersions {
		for _, httpTargetVer := range testHttpTargetVersions {
			for _, resource := range testResources {
				response, err := connectAndGetViaProxy(casketTestTarget.addr, resource, casketForwardProxyAuth.addr, httpTargetVer, credentialsCorrect, httpProxyVer, useTls)
				if err != nil {
					t.Fatal(httpProxyVer, httpTargetVer, err)
				} else if err = responseExpected(response, casketTestTarget.contents[resource]); err != nil {
					t.Fatal(httpProxyVer, httpTargetVer, err)
				}
			}
		}
	}
}

func TestConnectAuthWrong(t *testing.T) {
	useTls := true
	for _, wrongCreds := range credentialsWrong {
		for _, httpProxyVer := range testHttpProxyVersions {
			for _, httpTargetVer := range testHttpTargetVersions {
				for _, resource := range testResources {
					response, err := connectAndGetViaProxy(casketTestTarget.addr, resource, casketForwardProxyAuth.addr, httpTargetVer, wrongCreds, httpProxyVer, useTls)
					if err != nil {
						t.Fatal(err)
					}
					if response.StatusCode != http.StatusProxyAuthRequired {
						t.Fatalf("Expected response: 407 StatusProxyAuthRequired, Got: %d %s\n",
							response.StatusCode, response.Status)
					}
				}
			}
		}
	}
}

func TestPAC(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 2 * time.Second}
	resp, err := client.Get("https://" + casketForwardProxy.addr + "/proxy.pac")
	if err != nil {
		t.Fatal(err)
	}
	splitAddr := strings.Split(casketForwardProxy.addr, ":")
	if err = responseExpected(resp, fmt.Appendf(nil, pacFile, splitAddr[0], splitAddr[1])); err != nil {
		t.Fatal(err)
	}

	resp, err = client.Get("https://" + casketForwardProxyProbeResist.addr + "/superhiddenfile.pac")
	if err != nil {
		t.Fatal(err)
	}
	splitAddr = strings.Split(casketForwardProxyProbeResist.addr, ":")
	if err = responseExpected(resp, fmt.Appendf(nil, pacFile, splitAddr[0], splitAddr[1])); err != nil {
		t.Fatal(err)
	}
}

func TestCONNECTViaUpstream(t *testing.T) {
	useTls := true
	for range make([]byte, 5) { // do several times to test http2 connection reuse
		for _, httpProxyVer := range testHttpProxyVersions {
			for _, httpTargetVer := range testHttpTargetVersions {
				for _, resource := range testResources {
					response, err := connectAndGetViaProxy(casketTestTarget.addr, resource, casketAuthedUpstreamEnter.addr,
						httpTargetVer, credentialsUpstreamCorrect, httpProxyVer, useTls)
					if err != nil {
						t.Fatal(err)
					} else if err = responseExpected(response, casketTestTarget.contents[resource]); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}

func TestGETViaUpstream(t *testing.T) {
	useTls := true
	for range make([]byte, 5) { // do several times to test http2 connection reuse
		for _, httpProxyVer := range testHttpProxyVersions {
			for _, resource := range testResources {
				response, err := getViaProxy(casketHTTPTestTarget.addr, resource, casketAuthedUpstreamEnter.addr, httpProxyVer,
					credentialsUpstreamCorrect, useTls)
				if err != nil {
					t.Fatal(err)
				} else if err = responseExpected(response, casketHTTPTestTarget.contents[resource]); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestUpstreamPassthrough(t *testing.T) {
	// Usptreaming proxy still hosts things as expected
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 2 * time.Second}
	resp, err := client.Get("https://" + casketAuthedUpstreamEnter.addr)
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketAuthedUpstreamEnter.contents[""]); err != nil {
		t.Fatal(err)
	}

	resp, err = client.Get("https://" + casketAuthedUpstreamEnter.addr + "/pic.png")
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketAuthedUpstreamEnter.contents["/pic.png"]); err != nil {
		t.Fatal(err)
	}

	resp, err = client.Get("https://" + casketAuthedUpstreamEnter.addr + "/idontexist")
	if err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected: 404 StatusNotFound, got %d. Response: %#v\n", resp.StatusCode, resp)
	}
}
