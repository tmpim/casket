// Package digitalocean adapts the libdns DigitalOcean DNS
// provider for Casket. Importing this package plugs it in.
package digitalocean

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/digitalocean"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "DO_"

	EnvAuthToken = envNamespace + "AUTH_TOKEN"
)

func init() {
	caskettls.RegisterDNSProvider("digitalocean", NewDNSProvider)
}

// NewDNSProvider returns a new DigitalOcean DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(1): credentials[0] = access token (API key)
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &digitalocean.Provider{
		APIToken: env.GetOrDefaultString(EnvAuthToken, ""),
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
		return nil, errors.New("digitalocean: missing credentials")
	}

	return provider, nil
}
