package browse

import (
	"net/http"
	"path"
	"strings"

	"github.com/tmpim/casket"
	"github.com/tmpim/casket/caskethttp/httpserver"
	"github.com/tmpim/casket/caskethttp/rewrite"
	"github.com/tmpim/casket/caskethttp/staticfiles"
)

func init() {
	casket.RegisterPlugin("tryfiles", casket.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type Config struct {
	To      string
	Except  []string
	Without string
}

type TryFiles struct {
	Next    httpserver.Handler
	FileSys http.FileSystem
	Config  *Config
}

func (t *TryFiles) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for _, p := range t.Config.Except {
		if strings.HasPrefix(r.URL.Path, p) {
			return t.Next.ServeHTTP(w, r)
		}
	}

	if t.Config.Without != "" {
		rewrite.To(t.FileSys, r, t.Config.To, httpserver.NewReplacer(r, nil, ""), t.Config.Without)
	} else {
		rewrite.To(t.FileSys, r, t.Config.To, httpserver.NewReplacer(r, nil, ""))
	}

	return t.Next.ServeHTTP(w, r)
}

// setup configures a new Browse middleware instance.
func setup(c *casket.Controller) error {
	config, err := tryFilesParse(c)
	if err != nil {
		return err
	}

	cfg := httpserver.GetConfig(c)

	b := &TryFiles{
		Config:  config,
		FileSys: http.Dir(cfg.Root),
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		b.Next = next
		return b
	})

	return nil
}

func tryFilesParse(c *casket.Controller) (*Config, error) {
	cfg := httpserver.GetConfig(c)

	config := &Config{
		To:      "{path} " + strings.Join(staticfiles.DefaultIndexPages, " "),
		Except:  []string{"/.well-known"},
		Without: cfg.Addr.Path,
	}

	if config.Without == "/" {
		config.Without = ""
	}

	if config.Without != "" {
		config.Except = append(config.Except, path.Join(config.Without, "/.well-known"))
	}

	for c.Next() {
		tryFileArgs := c.RemainingArgs()
		if len(tryFileArgs) == 0 {
			continue
		}

		config.To = strings.Join(tryFileArgs, " ")

		for c.NextBlock() {
			val := c.Val()
			args := c.RemainingArgs()

			switch val {
			case "except":
				config.Except = args
			case "without":
				if len(args) != 1 {
					return nil, c.Err("`without` directive must have exactly one argument")
				}

				config.Without = args[0]
			default:
				return nil, c.ArgErr()
			}
		}
	}

	return config, nil
}
