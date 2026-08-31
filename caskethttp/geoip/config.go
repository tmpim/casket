package geoip

import (
	"github.com/tmpim/casket"
)

// Config specifies configuration parsed for Casketfile
type Config struct {
	DatabasePath string
}

func parseConfig(c *casket.Controller) (Config, error) {
	var config = Config{}
	for c.Next() {
		value := c.Val()
		switch value {
		case "geoip":
			if !c.NextArg() {
				continue
			}
			config.DatabasePath = c.Val()
		}
	}
	return config, nil
}
