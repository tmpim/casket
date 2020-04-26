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

package requestid

import (
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

func init() {
	casket.RegisterPlugin("request_id", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *casket.Controller) error {
	var headerName string

	for c.Next() {
		if c.NextArg() {
			headerName = c.Val()
		}
		if c.NextArg() {
			return c.ArgErr()
		}
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return Handler{Next: next, HeaderName: headerName}
	})

	return nil
}
