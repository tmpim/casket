// Package transip adapts the libdns TransIP DNS
// provider for Casket. Importing this package plugs it in.
package transip

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/transip"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "TRANSIP_"

	EnvAccountName    = envNamespace + "ACCOUNT_NAME"
	EnvPrivateKeyPath = envNamespace + "PRIVATE_KEY_PATH"
)

func init() {
	caskettls.RegisterDNSProvider("transip", NewDNSProvider)
}

// NewDNSProvider returns a new TransIP DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment
// len(2): credentials[0] = Account Name
// ------- credentials[1] = Private Key Path
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &transip.Provider{
		AccountName:    env.GetOrDefaultString(EnvAccountName, ""),
		PrivateKeyPath: env.GetOrDefaultString(EnvPrivateKeyPath, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// check the configuration block for options { account_name ..., private_key_path ... }
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "account_name":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.AccountName = c.Val()
			case "private_key_path":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.PrivateKeyPath = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 2:
		provider.AccountName = credentials[0]
		provider.PrivateKeyPath = credentials[1]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.AccountName == "" {
		return nil, errors.New("account name is missing")
	}

	if provider.PrivateKeyPath == "" {
		return nil, errors.New("private key path is missing")
	}

	return provider, nil
}
