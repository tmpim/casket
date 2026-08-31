// Package gandi adapts the libdns Gandi DNS provider
// for Casket. Importing this package plugs it in.
package gandi

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/gandi"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "GANDI_"

	EnvAPIKey = envNamespace + "API_KEY"
)

func init() {
	caskettls.RegisterDNSProvider("gandi", NewDNSProvider)
}

// NewDNSProvider returns a new Gandi DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment
// len(1): credentials[0] = API key
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &gandi.Provider{
		BearerToken: env.GetOrDefaultString(EnvAPIKey, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ key ... }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.BearerToken = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 1:
		provider.BearerToken = credentials[0]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.BearerToken == "" {
		return nil, errors.New("gandi: missing credentials")
	}

	return provider, nil
}
