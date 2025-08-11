package supabase

import "github.com/urfave/cli/v2"

type Config struct {
	Url     string
	AnonKey string
}

func GetConfigFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "url",
			Usage:       "fqdn http(s) URL where supabase is accessible",
			Value:       "http://127.0.0.1:54321",
			Destination: &cfg.Url,
			EnvVars:     []string{"PROJCONF_SUPABASE_URL"},
		},
		&cli.StringFlag{
			Name:        "anon-key",
			Usage:       "anonymous key for supabase access",
			Value:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0",
			Destination: &cfg.AnonKey,
			EnvVars:     []string{"PROJCONF_SUPABASE_ANON_KEY"},
		},
	}
}
