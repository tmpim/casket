// Package ovh adapts the libdns OVH DNS
// provider for Casket. Importing this package plugs it in.
package ovh

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/ovh"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "OVH_"

	EnvEndpoint          = envNamespace + "ENDPOINT"
	EnvApplicationKey    = envNamespace + "APPLICATION_KEY"
	EnvApplicationSecret = envNamespace + "APPLICATION_SECRET"
	EnvConsumerKey       = envNamespace + "CONSUMER_KEY"
)

func init() {
	caskettls.RegisterDNSProvider("ovh", NewDNSProvider)
}

// NewDNSProvider returns a new OVH DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(4): credentials[0] = API Endpoint
// ------- credentials[1] = Application Key
// ------- credentials[2] = Application Secret
// ------- credentials[3] = Consumer Key
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &ovh.Provider{
		Endpoint:          env.GetOrDefaultString(EnvEndpoint, ""),
		ApplicationKey:    env.GetOrDefaultString(EnvApplicationKey, ""),
		ApplicationSecret: env.GetOrDefaultString(EnvApplicationSecret, ""),
		ConsumerKey:       env.GetOrDefaultString(EnvConsumerKey, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ endpoint ..., application_key ..., application_secret ..., consumer_key ... }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "endpoint":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Endpoint = c.Val()
			case "application_key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.ApplicationKey = c.Val()
			case "application_secret":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.ApplicationSecret = c.Val()
			case "consumer_key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.ConsumerKey = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 4:
		provider.Endpoint = credentials[0]
		provider.ApplicationKey = credentials[1]
		provider.ApplicationSecret = credentials[2]
		provider.ConsumerKey = credentials[3]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.Endpoint == "" || provider.ApplicationKey == "" || provider.ApplicationSecret == "" || provider.ConsumerKey == "" {
		return nil, errors.New("ovh: missing credentials")
	}

	return provider, nil
}
