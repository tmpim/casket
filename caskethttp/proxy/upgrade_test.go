package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tmpim/casket/casketfile"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

func TestUpgradePresets(t *testing.T) {
	for _, preset := range []string{"upgrade", "websocket", "websocket\nupgrade", "upgrade\nwebsocket"} {
		t.Run(preset, func(t *testing.T) {
			config := "proxy / localhost:8080 {\n" + preset + "\n}"
			upstreams, err := NewStaticUpstreams(casketfile.NewDispenser("Testfile", strings.NewReader(config)), "")
			if err != nil {
				t.Fatal(err)
			}
			host := upstreams[0].Select(httptest.NewRequest("GET", "/", nil))
			if host.AllowAllUpgrades != strings.Contains(preset, "upgrade") {
				t.Fatalf("unexpected upgrade policy: %v", host.AllowAllUpgrades)
			}
			headers := host.UpstreamHeaders
			for _, name := range []string{"Connection", "Upgrade"} {
				if got := headers.Get(name); got != "{>"+name+"}" {
					t.Fatalf("%s = %q", name, got)
				}
			}
		})
	}
}

func TestRequestIsUpgrade(t *testing.T) {
	for _, tc := range []struct {
		connection []string
		protocol   string
		want       bool
	}{
		{[]string{"keep-alive, UpGrAdE"}, "DERP", true},
		{[]string{"keep-alive", "upgrade"}, "tailscale-control-protocol", true},
		{[]string{"not-upgrade"}, "websocket", false},
		{[]string{"upgrade"}, "   ", false},
		{nil, "websocket", false},
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header["Connection"] = tc.connection
		r.Header.Set("Upgrade", tc.protocol)
		if got := requestIsUpgrade(r); got != tc.want {
			t.Errorf("Connection=%v Upgrade=%q: got %v, want %v", tc.connection, tc.protocol, got, tc.want)
		}
	}
}

func TestUpgradeProtocolOffers(t *testing.T) {
	header := http.Header{"Upgrade": {"websocket, tailscale-control-protocol", "DERP"}}
	for _, protocol := range []string{"websocket", "tailscale-control-protocol", "derp"} {
		if !headerHasToken(header, "Upgrade", protocol) {
			t.Errorf("offered protocol %q was not recognized", protocol)
		}
	}
	for _, protocol := range []string{"socket", "unknown", "websocket, tailscale-control-protocol"} {
		if headerHasToken(header, "Upgrade", protocol) {
			t.Errorf("unoffered protocol %q was accepted", protocol)
		}
	}
}

// Exercise the real proxy over TCP, including bytes buffered on both sides of
// the HTTP handshake and continued bidirectional traffic after the upgrade.
func TestProtocolUpgradeTunnel(t *testing.T) {
	for _, tc := range []struct{ preset, protocol string }{
		{"upgrade", "websocket"},
		{"upgrade", "tailscale-control-protocol"},
		{"upgrade", "DERP"},
		{"websocket", "websocket"},
	} {
		t.Run(tc.preset+"/"+tc.protocol, func(t *testing.T) {
			protocol := tc.protocol
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.Header.Get("Upgrade") != protocol {
					t.Errorf("unexpected upstream request: %s %v", r.Method, r.Header)
				}
				conn, rw, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(5 * time.Second))
				fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: %s\r\n\r\nhello", protocol)
				rw.Flush()
				io.Copy(conn, rw)
			}))
			defer backend.Close()
			p := newUpgradeTestProxy(t, backend.URL, tc.preset)
			front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := p.ServeHTTP(w, r); err != nil {
					t.Error(err)
				}
			}))
			defer front.Close()
			conn, err := net.Dial("tcp", front.Listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			fmt.Fprintf(conn, "POST /ts2021 HTTP/1.1\r\nHost: example.com\r\nConnection: keep-alive, Upgrade\r\nUpgrade: %s\r\nContent-Length: 0\r\n\r\nearly", protocol)
			reader := bufio.NewReader(conn)
			res, err := http.ReadResponse(reader, nil)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != 101 || res.Header.Get("Upgrade") != protocol {
				t.Fatalf("unexpected response: %s %v", res.Status, res.Header)
			}
			for _, want := range []string{"helloearly", "later"} {
				if want == "later" {
					fmt.Fprint(conn, want)
				}
				got := make([]byte, len(want))
				if _, err := io.ReadFull(reader, got); err != nil {
					t.Fatal(err)
				}
				if string(got) != want {
					t.Fatalf("got %q, want %q", got, want)
				}
			}
		})
	}
}

