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

package browse

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/inhies/go-bytesize"
	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
	"github.com/tmpim/casket/caskethttp/staticfiles"
)

func init() {
	casket.RegisterPlugin("browse", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new Browse middleware instance.
func setup(c *casket.Controller) error {
	configs, err := browseParse(c)
	if err != nil {
		return err
	}

	b := Browse{
		Configs:       configs,
		IgnoreIndexes: false,
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		b.Next = next
		return b
	})

	return nil
}

func browseParse(c *casket.Controller) ([]Config, error) {
	var configs []Config

	cfg := httpserver.GetConfig(c)

	appendCfg := func(bc Config) error {
		for _, c := range configs {
			if c.PathScope == bc.PathScope {
				return fmt.Errorf("duplicate browsing config for %s", c.PathScope)
			}
		}
		configs = append(configs, bc)
		return nil
	}

	for c.Next() {
		var bc Config
		bc.BufferSize = 10 * 1024 * 1024 // 10 MB

		args := c.RemainingArgs()

		// First argument is directory to allow browsing; default is site root
		if len(args) > 0 {
			bc.PathScope = args[0]
		} else {
			bc.PathScope = "/"
		}

		bc.Fs = staticfiles.FileServer{
			Root:       http.Dir(cfg.Root),
			Hide:       cfg.HiddenFiles,
			IndexPages: cfg.IndexPages,
		}

		// Second argument would be the template file to use
		var tplText string
		if len(args) > 1 {
			tplBytes, err := ioutil.ReadFile(args[1])
			if err != nil {
				return configs, err
			}
			tplText = string(tplBytes)
		} else {
			tplText = defaultTemplate
		}

		// If nested block is present, process it here
		for c.NextBlock() {
			switch c.Val() {
			case "path":
				if c.NextArg() {
					bc.PathScope = c.Val()
				} else {
					return configs, c.ArgErr()
				}
			case "tplfile":
				if c.NextArg() {
					tplBytes, err := ioutil.ReadFile(c.Val())
					if err != nil {
						return configs, err
					}
					tplText = string(tplBytes)
				} else {
					return configs, c.ArgErr()
				}
			case "servearchive":
				types := make([]ArchiveType, 0)

				for c.NextArg() {
					archiveType := ArchiveType(c.Val())
					if _, found := ArchiveTypeToMime[archiveType]; found {
						types = append(types, archiveType)
					} else {
						return configs, c.Errf("invalid archive type: %s", archiveType)
					}
				}

				if len(types) == 0 {
					bc.ArchiveTypes = ArchiveTypes
				} else {
					bc.ArchiveTypes = types
				}
			case "buffer":
				bufSizeStr := strings.Join(c.RemainingArgs(), " ")
				size, err := bytesize.Parse(bufSizeStr)
				if err != nil {
					return configs, c.Errf("error parsing buffer size: %v", err)
				}

				bc.BufferSize = uint64(size)
			default:
				return configs, c.Errf("unknown property '%s'", c.Val())
			}
		}

		// Build the template
		tpl, err := template.New("listing").Parse(tplText)
		if err != nil {
			return configs, err
		}
		bc.Template = tpl

		// Save configuration
		err = appendCfg(bc)
		if err != nil {
			return configs, err
		}
	}

	return configs, nil
}

// The default template to use when serving up directory listings
//
//go:embed default_template.html
var defaultTemplate string
