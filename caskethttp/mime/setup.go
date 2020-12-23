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

package mime

import (
	"fmt"
	"strings"

	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

func init() {
	casket.RegisterPlugin("mime", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new mime middleware instance.
func setup(c *casket.Controller) error {
	config, err := mimeParse(c)
	if err != nil {
		return err
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return Mime{Next: next, Configs: config}
	})

	return nil
}

func mimeParse(c *casket.Controller) (Config, error) {
	config := Config{
		UseDefaults: false,
		Extensions:  make(map[string]string),
	}

	for c.Next() {
		// At least one extension is required

		args := c.RemainingArgs()
		switch len(args) {
		case 2:
			if err := validateExt(config, args[0]); err != nil {
				return config, err
			}
			config.Extensions[args[0]] = args[1]
		case 1:
			if args[0] == "ext_defaults" {
				config.UseDefaults = true
			}

			return config, c.ArgErr()
		case 0:
			for c.NextBlock() {
				ext := c.Val()
				if ext == "ext_defaults" {
					config.UseDefaults = true
				}

				if err := validateExt(config, ext); err != nil {
					return config, err
				}
				if !c.NextArg() {
					return config, c.ArgErr()
				}

				config.Extensions[ext] = c.Val()
			}
		}

	}

	return config, nil
}

// validateExt checks for valid file name extension.
func validateExt(config Config, ext string) error {
	if !strings.HasPrefix(ext, ".") {
		return fmt.Errorf(`mime: invalid extension "%v" (must start with dot)`, ext)
	}
	if _, ok := config.Extensions[ext]; ok {
		return fmt.Errorf(`mime: duplicate extension "%v" found`, ext)
	}
	return nil
}
