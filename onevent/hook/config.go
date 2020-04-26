package hook

import (
	"github.com/tmpim/casket"
)

// Config describes how Hook should be configured and used.
type Config struct {
	ID      string
	Event   casket.EventName
	Command string
	Args    []string
}

// SupportedEvents is a map of supported events.
var SupportedEvents = map[string]casket.EventName{
	"startup":   casket.InstanceStartupEvent,
	"shutdown":  casket.ShutdownEvent,
	"certrenew": casket.CertRenewEvent,
}
