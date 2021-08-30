package httpserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/staticfiles"
	"github.com/tmpim/casket/caskethttp/tun"
)

type TunServer struct {
	Server *http.Server
	site   *SiteConfig
}

var _ casket.GracefulServer = new(TunServer)

func NewTunServer(site *SiteConfig) (*TunServer, error) {
	stack := Handler(staticfiles.FileServer{Root: http.Dir(site.Root), Hide: site.HiddenFiles, IndexPages: site.IndexPages})
	for i := len(site.middleware) - 1; i >= 0; i-- {
		stack = site.middleware[i](stack)
	}
	site.middlewareChain = stack

	server := makeHTTPServerWithTimeouts(site.Addr.Original, []*SiteConfig{site})
	s := &TunServer{
		Server: server,
		site:   site,
	}
	server.Handler = s

	return s, nil
}

func (s *TunServer) Listen() (net.Listener, error) {
	return s.listenInternal()
}

func (s *TunServer) listenInternal() (net.Listener, error) {
	base := strings.TrimSuffix(s.site.Addr.Original, "tun://")
	cfg := tun.GetConfig(s.site.Addr.Original)
	conn, _, err := websocket.DefaultDialer.Dial(cfg.Upstream, http.Header{
		"Authorization": []string{"Bearer " + cfg.Secret},
	})
	if err != nil {
		return nil, err
	}
}

func (s *TunServer)

func (s *TunServer) Serve(listener net.Listener) error {
	return s.Server.Serve(listener)
}

func (s *TunServer) Stop() error {
	return nil
}

func (s *TunServer) OnStartupComplete() {
	fmt.Println("todo TunServer: OnStartupComplete")
}

func (s *TunServer) Address() string {
	return s.site.Addr.String()
}

func (s *TunServer) WrapListener(listener net.Listener) net.Listener {
	return listener
}

func (s *TunServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// We absolutely need to be sure we stay alive up here,
		// even though, in theory, the errors middleware does this.
		if rec := recover(); rec != nil {
			log.Printf("[PANIC] %v", rec)
			DefaultErrorFunc(w, r, http.StatusInternalServerError)
		}
	}()

	// copy the original, unchanged URL into the context
	// so it can be referenced by middlewares
	urlCopy := *r.URL
	if r.URL.User != nil {
		userInfo := new(url.Userinfo)
		*userInfo = *r.URL.User
		urlCopy.User = userInfo
	}
	c := context.WithValue(r.Context(), OriginalURLCtxKey, urlCopy)
	r = r.WithContext(c)

	// Setup a replacer for the request that keeps track of placeholder
	// values across plugins.
	replacer := NewReplacer(r, nil, "")
	c = context.WithValue(r.Context(), ReplacerCtxKey, replacer)
	r = r.WithContext(c)

	w.Header().Set("Server", casket.AppName)

	status, _ := s.site.middlewareChain.ServeHTTP(w, r)

	// Fallback error response in case error handling wasn't chained in
	if status >= 400 {
		DefaultErrorFunc(w, r, status)
	}
}

func (s *TunServer) ListenPacket() (net.PacketConn, error) {
	return nil, nil
}

func (s *TunServer) ServePacket(_ net.PacketConn) error {
	return nil
}
