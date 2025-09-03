/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/train360-corp/projconf/pkg/supabase/migrations"
	"io"
	"net/http"
	"strings"
	"time"
)

// PreviewString returns the first and last five characters of a string, separated by '...'
func PreviewString(s string) string {
	if len(s) <= 10 {
		return s
	}
	return fmt.Sprintf("%s...%s", s[:5], PreviewString(s[len(s)-5:]))
}

func WaitPostgresReady(ctx context.Context, id string) error {
	backoff := time.Second
	for retries := 0; retries < 5; retries++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		out, err := ExecInContainer(ctx, id, []string{"pg_isready"})
		if err == nil && strings.Contains(out, "accepting connections") {
			return nil
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if backoff < 5*time.Second {
			backoff *= 2
		}
	}
	return fmt.Errorf("pg_isready did not succeed")
}

func WaitHTTPReady(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	backoff := time.Second
	for retries := 0; retries < 5; retries++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := client.Get(url)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if backoff < 5*time.Second {
			backoff *= 2
		}
	}
	return fmt.Errorf("http not ready: %s", url)
}

func Migrate(ctx context.Context, pgContainerId string) error {

	MustLogger()

	// apply migrations to build migrations_schema
	sql := "BEGIN;\n"
	for _, stmt := range migrations.MigrationsSchemaStatements {
		sql += stmt + ";\n"
	}
	sql += "COMMIT;"
	output, err := ExecInContainer(ctx, pgContainerId, []string{
		"psql",
		"-h", "127.0.0.1",
		"-U", "supabase_admin",
		"-d", "postgres",
		"-v", "ON_ERROR_STOP=1",
		"-c", sql,
	})
	if err != nil {
		return fmt.Errorf("Migrate migrations meta-schema failed (%v): %s", err, strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n"))
	}
	Logger.Debug(fmt.Sprintf("Migrate migrations meta-schema succeeded: %s", strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n")))

	existingMigrations, err := loadSchemaMigrations(ctx, pgContainerId)
	if err != nil {
		return fmt.Errorf("failed to load existing schema migrations: %s", err.Error())
	}

	m, _ := migrations.Get()
	commands := make([]string, 0)
	applied := 0
	passed := 0
	for _, migration := range m {

		parts := strings.Split(migration.Name, "_")
		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		alreadyApplied := false
		for _, existingMigration := range existingMigrations {
			if existingMigration.Version == version {
				alreadyApplied = true
				passed++
			}
		}

		if !alreadyApplied {
			applied++
			commands = append(commands, string(migration.Data), fmt.Sprintf("INSERT INTO supabase_migrations.schema_migrations (version, name) VALUES ('%s', '%s')", version, name))
		}
	}

	sql = "BEGIN;\n"
	for _, stmt := range commands {
		sql += stmt + ";\n"
	}
	sql += "COMMIT;"
	if output, err := ExecInContainer(ctx, pgContainerId, []string{
		"psql",
		"-h", "127.0.0.1",
		"-U", "supabase_admin",
		"-d", "postgres",
		"-v", "ON_ERROR_STOP=1",
		"-c", sql,
	}); err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	} else {
		Logger.Debug(fmt.Sprintf("apply migrations succeeded: %s", strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n")))
	}
	Logger.Debug(fmt.Sprintf("%v new migrations applied (%v already applied, %d total)", applied, passed, len(m)))

	return nil
}

type SchemaMigration struct {
	Version    string   `json:"version"`        // primary key
	Statements []string `json:"statements"`     // list of SQL statements
	Name       *string  `json:"name,omitempty"` // optional human-readable name
}

func loadSchemaMigrations(ctx context.Context, containerID string) ([]SchemaMigration, error) {
	sql := `
      SELECT coalesce(json_agg(row_to_json(t)), '[]')
      FROM (
        SELECT version, statements, name
        FROM supabase_migrations.schema_migrations
        ORDER BY version
      ) t;
    `

	out, err := ExecInContainer(ctx, containerID, []string{
		"psql",
		"-h", "127.0.0.1",
		"-U", "supabase_admin",
		"-d", "postgres",
		"-t", "-A", "-F", ",", // no headers, unaligned
		"-v", "ON_ERROR_STOP=1",
		"-c", sql,
	})
	if err != nil {
		return nil, fmt.Errorf("psql ExecInContainer: %w (out=%s)", err, strings.TrimSpace(out))
	}

	// out will be something like:
	//  [{"version":"20240901","statements":["..."],"name":"init"}]
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}

	var migrations []SchemaMigration
	if err := json.Unmarshal([]byte(trimmed), &migrations); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (out=%s)", err, trimmed)
	}
	return migrations, nil
}

// RemoveDanglingContainers ensures that any old or dangling projconf containers are stopped and removed.
func RemoveDanglingContainers(ctx context.Context) error {

	MustLogger()
	MustCli()

	// Filter containers whose names contain "projconf"
	f := filters.NewArgs()
	f.Add("label", "com.docker.compose.project=projconf")

	containers, err := Cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	for _, c := range containers {
		name := strings.Join(c.Names, ",")
		Logger.Debug(fmt.Sprintf("cleaning up container %s (%s)", name, c.ID[:12]))

		// Try to stop it first (ignore error if already exited)
		if c.State == "running" {
			if err := Cli.ContainerStop(ctx, c.ID, container.StopOptions{}); err != nil {
				Logger.Warn(fmt.Sprintf("failed to stop container %s: %v", c.ID[:12], err))
			}
		}

		// Remove it (force if necessary)
		if err := Cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
			if strings.Index(err.Error(), "is already in progress") != -1 {
				Logger.Debug("container removal(s) already in progress (sleeping 5 seconds...)")
				time.Sleep(5 * time.Second)
			} else {
				return fmt.Errorf("failed to remove container %s: %v", c.ID[:12], err)
			}
		} else {
			Logger.Debug(fmt.Sprintf("removed stale container %s", c.ID[:12]))
		}
	}

	return nil
}
