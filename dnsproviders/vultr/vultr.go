// Package vultr adapts the libdns Vultr DNS provider
// for Casket. Importing this package plugs it in.
package vultr

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/vultr"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "VULTR_"

	EnvAPIKey = envNamespace + "API_KEY"
)

func init() {
	caskettls.RegisterDNSProvider("vultr", NewDNSProvider)
}

// NewDNSProvider returns a new Vultr DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(1): credentials[0] = API key
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &vultr.Provider{
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
		return nil, errors.New("vultr: missing credentials")
	}

	return provider, nil
}
