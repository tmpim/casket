// Copyright 2015 Light Code Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package basicauth

import (
	"fmt"
	"strings"
	"time"

	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

func init() {
	casket.RegisterPlugin("basicauth", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new BasicAuth middleware instance.
func setup(c *casket.Controller) error {
	cfg := httpserver.GetConfig(c)
	root := cfg.Root

	rules, cookie, err := basicAuthParse(c)
	if err != nil {
		return err
	}

	basic := BasicAuth{Rules: rules, Cookie: cookie}

	cfg.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		basic.Next = next
		basic.SiteRoot = root
		return basic
	})

	return nil
}

func basicAuthParse(c *casket.Controller) ([]Rule, CookieConfig, error) {
	var rules []Rule
	cfg := httpserver.GetConfig(c)
	cookie := CookieConfig{
		Enabled:   false,
		TTL:       defaultCookieTTL,
		Refresh:   0,
		NoRefresh: false,
		Name:      defaultCookieName,
	}

	var err error
	for c.Next() {
		var rule Rule

		args := c.RemainingArgs()

		switch len(args) {
		case 2:
			rule.Username = args[0]
			if rule.Password, err = passwordMatcher(rule.Username, args[1], cfg.Root); err != nil {
				return rules, cookie, c.Errf("Get password matcher from %s: %v", c.Val(), err)
			}
		case 3:
			rule.Resources = append(rule.Resources, args[0])
			rule.Username = args[1]
			if rule.Password, err = passwordMatcher(rule.Username, args[2], cfg.Root); err != nil {
				return rules, cookie, c.Errf("Get password matcher from %s: %v", c.Val(), err)
			}
		default:
			return rules, cookie, c.ArgErr()
		}

		// If nested block is present, process it here
		for c.NextBlock() {
			val := c.Val()
			args = c.RemainingArgs()
			switch len(args) {
			case 0:
				if val == "cookie" {
					if cookie.Enabled {
						return rules, cookie, c.Errf("\"cookie\" subdirective can only be specified once")
					}

					cookie.Enabled = true

					for nesting := c.Nesting(); c.NextBlockNesting(nesting); {
						switch c.Val() {
						case "ttl":
							if !c.NextArg() {
								return rules, cookie, c.ArgErr()
							}

							cookie.TTL, err = time.ParseDuration(c.Val())
							if err != nil {
								return rules, cookie, fmt.Errorf("could not parse cookie.ttl: %v", err)
							}
						case "refresh":
							if !c.NextArg() {
								return rules, cookie, c.ArgErr()
							}

							cookie.Refresh, err = time.ParseDuration(c.Val())
							if err != nil {
								return rules, cookie, fmt.Errorf("could not parse cookie.refresh: %v", err)
							}

							if cookie.Refresh == 0 {
								cookie.NoRefresh = true
							}
						case "name":
							if !c.NextArg() {
								return rules, cookie, c.ArgErr()
							}

							cookie.Name = c.Val()
						}
					}

					if cookie.Refresh == 0 && !cookie.NoRefresh {
						// refresh wasn't configured, set it to half ttl
						cookie.Refresh = cookie.TTL / 2
					}
				} else {
					// Assume single argument is path resource
					rule.Resources = append(rule.Resources, val)
				}
			case 1:
				if val == "realm" {
					if rule.Realm == "" {
						rule.Realm = strings.Replace(args[0], `"`, `\"`, -1)
					} else {
						return rules, cookie, c.Errf("\"realm\" subdirective can only be specified once")
					}
				} else if val == "exclude" {
					rule.Exclude = append(rule.Exclude, args[0])
				} else {
					return rules, cookie, c.Errf("expecting \"realm\", \"exclude\", or \"cookie\", got \"%s\"", val)
				}
			default:
				return rules, cookie, c.ArgErr()
			}
		}

		rules = append(rules, rule)
	}

	return rules, cookie, nil
}

func passwordMatcher(username, passw, siteRoot string) (PasswordMatcher, error) {
	htpasswdPrefix := "htpasswd="
	if !strings.HasPrefix(passw, htpasswdPrefix) {
		return PlainMatcher(passw), nil
	}
	return GetHtpasswdMatcher(passw[len(htpasswdPrefix):], username, siteRoot)
}
