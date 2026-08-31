// Package rfc2136 adapts the libdns RFC 2136 dynamic update DNS
// provider for Casket. Importing this package plugs it in.
package rfc2136

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/rfc2136"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls"
	"github.com/tmpim/casket/caskettls/env"
)

const (
	envNamespace = "RFC2136_"

	EnvTSIGKey       = envNamespace + "TSIG_KEY"
	EnvTSIGSecret    = envNamespace + "TSIG_SECRET"
	EnvTSIGAlgorithm = envNamespace + "TSIG_ALGORITHM"
	EnvNameserver    = envNamespace + "NAMESERVER"
)

func init() {
	caskettls.RegisterDNSProvider("rfc2136", NewDNSProvider)
}

// NewDNSProvider returns a new RFC 2136 DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(4): credentials[0] = nameserver
// ------- credentials[1] = TSIG algorithm
// ------- credentials[2] = TSIG key
// ------- credentials[3] = TSIG secret
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &rfc2136.Provider{
		KeyName: env.GetOrDefaultString(EnvTSIGKey, ""),
		Key:     env.GetOrDefaultString(EnvTSIGSecret, ""),
		KeyAlg:  env.GetOrDefaultString(EnvTSIGAlgorithm, ""),
		Server:  env.GetOrDefaultString(EnvNameserver, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// check the configuration block for options { key ..., secret ..., algorithm ..., nameserver ... }
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.KeyName = c.Val()
			case "secret":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Key = c.Val()
			case "algorithm":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.KeyAlg = c.Val()
			case "nameserver":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Server = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 4:
		provider.Server = credentials[0]
		provider.KeyAlg = credentials[1]
		provider.KeyName = credentials[2]
		provider.Key = credentials[3]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.Server == "" {
		return nil, errors.New("missing nameserver")
	}

	if provider.KeyName == "" || provider.Key == "" || provider.KeyAlg == "" {
		return nil, errors.New("missing TSIG key configuration")
	}

	return provider, nil
}
