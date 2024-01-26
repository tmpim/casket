package caskettls

import (
	"errors"
	"fmt"
	"github.com/caddyserver/certmagic"
	"github.com/libdns/cloudflare"
	"github.com/tmpim/casket/caskettls/env"
	"strings"
)

const tokenErr = "cloudflare: email and API tokens are no longer supported in Casket, please use Scoped Tokens only. " +
	"More info: https://pkg.go.dev/github.com/libdns/cloudflare#readme-authenticating"

func init() {
	RegisterDNSProvider("cloudflare", func(credentials ...string) (certmagic.ACMEDNSProvider, error) {
		switch len(credentials) {
		case 0:
			values, err := env.GetWithFallback([]string{
				"CLOUDFLARE_ZONE_API_TOKEN",
				"CF_ZONE_API_TOKEN",
				"CLOUDFLARE_DNS_API_TOKEN",
				"CF_DNS_API_TOKEN",
			})
			if err != nil {
				return nil, fmt.Errorf("cloudflare: %v", err)
			}

			return &cloudflare.Provider{APIToken: values["CLOUDFLARE_ZONE_API_TOKEN"]}, nil
		case 1:
			return &cloudflare.Provider{APIToken: credentials[0]}, nil
		case 2:
			if strings.Contains(credentials[0], "@") {
				return nil, errors.New(tokenErr)
			}

			switch credentials[0] {
			case "zonetoken":
				return &cloudflare.Provider{APIToken: credentials[1]}, nil
			default:
				return nil, errors.New(tokenErr)
			}
		default:
			return nil, errors.New("invalid credentials length")
		}
	})
}
