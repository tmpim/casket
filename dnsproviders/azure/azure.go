// Package azure adapts the libdns azure DNS
// provider for Casket. Importing this package plugs it in.
package azure

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/azure"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"

	"github.com/tmpim/casket/caskettls"
)

func init() {
	caskettls.RegisterDNSProvider("azure", NewDNSProvider)
}

// NewDNSProvider returns a new azure DNS challenge provider.
// See underlying package docs for details: https://github.com/libdns/azure
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	// Try to get credentials from environment variables.
	provider := &azure.Provider{
		TenantId:          env.GetOrDefaultString("AZURE_TENANT_ID", ""),
		ClientId:          env.GetOrDefaultString("AZURE_CLIENT_ID", ""),
		ClientSecret:      env.GetOrDefaultString("AZURE_CLIENT_SECRET", ""),
		SubscriptionId:    env.GetOrDefaultString("AZURE_SUBSCRIPTION_ID", ""),
		ResourceGroupName: env.GetOrDefaultString("AZURE_RESOURCE_GROUP", ""),
	}

	credentials := c.RemainingArgs()
	if len(credentials) != 0 {
		return nil, errors.New("invalid credentials length")
	}

	// Or, override configuration from the config block
	for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
		switch c.Val() {
		case "tenant_id":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			provider.TenantId = c.Val()
		case "client_id":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			provider.ClientId = c.Val()
		case "client_secret":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			provider.ClientSecret = c.Val()
		case "subscription_id":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			provider.SubscriptionId = c.Val()
		case "resource_group":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			provider.ResourceGroupName = c.Val()
		default:
			return nil, c.Errf("unknown property '%s'", c.Val())
		}
	}

	return provider, nil
}
