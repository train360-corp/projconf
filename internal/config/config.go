package config

import (
	"fmt"
	"os"
	"strings"
)

type EnvironmentVariable string

const (
	PROJCONF_POSTGRES_PASSWORD    EnvironmentVariable = "PROJCONF_POSTGRES_PASSWORD"
	PROJCONF_JWT_SECRET           EnvironmentVariable = "PROJCONF_JWT_SECRET"
	PROJCONF_SUPABASE_URL         EnvironmentVariable = "PROJCONF_SUPABASE_URL"
	PROJCONF_SUPABASE_ANON_KEY    EnvironmentVariable = "PROJCONF_SUPABASE_ANON_KEY"
	PROJCONF_SUPABASE_SERVICE_KEY EnvironmentVariable = "PROJCONF_SUPABASE_SERVICE_KEY"
)

type RuntimeEnvironment map[EnvironmentVariable]string

// GetConfig returns the singleton instance of envConfig
func GetConfig() RuntimeEnvironment {
	return RuntimeEnvironment{
		PROJCONF_POSTGRES_PASSWORD:    os.Getenv(string(PROJCONF_POSTGRES_PASSWORD)),
		PROJCONF_JWT_SECRET:           os.Getenv(string(PROJCONF_JWT_SECRET)),
		PROJCONF_SUPABASE_URL:         os.Getenv(string(PROJCONF_SUPABASE_URL)),
		PROJCONF_SUPABASE_ANON_KEY:    os.Getenv(string(PROJCONF_SUPABASE_ANON_KEY)),
		PROJCONF_SUPABASE_SERVICE_KEY: os.Getenv(string(PROJCONF_SUPABASE_SERVICE_KEY)),
	}
}

func MustLoad(variable EnvironmentVariable, variables ...EnvironmentVariable) {

	vars := []EnvironmentVariable{variable}
	vars = append(vars, variables...)
	config := GetConfig()

	for _, envVar := range vars {
		if strings.TrimSpace(config[envVar]) == "" {
			panic(fmt.Sprintf("environment variable '%s' is required but not set", envVar))
		}
	}
}

// Get returns the value of the environment variable
func Get(key EnvironmentVariable) string {
	return GetConfig()[key]
}
