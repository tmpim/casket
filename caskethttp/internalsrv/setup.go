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

package internalsrv

import (
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

func init() {
	casket.RegisterPlugin("internal", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// Internal configures a new Internal middleware instance.
func setup(c *casket.Controller) error {
	paths, err := internalParse(c)
	if err != nil {
		return err
	}

	// Append Internal paths to Casket config HiddenFiles to ensure
	// files do not appear in Browse
	config := httpserver.GetConfig(c)
	config.HiddenFiles = append(config.HiddenFiles, paths...)

	config.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return Internal{Next: next, Paths: paths}
	})

	return nil
}

func internalParse(c *casket.Controller) ([]string, error) {
	var paths []string

	for c.Next() {
		if c.NextArg() {
			paths = append(paths, c.Val())
		}
		if c.NextArg() {
			return nil, c.ArgErr()
		}
	}

	return paths, nil
}
