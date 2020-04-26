package onevent

import (
	"strings"

	"github.com/tmpim/casket"
	"github.com/tmpim/casket/onevent/hook"
	"github.com/google/uuid"
)

func init() {
	// Register Directive.
	casket.RegisterPlugin("on", casket.Plugin{Action: setup})
}

func setup(c *casket.Controller) error {
	config, err := onParse(c)
	if err != nil {
		return err
	}

	// Register Event Hooks.
	err = c.OncePerServerBlock(func() error {
		for _, cfg := range config {
			casket.RegisterEventHook("on-"+cfg.ID, cfg.Hook)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func onParse(c *casket.Controller) ([]*hook.Config, error) {
	var config []*hook.Config

	for c.Next() {
		cfg := new(hook.Config)

		if !c.NextArg() {
			return config, c.ArgErr()
		}

		// Configure Event.
		event, ok := hook.SupportedEvents[strings.ToLower(c.Val())]
		if !ok {
			return config, c.Errf("Wrong event name or event not supported: '%s'", c.Val())
		}
		cfg.Event = event

		// Assign an unique ID.
		cfg.ID = uuid.New().String()

		args := c.RemainingArgs()

		// Extract command and arguments.
		command, args, err := casket.SplitCommandAndArgs(strings.Join(args, " "))
		if err != nil {
			return config, c.Err(err.Error())
		}

		cfg.Command = command
		cfg.Args = args

		config = append(config, cfg)
	}

	return config, nil
}
