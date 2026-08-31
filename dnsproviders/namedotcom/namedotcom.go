// Package namedotcom adapts the libdns Name.com DNS provider
// for Casket. Importing this package plugs it in.
package namedotcom

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/namedotcom"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "NAMECOM_"

	EnvUsername = envNamespace + "USERNAME"
	EnvAPIToken = envNamespace + "API_TOKEN"
	EnvServer   = envNamespace + "SERVER"
)

func init() {
	caskettls.RegisterDNSProvider("namedotcom", NewDNSProvider)
}

// NewDNSProvider returns a new Name.com DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(2): credentials[0] = username
// ------- credentials[1] = API token
// len(3): credentials[0] = username
// ------- credentials[1] = API token
// ------- credentials[2] = Server
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &namedotcom.Provider{
		User:  env.GetOrDefaultString(EnvUsername, ""),
		Token: env.GetOrDefaultString(EnvAPIToken, ""),
	}

	server, err := env.Get(EnvServer)
	if err == nil {
		provider.Server = server[EnvServer]
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ username ..., token ..., server ... }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "username":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.User = c.Val()
			case "token":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Token = c.Val()
			case "server":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Server = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 3:
		provider.Server = credentials[2]
		fallthrough
	case 2:
		provider.User = credentials[0]
		provider.Token = credentials[1]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.User == "" || provider.Token == "" {
		return nil, errors.New("namedotcom: missing credentials")
	}

	return provider, nil
}
