// Package googlecloud adapts the libdns Google Cloud DNS
// provider for Casket. Importing this package plugs it in.
package googlecloud

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/googleclouddns"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "GCE_"

	EnvServiceAccount = envNamespace + "SERVICE_ACCOUNT"
	EnvProject        = envNamespace + "PROJECT"
)

func init() {
	caskettls.RegisterDNSProvider("googlecloud", NewDNSProvider)
}

// NewDNSProvider returns a new Google Cloud DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment
// len(1): credentials[0] = project_id
// len(2): credentials[0] = project_id
// ------- credentials[1] = service_account_json_path
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &googleclouddns.Provider{
		Project:            env.GetOrDefaultString(EnvProject, ""),
		ServiceAccountJSON: env.GetOrDefaultString(EnvServiceAccount, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// Try to get credentials from the block (`{ project ..., service_account ... }`)
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "project":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Project = c.Val()
			case "service_account":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.ServiceAccountJSON = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	case 1:
		provider.Project = credentials[0]
	case 2:
		provider.Project = credentials[0]
		provider.ServiceAccountJSON = credentials[1]
	default:
		return nil, errors.New("invalid credentials length")
	}

	if provider.Project == "" {
		return nil, errors.New("googlecloud: missing project")
	}

	if provider.ServiceAccountJSON == "" {
		return nil, errors.New("googlecloud: missing service_account")
	}

	return provider, nil
}
