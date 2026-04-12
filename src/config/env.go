package config

import (
	"fmt"
	"os"

	"github.com/Aldisti/CloudflareDynDNS/common"
)

var env *Environment = nil

// Constant values
const (
	MODE_POLLER   = "POLLER"
	MODE_LISTENER = "LISTENER"
)

// Environment variables names
const (
	ENV_MODE      = "MODE"
	ENV_API_TOKEN = "API_TOKEN"
	ENV_TIMEOUT   = "TIMEOUT"

	ENV_DOMAINS   = "DOMAINS"
	ENV_INTERVAL  = "INTERVAL"
	ENV_MAX_FAILS = "MAX_FAILURES"
	ENV_COOLDOWN  = "COOLDOWN"

	ENV_ADDRESS  = "ADDRESS"
	ENV_PORT     = "PORT"
	ENV_USERNAME = "USERNAME"
	ENV_PASSWORD = "PASSWORD"
)

type Environment struct {
	Mode     string
	ApiToken string
	Timeout  string

	// Poller mode-only variables

	Domains  string
	Interval string
	MaxFails string
	Cooldown string

	// Listener mode-only variables

	Address  string
	Port     string
	Username string
	Password string
}

func init() {
	e := loadEnvironment()
	env = &e
}

func GetEnv() *Environment {
	return env
}

func loadEnvironment() Environment {
	env := Environment{
		Timeout:  "5",
		Interval: "60",
		MaxFails: "-1",
		Cooldown: "-1",
		Address:  "0.0.0.0",
		Port:     "8080",
	}

	setEnvVar(ENV_MODE, &env.Mode)
	setEnvVar(ENV_API_TOKEN, &env.ApiToken)
	setEnvVar(ENV_DOMAINS, &env.Domains)
	setEnvVar(ENV_TIMEOUT, &env.Timeout)
	setEnvVar(ENV_INTERVAL, &env.Interval)
	setEnvVar(ENV_MAX_FAILS, &env.MaxFails)
	setEnvVar(ENV_COOLDOWN, &env.Cooldown)
	setEnvVar(ENV_ADDRESS, &env.Address)
	setEnvVar(ENV_PORT, &env.Port)
	setEnvVar(ENV_USERNAME, &env.Username)
	setEnvVar(ENV_PASSWORD, &env.Password)

	if common.IsBlank(env.Mode) {
		panic(fmt.Errorf("Missing env var: %s", ENV_MODE))
	}
	if common.IsBlank(env.ApiToken) {
		panic(fmt.Errorf("Missing env var: %s", ENV_API_TOKEN))
	}

	return env
}

func setEnvVar(name string, toSet *string) {
	if s, ok := os.LookupEnv(name); ok && !common.IsBlank(s) {
		*toSet = s
	}
}
