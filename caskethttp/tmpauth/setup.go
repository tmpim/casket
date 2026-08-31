package tmpauth

import (
	"crypto/hmac"
	"crypto/sha1"
	"hash"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

const (
	TmpAuthHost = "auth.tmpim.pw"
)

type Tmpauth struct {
	Next       httpserver.Handler
	Config     *Config
	Logger     *log.Logger
	TokenCache map[[32]byte]*CachedToken
	HttpClient *http.Client
	HMAC       hash.Hash

	stateIDCache    *cache.Cache
	tokenCacheMutex *sync.RWMutex
	hmacMutex       *sync.Mutex
	janitorOnce     *sync.Once
	done            chan struct{}
}

func init() {
	casket.RegisterPlugin("tmpauth", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *casket.Controller) error {
	config, err := parseConfig(c)
	if err != nil {
		return err
	}

	cfg := httpserver.GetConfig(c)

	if config.Debug {
		backgroundWorker.EnableDebug()
	}

	done := make(chan struct{})

	mid := func(next httpserver.Handler) httpserver.Handler {
		return &Tmpauth{
			Next:   next,
			Config: config,
			Logger: log.New(os.Stderr, "tmpauth", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
			HttpClient: &http.Client{
				Transport: &Transport{
					config: config,
					base:   http.DefaultTransport,
				},
			},
			TokenCache:      make(map[[32]byte]*CachedToken),
			HMAC:            hmac.New(sha1.New, config.Secret),
			hmacMutex:       new(sync.Mutex),
			tokenCacheMutex: new(sync.RWMutex),
			stateIDCache:    cache.New(time.Minute*5, time.Minute),
			janitorOnce:     new(sync.Once),
			done:            done,
		}
	}
	cfg.AddMiddleware(mid)

	// stop background jobs on reloads/shutdowns
	c.OnShutdown(func() error {
		close(done)
		return nil
	})

	return nil
}
