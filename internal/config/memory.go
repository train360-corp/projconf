package config

import (
	"os"
	"sync"

	"github.com/train360-corp/projconf/internal/utils"
)

type Config struct {
	AdminAccessKey string
}

var (
	global Config
	once   sync.Once
)

// GetGlobal returns the process-global config, initializing it once.
func GetGlobal() Config {
	once.Do(func() {
		global = Config{
			AdminAccessKey: utils.Coalesce(
				os.Getenv("PROJCONF_ADMIN_ACCESS_KEY"),
				utils.RandomString(32),
			),
		}
	})
	return global // returns a copy (value type) — safe from accidental mutation
}
