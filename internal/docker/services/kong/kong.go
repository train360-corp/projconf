package kong

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/fs"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed embeds/kong.yml
var KongConfig []byte

type Service struct{}

const ContainerName = "projconf-internal-supabase-kong"

func (s Service) GetDisplay() string {
	return "Kong"
}

func (s Service) Run(evn *types.SharedEvn) error {

	root, err := fs.GetTempRoot()
	if err != nil {
		return errors.New(fmt.Sprintf("could not get temp root directory: %v", err))
	}

	cfg := types.Writeable{
		LocalPath: filepath.Join(root, "kong", "config.yml"),
		Data: []byte(os.Expand(string(KongConfig), func(s string) string {
			switch s {
			case "SUPABASE_ANON_KEY":
				return evn.ANON_KEY
			case "SUPABASE_SERVICE_KEY":
				return evn.SERVICE_KEY
			default:
				panic(fmt.Sprintf("kong environment variable %s not handled", s))
			}
		})),
		Perm:          0o600,
		ContainerPath: "/home/kong/kong.yml:ro,z",
	}

	defer os.Remove(cfg.LocalPath)

	if err := fs.WriteDependencies(cfg.LocalPath, cfg.Data, cfg.Perm); err != nil {
		return errors.New(fmt.Sprintf("could not write kong config: %v", err))
	}

	args := []string{
		"run",
		"--rm",
		"--name", ContainerName,
		"--label", "com.docker.compose.project=projconf",
		"--label", "com.docker.compose.service=kong",
		"--label", "com.docker.compose.version=2.0",
		"--network", "projconf-net",
		"--network-alias", "kong",
		"-p", "127.0.0.1:8000:8000",
		"-v", fmt.Sprintf("%s:%s", cfg.LocalPath, cfg.ContainerPath),
		"-e", "KONG_DATABASE=off",
		"-e", "KONG_DECLARATIVE_CONFIG=/home/kong/kong.yml",
		"-e", "KONG_DNS_ORDER=LAST,A,CNAME",
		"-e", "KONG_PLUGINS=request-transformer,cors,key-auth,acl,basic-auth",
		"-e", "KONG_NGINX_PROXY_PROXY_BUFFER_SIZE=160k",
		"-e", "KONG_NGINX_PROXY_PROXY_BUFFERS=64 160k",
		"-e", "DASHBOARD_USERNAME=not-used",
		"-e", "DASHBOARD_PASSWORD=not-used",
		"-e", fmt.Sprintf("SUPABASE_ANON_KEY=%s", evn.ANON_KEY),
		"-e", fmt.Sprintf("SUPABASE_SERVICE_KEY=%s", evn.SERVICE_KEY),
		"kong:2.8.1",
		"/docker-entrypoint.sh",
		"kong",
		"docker-start",
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf("failed to run docker command: %s", err))
	}

	return nil
}

func (s Service) GetWriteables() []types.Writeable {
	return []types.Writeable{}
}

func (s Service) WaitFor(ctx context.Context) error {
	return nil
}
