// Package acmedns adapts the libdns exec DNS
// provider for Casket. Importing this package plugs it in.
package acmedns

import (
	"encoding/json"
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/acmedns"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls"
	"github.com/tmpim/casket/caskettls/env"
	"io/ioutil"
)

const (
	// envNamespace is the prefix for ACME-DNS environment variables.
	envNamespace = "ACME_DNS_"

	// EnvAPIBase is the environment variable name for the ACME-DNS API address.
	// (e.g. https://acmedns.your-domain.com).
	EnvAPIBase = envNamespace + "API_BASE"
	// EnvStoragePath is the environment variable name for the ACME-DNS JSON account data file.
	// A per-domain account will be registered/persisted to this file and used for TXT updates.
	EnvStoragePath = envNamespace + "STORAGE_PATH"
)

func init() {
	caskettls.RegisterDNSProvider("acmedns", NewDNSProvider)
}

// NewDNSProvider returns a new acmedns DNS challenge provider.
// The credentials are interpreted as follows:
//
// len(0): use credentials from environment or block:
//
//	        tls dns acmedns {
//	            server_url <server_url>
//	            storage /path/to/storage.json
//	            username <username>
//	            password <password>
//		        subdomain <subdomain>
//	        }
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		configPath := env.GetOrDefaultString(EnvStoragePath, "")
		provider := &acmedns.Provider{
			// Try to get credentials from environment variables.
			ServerURL: env.GetOrDefaultString(EnvAPIBase, ""),
		}

		// Try to get credentials from the config block.
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "server_url":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
			case "storage":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				configPath = c.Val()
			case "username":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Username = c.Val()
			case "password":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Password = c.Val()
			case "subdomain":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Subdomain = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}

		if configPath != "" {
			provider.Configs = unmarshalAccountDataFile(configPath)
		}

		return provider, nil
	default:
		return nil, errors.New("invalid credentials length")
	}
}

func unmarshalAccountDataFile(path string) map[string]acmedns.DomainConfig {
	accounts := make(map[string]acmedns.DomainConfig)
	// Opportunistically try to load the account data. Return an empty account if
	// any errors occur.
	if jsonData, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(jsonData, &accounts); err != nil {
			return accounts
		}
	}
	return accounts
}
