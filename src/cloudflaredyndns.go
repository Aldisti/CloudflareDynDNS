package main

import (
	"fmt"
	"time"

	"github.com/Aldisti/CloudflareDynDNS/cloudflare"
	"github.com/Aldisti/CloudflareDynDNS/config"
	"github.com/Aldisti/CloudflareDynDNS/listener"
	"github.com/Aldisti/CloudflareDynDNS/poller"
)

type Context struct {
	Env         *config.Environment
	Records     []cloudflare.Record
	CurrentIP   string
	LastUpdate  time.Time
	Failures    int
	LastFailure time.Time
}

func main() {
	env := config.GetEnv()

	switch env.Mode {
	case config.MODE_POLLER:
		poller.Run(env)
	case config.MODE_LISTENER:
		listener.Run(env)
	default:
		panic(fmt.Errorf(
			"Invalid mode '%s', valid values are: %s and %s",
			env.Mode, config.MODE_POLLER, config.MODE_POLLER,
		))
	}
}
