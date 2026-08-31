// Package dnspod adapts the libdns Dnspod DNS
// provider for Casket. Importing this package plugs it in.
package dnspod

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/dnspod"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "DNSPOD_"

	EnvAPIKey = envNamespace + "API_KEY"
)

func init() {
	caskettls.RegisterDNSProvider("dnspod", NewDNSProvider)
}

// NewDNSProvider returns a new dnspod DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment
// len(1): credentials[0] = access token (API key)
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &dnspod.Provider{
		APIToken: env.GetOrDefaultString(EnvAPIKey, ""),
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
		return nil, errors.New("dnspod: missing credentials")
	}

	return provider, nil
}
