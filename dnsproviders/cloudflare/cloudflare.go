// Package cloudflare adapts the libdns Cloudflare DNS
// provider for Casket. Importing this package plugs it in.
package cloudflare

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/cloudflare"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls"
	"github.com/tmpim/casket/caskettls/env"
	"strings"
)

const tokenErr = "cloudflare: email and API tokens are no longer supported in Casket, please use Scoped Tokens only. " +
	"More info: https://pkg.go.dev/github.com/libdns/cloudflare#readme-authenticating"

func init() {
	caskettls.RegisterDNSProvider("cloudflare", NewDNSProvider)
}

// NewDNSProvider returns a new Cloudflare DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment, or block (`{ token ... }`)
// len(1): credentials[0] = Scoped API token
// len(2): credentials[0] = token type (must be "zonetoken")
// ------- credentials[1] = Scoped API token
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &cloudflare.Provider{}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from environment variables.
		values, err := env.GetWithFallback([]string{
			"CLOUDFLARE_ZONE_API_TOKEN",
			"CF_ZONE_API_TOKEN",
			"CLOUDFLARE_DNS_API_TOKEN",
			"CF_DNS_API_TOKEN",
		})
		if err != nil {
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
		} else {
			// Use the first available token environment variable
			provider.APIToken = values["CLOUDFLARE_ZONE_API_TOKEN"]
		}
	case 1:
		provider.APIToken = credentials[0]
	case 2:
		if strings.Contains(credentials[0], "@") {
			return nil, errors.New(tokenErr)
		}

		switch credentials[0] {
		case "zonetoken":
			provider.APIToken = credentials[1]
		default:
			return nil, errors.New(tokenErr)
		}
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.APIToken == "" {
		return nil, errors.New("cloudflare: missing credentials")
	}

	return provider, nil
}
