// Package pdns adapts the libdns PowerDNS
// provider for Casket. Importing this package plugs it in.
package pdns

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/powerdns"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls"
	"github.com/tmpim/casket/caskettls/env"
)

const (
	envNamespace = "PDNS_"

	EnvAPIKey   = envNamespace + "API_KEY"
	EnvAPIURL   = envNamespace + "API_URL"
	EnvServerID = envNamespace + "SERVER_ID"
)

func init() {
	caskettls.RegisterDNSProvider("powerdns", NewDNSProvider)
}

// NewDNSProvider returns a new PowerDNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(2): credentials[0] = PowerDNS API URL,
// ------- credentials[1] = PowerDNS API key
// len(3): credentials[0] = PowerDNS API URL,
// ------- credentials[1] = PowerDNS API key
// ------- credentials[2] = PowerDNS server id (if empty, defaults to localhost
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &powerdns.Provider{
		ServerURL: env.GetOrDefaultString(EnvAPIURL, ""),
		APIToken:  env.GetOrDefaultString(EnvAPIKey, ""),
		ServerID:  env.GetOrDefaultString(EnvServerID, "localhost"),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ api_url ..., api_key ..., server_id ... }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "api_url":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.ServerURL = c.Val()
			case "api_key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.APIToken = c.Val()
			case "server_id":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.ServerID = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 3:
		provider.ServerID = credentials[2]
		fallthrough
	case 2:
		provider.ServerURL = credentials[0]
		provider.APIToken = credentials[1]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.ServerURL == "" || provider.APIToken == "" {
		return nil, errors.New("powerdns: missing credentials")
	}

	return provider, nil
}
