package forwardproxy

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tmpim/casket"
)

var credentialsEmpty = ""
var credentialsCorrectPlain = "test:pass"
var credentialsCorrect = "Basic dGVzdDpwYXNz"                                 // test:pass
var credentialsUpstreamCorrect = "basic dXBzdHJlYW10ZXN0OnVwc3RyZWFtcGFzcw==" // upstreamtest:upstreampass
var credentialsWrong = []string{
	"",
	"\"\"",
	"Basic dzp3",
	"Basic \"\"",
	"Foo bar",
	"Tssssssss",
	"Basic dpz3 asp",
}

/*
Test naming: Test{httpVer}Proxy{Method}{Auth}{Credentials}{httpVer}
GET/CONNECT -- get gets, connect connects and gets
Auth/NoAuth
Empty/Correct/Wrong -- tries different credentials
*/
var testResources = []string{"", "/pic.png"}
var testHttpProxyVersions = []string{"HTTP/2.0", "HTTP/1.1"}
var testHttpTargetVersions = []string{"HTTP/1.1"}
var httpVersionToAlpn = map[string]string{
	"HTTP/1.1": "http/1.1",
	"HTTP/2.0": "h2",
}

var blacklistedDomain = "google-public-dns-a.google.com" // supposed to ever resolve to one of 2 IP addresses below
var blacklistedIPv4 = "8.8.8.8"
var blacklistedIPv6 = "2001:4860:4860::8888"

type casketTestServer struct {
	*casket.Instance
	addr string // could be http or https

	HTTPRedirectPort string // used in probe-resist tests to simulate default Casket's http->https redirect
	root             string // expected to have index.html and pic.png
	directives       []string
	proxyEnabled     bool
	proxyDirectives  []string
	contents         map[string][]byte
}

var (
	casketForwardProxy            casketTestServer
	casketForwardProxyAuth        casketTestServer // requires auth
	casketHTTPForwardProxyAuth    casketTestServer // requires auth, does not use TLS
	casketForwardProxyProbeResist casketTestServer // requires auth, and has probing resistance on
	casketDummyProbeResist        casketTestServer // same as casketForwardProxyProbeResist, but w/o forwardproxy

	casketForwardProxyWhiteListing        casketTestServer
	casketForwardProxyBlackListing        casketTestServer
	casketForwardProxyNoBlacklistOverride casketTestServer // to test default blacklist

	// authenticated server upstreams to authenticated https proxy with different credentials
	casketAuthedUpstreamEnter casketTestServer

	casketTestTarget     casketTestServer // whitelisted by casketForwardProxyWhiteListing
	casketHTTPTestTarget casketTestServer // serves plain http on 6480
)

func (c *casketTestServer) marshal() []byte {
	mainBlock := []string{c.addr + " {",
		"root " + c.root}
	mainBlock = append(mainBlock, c.directives...)
	if c.proxyEnabled {
		if len(c.proxyDirectives) == 0 {
			mainBlock = append(mainBlock, "forwardproxy")
		} else {
			forwardProxyBlock := []string{"forwardproxy {"}
			forwardProxyBlock = append(forwardProxyBlock, strings.Join(c.proxyDirectives, "\n"))
			forwardProxyBlock = append(forwardProxyBlock, "}")
			mainBlock = append(mainBlock, strings.Join(forwardProxyBlock, "\n"))
		}
	}
	mainBlock = append(mainBlock, "}")
	if len(c.HTTPRedirectPort) > 0 {
		// TODO: this is not good enough, since `func redirPlaintextHost(cfg *SiteConfig) *SiteConfig`
		// https://github.com/tmpim/casket/blob/master/caskethttp/httpserver/https.go#L142 can change in future
		// and we won't know.
		redirectBlock := []string{"http://*:" + c.HTTPRedirectPort + " {",
			"redir https://" + c.addr + "{uri}",
			"header / Connection close",
			"}"}
		mainBlock = append(mainBlock, redirectBlock...)
	}
	return []byte(strings.Join(mainBlock, "\n"))
}

