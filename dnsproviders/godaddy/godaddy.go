// Package godaddy adapts the libdns GoDaddy DNS
// provider for Casket. Importing this package plugs it in.
package godaddy

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/godaddy"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "GODADDY_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"
)

func init() {
	caskettls.RegisterDNSProvider("godaddy", NewDNSProvider)
}

// NewDNSProvider returns a new GoDaddy DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(1): credentials[0] = API key + API secret joined by a colon
// len(2): credentials[0] = API key
// ------- credentials[1] = API secret
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &godaddy.Provider{}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ key ..., secret ... }`)
		key, secret := "", ""
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				key = c.Val()
			case "secret":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				secret = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}

		if key != "" && secret != "" {
			provider.APIToken = key + ":" + secret
		} else if key != "" && secret == "" {
			provider.APIToken = key
		} else {
			// Try to get credentials from the environment.
			values, err := env.Get(EnvAPIKey, EnvAPISecret)
			if err != nil {
				return nil, err
			}

			provider.APIToken = values[EnvAPIKey] + ":" + values[EnvAPISecret]
		}
	case 1:
		provider.APIToken = credentials[0]
	case 2:
		provider.APIToken = credentials[0] + ":" + credentials[1]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.APIToken == "" {
		return nil, errors.New("godaddy: missing credentials")
	}

	return provider, nil
}
