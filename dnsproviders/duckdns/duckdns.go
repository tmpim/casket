// Package duckdns adapts the libdns duckdns DNS
// provider for Casket. Importing this package plugs it in.
package duckdns

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/duckdns"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "DUCKDNS_"

	EnvToken = envNamespace + "TOKEN"
)

func init() {
	caskettls.RegisterDNSProvider("duckdns", NewDNSProvider)
}

// NewDNSProvider returns a new duckdns DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment
// len(1): credentials[0] = duckdns token
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &duckdns.Provider{
		APIToken: env.GetOrDefaultString(EnvToken, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ token ... }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "token":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.APIToken = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 1:
		provider.APIToken = credentials[0]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.APIToken == "" {
		return nil, errors.New("duckdns: missing credentials")
	}

	return provider, nil
}