func (c *casketTestServer) StartTestServer() {
	var err error
	c.Instance, err = casket.Start(casket.CasketfileInput{Contents: c.marshal(), ServerTypeName: "http"})
	if err != nil {
		panic(err)
	}
	if c.contents == nil {
		c.contents = make(map[string][]byte)
	}
	index, err := ioutil.ReadFile(c.root + "/index.html")
	if err != nil {
		panic(err)
	}
	c.contents[""] = index
	c.contents["/"] = index
	c.contents["/index.html"] = index

	c.contents["/pic.png"], err = ioutil.ReadFile(c.root + "/pic.png")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	casketForwardProxy = casketTestServer{addr: "127.0.19.84:1984", root: "./test/forwardproxy",
		directives:   []string{"tls self_signed {\nno_redirect\n}"},
		proxyEnabled: true, proxyDirectives: []string{"serve_pac",
			"acl {\nallow all\n}"}}
	casketForwardProxy.StartTestServer()

	casketForwardProxyAuth = casketTestServer{addr: "127.0.0.1:4891", root: "./test/forwardproxy",
		directives:   []string{"tls self_signed {\nno_redirect\n}"},
		proxyEnabled: true, proxyDirectives: []string{"basicauth test pass",
			"acl {\nallow all\n}"}}
	casketForwardProxyAuth.StartTestServer()

	casketHTTPForwardProxyAuth = casketTestServer{addr: "127.0.69.73:6973", root: "./test/forwardproxy",
		directives:   []string{"tls off"},
		proxyEnabled: true, proxyDirectives: []string{"basicauth test pass",
			"acl {\nallow all\n}"}}
	casketHTTPForwardProxyAuth.StartTestServer()

	casketForwardProxyProbeResist = casketTestServer{addr: "127.0.88.88:8888", root: "./test/forwardproxy",
		directives: []string{"tls self_signed {\nno_redirect\n}"}, HTTPRedirectPort: "8880",
		proxyEnabled: true, proxyDirectives: []string{"basicauth test pass",
			"probe_resistance test.localhost",
			"serve_pac superhiddenfile.pac",
			"acl {\nallow all\n}"}}
	casketForwardProxyProbeResist.StartTestServer()

	casketDummyProbeResist = casketTestServer{addr: "127.0.99.99:9999", root: "./test/forwardproxy",
		directives: []string{"tls self_signed {\nno_redirect\n}"}, HTTPRedirectPort: "9980",
		proxyEnabled: false}
	casketDummyProbeResist.StartTestServer()

	// 127.0.0.1 and localhost are both used to avoid Casket matching and routing proxy requests internally
	casketTestTarget = casketTestServer{addr: "127.0.64.51:6451", root: "./test/index",
		directives:   []string{},
		proxyEnabled: false}
	casketTestTarget.StartTestServer()

	casketHTTPTestTarget = casketTestServer{addr: "localhost:6480", root: "./test/index",
		directives:   []string{"tls off"},
		proxyEnabled: false}
	casketHTTPTestTarget.StartTestServer()

	casketAuthedUpstreamEnter = casketTestServer{addr: "127.0.65.25:6585", root: "./test/upstreamingproxy",
		directives:   []string{"tls self_signed {\nno_redirect\n}"},
		proxyEnabled: true, proxyDirectives: []string{"upstream https://test:pass@127.0.0.1:4891",
			"basicauth upstreamtest upstreampass"}}
	casketAuthedUpstreamEnter.StartTestServer()

	casketForwardProxyWhiteListing = casketTestServer{addr: "127.0.87.76:8776", root: "./test/forwardproxy",
		directives:   []string{"tls self_signed {\nno_redirect\n}"},
		proxyEnabled: true, proxyDirectives: []string{"acl {\nallow 127.0.64.51\n deny all\n}",
			"ports 6451"}}
	casketForwardProxyWhiteListing.StartTestServer()

	casketForwardProxyBlackListing = casketTestServer{addr: "127.0.66.76:6676", root: "./test/forwardproxy",
		directives:   []string{"tls self_signed {\nno_redirect\n}"},
		proxyEnabled: true, proxyDirectives: []string{"acl {\ndeny " + blacklistedIPv4 + "/30\n" +
			"deny " + blacklistedIPv6 + "\nallow all\n}"},
	}
	casketForwardProxyBlackListing.StartTestServer()

	casketForwardProxyNoBlacklistOverride = casketTestServer{addr: "127.0.66.79:6679", root: "./test/forwardproxy",
		directives:   []string{"tls self_signed {\nno_redirect\n}"},
		proxyEnabled: true, proxyDirectives: []string{}}
	casketForwardProxyNoBlacklistOverride.StartTestServer()

	retCode := m.Run()

	casketForwardProxy.Stop()
	casketForwardProxyAuth.Stop()
	casketHTTPForwardProxyAuth.Stop()
	casketForwardProxyProbeResist.Stop()
	casketDummyProbeResist.Stop()
	casketTestTarget.Stop()
	casketHTTPTestTarget.Stop()
	casketAuthedUpstreamEnter.Stop()
	casketForwardProxyWhiteListing.Stop()
	casketForwardProxyBlackListing.Stop()
	casketForwardProxyNoBlacklistOverride.Stop()

	os.Exit(retCode)
}