func TestUpgradeResponseValidation(t *testing.T) {
	for _, tc := range []struct {
		name, offered, selected, connection string
		status                              int
		reject                              bool
	}{
		{"unsolicited", "", "DERP", "Upgrade", 101, true},
		{"mismatch", "DERP", "websocket", "Upgrade", 101, true},
		{"missing protocol", "DERP", "", "Upgrade", 101, true},
		{"missing connection", "DERP", "DERP", "", 101, true},
		{"substring", "DERP", "DERP", "not-upgrade", 101, true},
		{"refused", "DERP", "", "close", 403, false},
		{"ordinary", "", "", "close", 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Upgrade", tc.selected)
				w.Header().Set("Connection", tc.connection)
				w.WriteHeader(tc.status)
			}))
			defer backend.Close()
			p := newUpgradeTestProxy(t, backend.URL, "upgrade")
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Connection", "Upgrade")
			r.Header.Set("Upgrade", tc.offered)
			w := httptest.NewRecorder()
			_, err := p.ServeHTTP(w, r)
			if (err != nil) != tc.reject {
				t.Fatalf("error = %v, reject = %v", err, tc.reject)
			}
			if !tc.reject && w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
		})
	}
}

func newUpgradeTestProxy(t *testing.T, backend, preset string) *Proxy {
	t.Helper()
	config := "proxy / " + backend + " {\n" + preset + "\n}"
	upstreams, err := NewStaticUpstreams(casketfile.NewDispenser("Testfile", strings.NewReader(config)), "")
	if err != nil {
		t.Fatal(err)
	}
	return &Proxy{Next: httpserver.EmptyNext, Upstreams: upstreams}
}

func TestWebsocketUpgradeRestriction(t *testing.T) {
	for _, preset := range []string{"websocket", "header_upstream Connection {>Connection}\nheader_upstream Upgrade {>Upgrade}"} {
		for _, protocol := range []string{"DERP", "tailscale-control-protocol", "websocket, DERP", "ordinary"} {
			t.Run(preset+"/"+protocol, func(t *testing.T) {
				var called atomic.Bool
				backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called.Store(true)
					if protocol != "ordinary" && r.Header.Get("Upgrade") != "websocket" {
						t.Errorf("backend received unrestricted offer: %q", r.Header.Get("Upgrade"))
					}
					w.WriteHeader(http.StatusNoContent)
				}))
				defer backend.Close()
				p := newUpgradeTestProxy(t, backend.URL, preset)
				r := httptest.NewRequest("GET", "/", nil)
				if protocol != "ordinary" {
					r.Header.Set("Connection", "Upgrade")
					r.Header.Set("Upgrade", protocol)
				}
				w := httptest.NewRecorder()
				code, err := p.ServeHTTP(w, r)
				reject := protocol == "DERP" || protocol == "tailscale-control-protocol"
				if reject {
					if code != http.StatusBadRequest || err == nil || called.Load() {
						t.Fatalf("code=%d err=%v backend called=%v", code, err, called.Load())
					}
				} else if err != nil || w.Code != http.StatusNoContent || !called.Load() {
					t.Fatalf("status=%d err=%v backend called=%v", w.Code, err, called.Load())
				}
			})
		}
	}
}
