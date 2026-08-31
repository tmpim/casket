package chuieauth

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

type ChuieAuth struct {
	Next         httpserver.Handler
	domainRules  []domainRule
	exceptions   []string
	authOptional bool
}

type domainRule struct {
	basePath string
	authId   string
}

var errAuthFail = errors.New("chuieauth: authentication failure")

func (c *ChuieAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) (int,
	error) {
	r.Header.Del("X-Auth-Username")
	r.Header.Del("X-Auth-Permitted")
	r.Header.Del("X-Auth-Internal-Error")

	if len(c.domainRules) == 0 && !c.authOptional {
		return c.authMux(w, r)
	}

	for _, domain := range c.exceptions {
		if httpserver.Path(r.URL.Path).Matches(domain) {
			return c.Next.ServeHTTP(w, r)
		}
	}

	var pathMatch struct {
		length int
		authId string
	}

	for _, domain := range c.domainRules {
		if len(domain.basePath) <= pathMatch.length {
			continue
		}

		if httpserver.Path(r.URL.Path).Matches(domain.basePath) {
			pathMatch.length = len(domain.basePath)
			pathMatch.authId = domain.authId
		}
	}

	if pathMatch.length == 0 && !c.authOptional {
		return c.Next.ServeHTTP(w, r)
	}

	return c.authenticateSessionAndReturn(w, r, pathMatch.authId)
}

func init() {
	casket.RegisterPlugin("chuieauth", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *casket.Controller) error {
	domainRules, optional, exceptions, err := parseRules(c)
	if err != nil {
		return err
	}

	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return &ChuieAuth{Next: next, domainRules: domainRules,
			exceptions: exceptions, authOptional: optional}
	}
	cfg.AddMiddleware(mid)

	return nil
}

func parseRules(c *casket.Controller) ([]domainRule, bool, []string, error) {
	var domainRules []domainRule
	var exceptions []string

	for c.Next() {
		args := c.RemainingArgs()

		if len(args) < 1 && len(domainRules) == 0 {
			return parseConfigBlock(c)
		} else if len(args) < 1 {
			continue
		}

		authId := args[0]

		if len(authId) != 3 {
			return []domainRule{}, false, []string{}, errors.New(
				"chuieauth: auth ID must be exactly 3 characters long")
		}

		if len(args) > 1 {
			for _, basePath := range args[1:] {
				domainRules = append(domainRules, domainRule{basePath, authId})
			}
		} else {
			domainRules = append(domainRules, domainRule{"/", authId})
		}

		for c.NextBlock() {
			switch c.Val() {
			case "except":
				exceptions = append(exceptions, c.RemainingArgs()...)
			}
		}
	}

	return domainRules, false, exceptions, nil
}

func parseConfigBlock(c *casket.Controller) ([]domainRule,
	bool, []string, error) {
	matchingBlock := false
	databaseHost := ""
	databaseKey := ""

	for c.NextBlock() {
		matchingBlock = true
		switch c.Val() {
		case "database_host":
			databaseHost = c.RemainingArgs()[0]
		case "database_password":
			databaseKey = c.RemainingArgs()[0]
		case "encryption_key":
			setupSessionStore(c.RemainingArgs()[0])
		}
	}

	if len(databaseHost) > 0 {
		var err error
		for i := 0; i < 5; i++ {
			log.Println("chuieauth: connecting to database...")
			err := connectToDatabase(databaseHost, databaseKey)
			if err != nil {
				log.Println("chuieauth: database connection failed:", err)
			}
			time.Sleep(time.Second * 2)
			if err == nil {
				log.Println("chuieauth: database connected!")
				break
			}
		}

		if err != nil {
			return []domainRule{}, true, []string{}, err
		}
	}

	if !matchingBlock {
		return []domainRule{}, true, []string{}, nil
	}

	// TODO: Change back to protocol independent
	rawHostName := c.Key
	rawHostName = strings.TrimPrefix(rawHostName, "http://")
	rawHostName = strings.TrimPrefix(rawHostName, "http//")
	rawHostName = strings.TrimPrefix(rawHostName, "https://")
	rawHostName = strings.TrimPrefix(rawHostName, "https//")
	authHostBase = "https://" + rawHostName
	return []domainRule{}, false, []string{}, nil
}
