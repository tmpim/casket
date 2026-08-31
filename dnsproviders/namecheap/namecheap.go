// Package namecheap adapts the libdns NameCheap DNS provider
// for Casket. Importing this package plugs it in.
package namecheap

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/namecheap"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "NAMECHEAP_"

	EnvAPIUser = envNamespace + "API_USER"
	EnvAPIKey  = envNamespace + "API_KEY"
	EnvSandbox = envNamespace + "SANDBOX"

	sandboxBaseURL = "https://api.sandbox.namecheap.com/xml.response"
)

func init() {
	caskettls.RegisterDNSProvider("namecheap", NewDNSProvider)
}

// NewDNSProvider returns a new NameCheap DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or configuration block
// len(2): credentials[0] = API user
// ------- credentials[1] = API key
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	sandbox := env.GetOrDefaultBool(EnvSandbox, false)
	provider := &namecheap.Provider{
		APIKey: env.GetOrDefaultString(EnvAPIKey, ""),
		User:   env.GetOrDefaultString(EnvAPIUser, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ username ..., password ..., shared_secret ..., sandbox }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "username":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.User = c.Val()
			case "key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.APIKey = c.Val()
			case "sandbox":
				if c.NextArg() {
					return nil, c.ArgErr() // Only one argument is allowed
				}
				sandbox = true
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 2:
		provider.User = credentials[0]
		provider.APIKey = credentials[1]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if sandbox {
		provider.APIEndpoint = sandboxBaseURL
	}

	if provider.User == "" || provider.APIKey == "" {
		return nil, errors.New("namecheap: credentials missing")
	}

	return provider, nil
}
