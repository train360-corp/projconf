package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type EnvironmentVariable string

const (
	PROJCONF_SUPABASE_URL      EnvironmentVariable = "PROJCONF_SUPABASE_URL"
	PROJCONF_SUPABASE_ANON_KEY EnvironmentVariable = "PROJCONF_SUPABASE_ANON_KEY"
)

type runtimeEnvironment struct {
	values map[EnvironmentVariable]string
}

var (
	instance *runtimeEnvironment
	once     sync.Once
)

// getConfig returns the singleton instance of envConfig
func getConfig() *runtimeEnvironment {
	once.Do(func() {
		instance = &runtimeEnvironment{
			values: map[EnvironmentVariable]string{
				PROJCONF_SUPABASE_URL:      os.Getenv(string(PROJCONF_SUPABASE_URL)),
				PROJCONF_SUPABASE_ANON_KEY: os.Getenv(string(PROJCONF_SUPABASE_ANON_KEY)),
			},
		}
		mustLoad()
	})
	return instance
}

func mustLoadVariable(variable EnvironmentVariable) {
	switch variable {
	case PROJCONF_SUPABASE_URL:
	case PROJCONF_SUPABASE_ANON_KEY:
		if strings.TrimSpace(os.Getenv(string(variable))) == "" {
			panic(fmt.Sprintf("environment variable %s is required", variable))
		}
	default:
		panic(fmt.Sprintf("[mustLoad] unhandled environment variable: %s", variable))
	}
}

// mustLoad requires certain environment variables to be set
func mustLoad() {
	mustLoadVariable(PROJCONF_SUPABASE_URL)
	mustLoadVariable(PROJCONF_SUPABASE_ANON_KEY)
}

func MustLoad() {
	getConfig()
}

// Get returns the value of the environment variable
func Get(key EnvironmentVariable) string {
	return getConfig().values[key]
}
