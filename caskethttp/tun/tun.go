package tun

import (
	"encoding/gob"
	"net"
)

func init() {
	gob.Register(&net.TCPAddr{})
	gob.Register(&net.UDPAddr{})
	gob.Register(&net.IPAddr{})
}

type Config struct {
	Secret   string
	Insecure bool
	Upstream string
}

func GetConfig(addr string) *Config {
	return &Config{
		Insecure: false,
		Upstream: "wss://test.home.chuie.io/.well-known/tun",
		Secret:   "secret",
	}
}
