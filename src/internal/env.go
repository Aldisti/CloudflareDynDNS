package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	env *Environment = nil
)

const (
	ENV_ZONE_ID   = "ZONE_ID"
	ENV_API_TOKEN = "API_TOKEN"
	ENV_DOMAIN    = "DOMAIN"
	ENV_INTERVAL  = "INTERVAL"
	ENV_MAX_FAILS = "MAX_FAILURES"
	ENV_TIMEOUT   = "TIMEOUT"
)

type Environment struct {
	ZoneId   string
	ApiToken string
	Domain   string
	Interval int
	MaxFails int
	Timeout  int
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
	if err := setEnvVar(ENV_ZONE_ID, func(s string) { env.ZoneId = s }); err != nil {
		return env, err
	}
	if err := setEnvVar(ENV_API_TOKEN, func(s string) { env.ApiToken = s }); err != nil {
		return env, err
	}
	if err := setEnvVar(ENV_DOMAIN, func(s string) { env.Domain = s }); err != nil {
		return env, err
	}
	if err := setEnvVarInt(ENV_INTERVAL, func(n int) { env.Interval = n }); err != nil {
		env.Interval = 30
		fmt.Printf("Using default value %d for %s\n", env.Interval, ENV_INTERVAL)
	}
	if err := setEnvVarInt(ENV_MAX_FAILS, func(n int) { env.MaxFails = n }); err != nil {
		env.Interval = 10
		fmt.Printf("Using default value %d for %s\n", env.MaxFails, ENV_MAX_FAILS)
	}
	if err := setEnvVarInt(ENV_TIMEOUT, func(n int) { env.Timeout = n }); err != nil {
		env.Timeout = 1
		fmt.Printf("Using default value %d for %s\n", env.Timeout, ENV_TIMEOUT)
	}

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
