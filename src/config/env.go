package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	env *Environment = nil
)

const (
	MODE_POLLER   = "POLLER"
	MODE_LISTENER = "LISTENER"
)

const (
	ENV_MODE      = "MODE"
	ENV_API_TOKEN = "API_TOKEN"
	ENV_DOMAIN    = "DOMAIN"
	ENV_TIMEOUT   = "TIMEOUT"
	ENV_INTERVAL  = "INTERVAL"
	ENV_MAX_FAILS = "MAX_FAILURES"
	ENV_COOLDOWN  = "COOLDOWN"
)

type Environment struct {
	Mode     string
	ApiToken string
	Domains  []string
	Interval int
	MaxFails int
	Timeout  time.Duration
	Cooldown time.Duration
}

func GetEnv() *Environment {
	if env, err := GetEnvSafe(); err != nil {
		panic(err)
	} else {
		return env
	}
}

func GetEnvSafe() (*Environment, error) {
	if env == nil {
		if e, err := loadEnvironment(); err != nil {
			return nil, fmt.Errorf("Couldn't load env vars: %s", err)
		} else {
			env = &e
		}
	}
	return env, nil
}

func loadEnvironment() (Environment, error) {
	env := Environment{}

	if err := setEnvVar(ENV_MODE, func(s string) { env.Mode = s }); err != nil {
		env.Mode = MODE_POLLER
	}
	if err := setEnvVar(ENV_API_TOKEN, func(s string) { env.ApiToken = s }); err != nil {
		return env, err
	}

	if err := setEnvVar(ENV_DOMAIN, func(s string) { env.Domains = strings.Split(s, ",") }); err != nil {
		return env, err
	}

	setEnvVarIntDefault(ENV_INTERVAL, 60, func(n int) { env.Interval = n })
	setEnvVarIntDefault(ENV_MAX_FAILS, -1, func(n int) { env.MaxFails = n })

	setEnvVarIntDefault(ENV_TIMEOUT, 2, func(n int) {
		env.Timeout = time.Duration(n) * time.Second
	})
	setEnvVarIntDefault(ENV_COOLDOWN, -1, func(n int) {
		env.Cooldown = time.Duration(n) * time.Second
	})

	return env, nil
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func setEnvVar(name string, consumer func(string)) error {
	if s, ok := os.LookupEnv(name); !ok || isBlank(s) {
		return fmt.Errorf("Error: missing env variable: %s", name)
	} else {
		consumer(s)
		return nil
	}
}

func setEnvVarIntDefault(name string, def int, consumer func(int)) {
	if err := setEnvVarInt(name, consumer); err != nil {
		consumer(def)
		fmt.Printf("Using default value %d for %s\n", def, name)
	}
}

func setEnvVarInt(name string, consumer func(int)) error {
	s, ok := os.LookupEnv(name)
	if !ok || isBlank(s) {
		return fmt.Errorf("Error: missing env variable: %s", name)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("Error: variable %s must be a valid int: %s", name, err)
	}
	consumer(n)
	return nil
}
