/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

//go:generate bash -c "cp ../../../supabase/migrations/*.sql ./embedded"
package migrations

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/train360-corp/projconf/internal/utils"
	"sort"
	"strings"
)

//go:embed "embedded/*.sql"
var sqlFiles embed.FS

type Migration struct {
	Name   string
	Data   []byte
	Number int
}

type SchemaMigrationRow struct {
	Version    string   `db:"version" json:"version"`       // primary key
	Statements []string `db:"statements" json:"statements"` // list of SQL statements
	Name       *string  `db:"name" json:"name,omitempty"`   // optional human-readable name
}

func get() *[]*Migration {

	migrations := make([]*Migration, 0)

	entries, err := sqlFiles.ReadDir("embedded")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded migrations: %v", err))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for index, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := sqlFiles.ReadFile("embedded/" + e.Name())
		if err != nil {
			panic(fmt.Sprintf("failed to read %s: %v", e.Name(), err))
		}
		migrations = append(migrations, &Migration{
			Name:   e.Name(),
			Data:   data,
			Number: index + 1,
		})
	}

	return &migrations
}

func ApplyMigrations(
	ctx context.Context,
	docker *client.Client,
	containerID string,
	existing []SchemaMigrationRow,
) (err error, applied int, passed int) {
	commands := make([]string, 0)
	applied = 0
	passed = 0
	for _, migration := range *get() {

		parts := strings.Split(migration.Name, "_")
		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		alreadyApplied := false
		for _, existingMigration := range existing {
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

	sql := "BEGIN;\n"
	for _, stmt := range commands {
		sql += stmt + ";\n"
	}
	sql += "COMMIT;"

	if _, e := utils.ExecInContainer(ctx, docker, containerID, []string{
		"psql",
		"-h", "127.0.0.1",
		"-U", "supabase_admin",
		"-d", "postgres",
		"-v", "ON_ERROR_STOP=1",
		"-c", sql,
	}); e != nil {
		err = fmt.Errorf("failed to apply migrations: %v", e)
	}
	return
}

func LoadExistingSchemaMigrations(
	ctx context.Context,
	docker *client.Client,
	containerID string,
) ([]SchemaMigrationRow, error) {

	// Emit EXACTLY a JSON array via SQL so we can unmarshal directly.
	sql := `
WITH rows AS (
  SELECT version, statements, name
  FROM supabase_migrations.schema_migrations
  ORDER BY version
)
SELECT COALESCE(json_agg(rows), '[]'::json)
FROM rows;
`

	// -X: no .psqlrc, -q: quiet, -t -A: tuples only / unaligned
	cmd := []string{
		"psql",
		"-h", "127.0.0.1",
		"-U", "supabase_admin",
		"-d", "postgres",
		"-v", "ON_ERROR_STOP=1",
		"-X",
		"-q",
		"-t",
		"-A",
		"-c", sql,
	}

	out, err := utils.ExecInContainer(ctx, docker, containerID, cmd)
	if err != nil {
		if strings.HasPrefix(out, `ERROR:  relation "supabase_migrations.schema_migrations" does not exist`) {
			return []SchemaMigrationRow{}, nil
		} else {
			return nil, fmt.Errorf("psql exec failed: %w (output=%q)", err, out)
		}
	}

	payload := strings.TrimSpace(out)
	if payload == "" {
		return nil, fmt.Errorf("empty psql output")
	}

	var rows []SchemaMigrationRow
	if err := json.Unmarshal([]byte(payload), &rows); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (payload=%s)", err, payload)
	}
	return rows, nil
}