// This is a sanity check confirming that target servers actually directly serve what they are expected to.
// (And that they don't serve what they should not)
func TestTheTest(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 2 * time.Second}

	// Request index
	resp, err := client.Get("http://" + casketTestTarget.addr)
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketTestTarget.contents[""]); err != nil {
		t.Fatal(err)
	}

	// Request pic
	resp, err = client.Get("http://" + casketTestTarget.addr + "/pic.png")
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketTestTarget.contents["/pic.png"]); err != nil {
		t.Fatal(err)
	}

	// Request pic, but expect index. Should fail
	resp, err = client.Get("http://" + casketTestTarget.addr + "/pic.png")
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketTestTarget.contents[""]); err == nil {
		t.Fatal(err)
	}

	// Request index, but expect pic. Should fail
	resp, err = client.Get("http://" + casketTestTarget.addr)
	if err != nil {
		t.Fatal(err)
	} else if err = responseExpected(resp, casketTestTarget.contents["/pic.png"]); err == nil {
		t.Fatal(err)
	}

	// Request non-existing resource
	resp, err = client.Get("http://" + casketTestTarget.addr + "/idontexist")
	if err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected: 404 StatusNotFound, got %d. Response: %#v\n", resp.StatusCode, resp)
	}
}

func debugIoCopy(dst io.Writer, src io.Reader, prefix string) (written int64, err error) {
	buf := make([]byte, 32*1024)
	flusher, ok := dst.(http.Flusher)
	for {
		nr, er := src.Read(buf)
		fmt.Printf("[%s] Read err %#v\n%s", prefix, er, hex.Dump(buf[0:nr]))
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if ok {
				flusher.Flush()
			}
			fmt.Printf("[%s] Wrote %v %v\n", prefix, nw, ew)
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	fmt.Printf("[%s] Returning with %#v %#v\n", prefix, written, err)
	return
}

func httpdump(r interface{}) string {
	switch v := r.(type) {
	case *http.Request:
		if v == nil {
			return "httpdump: nil"
		}
		b, err := httputil.DumpRequest(v, true)
		if err != nil {
			return err.Error()
		} else {
			return string(b)
		}
	case *http.Response:
		if v == nil {
			return "httpdump: nil"
		}
		b, err := httputil.DumpResponse(v, true)
		if err != nil {
			return err.Error()
		} else {
			return string(b)
		}
	default:
		return "httpdump: wrong type"
	}
}
