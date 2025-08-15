package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/train360-corp/projconf/internal/fs"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

type DiskConfigAccountClient struct {
	Id     string `yaml:"id"`
	Secret string `yaml:"secret"`
}

type DiskConfigAccount struct {
	Url    string                  `yaml:"url"`
	Client DiskConfigAccountClient `yaml:"client"`
}

type DiskConfigSupabaseKeys struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type DiskConfigSupabaseDb struct {
	Password string `yaml:"password"`
}

type DiskConfigSupabase struct {
	Url       string                 `yaml:"url"`
	JwtSecret string                 `yaml:"jwt-secret"`
	Keys      DiskConfigSupabaseKeys `yaml:"keys"`
	Db        DiskConfigSupabaseDb   `yaml:"db"`
}

type DiskConfig struct {
	Account  DiskConfigAccount  `yaml:"account"`
	Supabase DiskConfigSupabase `yaml:"supabase"`
}

// randomString returns a secure random string of length n.
func randomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes)
}

func genDefaultConfig() *DiskConfig {

	jwtSecret := randomString(32)
	iat := time.Now().Unix()
	exp := time.Now().AddDate(10, 0, 0).Unix()

	pubKey, pubKeyErr := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(map[string]any{
		"role": "anon",
		"iss":  "supabase",
		"iat":  iat,
		"exp":  exp,
	})).SignedString([]byte(jwtSecret))
	if pubKeyErr != nil {
		panic(errors.New(fmt.Sprintf("failed to generate default config - public key error: %s", pubKeyErr.Error())))
	}

	privKey, privKeyErr := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(map[string]any{
		"role": "service_role",
		"iss":  "supabase",
		"iat":  iat,
		"exp":  exp,
	})).SignedString([]byte(jwtSecret))
	if privKeyErr != nil {
		panic(errors.New(fmt.Sprintf("failed to generate default config - private key error: %s", privKeyErr.Error())))
	}

	return &DiskConfig{
		Account: DiskConfigAccount{Url: "http://127.0.0.1:8080"},
		Supabase: DiskConfigSupabase{
			Url:       "http://127.0.0.1:8000",
			JwtSecret: jwtSecret,
			Keys: DiskConfigSupabaseKeys{
				Public:  pubKey,
				Private: privKey,
			},
			Db: DiskConfigSupabaseDb{
				Password: randomString(32),
			},
		},
	}
}

func getConfigPath() string {
	path, err := fs.GetUserRoot()
	if err != nil {
		panic(errors.New(fmt.Sprintf("failed to get user root path: %s", err)))
	}
	return filepath.Join(path, "config.yaml")
}

// Read loads the local config file (and creates one if one does not exist)
func Read() (*DiskConfig, error) {

	path := getConfigPath()
	if !fs.FileExists(path) {
		err := genDefaultConfig().Flush()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("a config file does not exist and an error occurred while writing the default config: %s", err))
		}
		// even if we write default successfully,
		// load from file to ensure it is readable later
	}

	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read config file: %s", err))
	}

	cfg := &DiskConfig{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to unmarshal config file: %s", err))
	}

	return cfg, nil
}

// Flush save changes in the config object to the disk
func (cfg *DiskConfig) Flush() error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0o600)
}
