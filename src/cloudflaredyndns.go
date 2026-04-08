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
	Records      []cloudflare.Record
	CurrentIP   string
	LastUpdate  time.Time
	Failures    int
	LastFailure time.Time
}

func main() {
	env := config.GetEnv()

	switch env.Mode {
	case config.MODE_POLLER:
		poller.Run()
	case config.MODE_LISTENER:
		listener.Run()
	default:
		panic(fmt.Errorf("Invalid mode '%s'", env.Mode))
	}
}
