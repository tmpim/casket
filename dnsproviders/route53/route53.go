// Package route53 adapts the libdns Route53 DNS
// provider for Casket. Importing this package plugs it in.
package route53

import (
	"errors"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/route53"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskettls/env"
	"strconv"

	"github.com/tmpim/casket/caskettls"
)

const (
	envNamespace = "AWS_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"
	EnvRegion          = envNamespace + "REGION"
	EnvMaxRetries      = envNamespace + "MAX_RETRIES"
	EnvAWSProfile      = envNamespace + "PROFILE"
)

func init() {
	caskettls.RegisterDNSProvider("route53", NewDNSProvider)
}

// NewDNSProvider returns a new Route53 DNS challenge provider.
// The credentials are detected automatically; see underlying
// package docs for details:
// https://github.com/libdns/route53
func NewDNSProvider(c *casket.Controller) (certmagic.ACMEDNSProvider, error) {
	provider := &route53.Provider{
		AccessKeyId:     env.GetOrDefaultString(EnvAccessKeyID, ""),
		SecretAccessKey: env.GetOrDefaultString(EnvSecretAccessKey, ""),
		Region:          env.GetOrDefaultString(EnvRegion, ""),
		MaxRetries:      env.GetOrDefaultInt(EnvMaxRetries, 5),
		AWSProfile:      env.GetOrDefaultString(EnvAWSProfile, ""),
	}

	credentials := c.RemainingArgs()

	switch len(credentials) {
	case 0:
		// check the configuration block for options { access_key_id ..., secret_access_key ..., region ..., max_retries ..., aws_profile ... }
		for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
			switch c.Val() {
			case "access_key_id":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.AccessKeyId = c.Val()
			case "secret_access_key":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.SecretAccessKey = c.Val()
			case "region":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.Region = c.Val()
			case "max_retries":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				n, err := strconv.Atoi(c.Val())
				if err != nil {
					return nil, err
				}
				provider.MaxRetries = n
			case "aws_profile":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				provider.AWSProfile = c.Val()
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	default:
		return nil, errors.New("invalid credentials length")
	}

	// AWS SDK may load its own credentials from somewhere, so we shouldn't need to verify here?
	// if provider.AccessKeyId == "" || provider.SecretAccessKey == "" || provider.Region == "" {
	// 	return nil, errors.New("missing credentials")
	// }

	return provider, nil
}
