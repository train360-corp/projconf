/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package migrations

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

var MigrationsSchemaStatements []string = []string{
	"CREATE SCHEMA IF NOT EXISTS supabase_migrations",
	"CREATE TABLE IF NOT EXISTS supabase_migrations.schema_migrations ()",
	"ALTER TABLE supabase_migrations.schema_migrations ADD COLUMN IF NOT EXISTS version text NOT NULL PRIMARY KEY",
	"ALTER TABLE supabase_migrations.schema_migrations ADD COLUMN IF NOT EXISTS statements text[]",
	"ALTER TABLE supabase_migrations.schema_migrations ADD COLUMN IF NOT EXISTS name text",
	//"CREATE TABLE IF NOT EXISTS supabase_migrations.seed_files ()",
	//"ALTER TABLE supabase_migrations.seed_files ADD COLUMN IF NOT EXISTS path text NOT NULL PRIMARY KEY",
	//"ALTER TABLE supabase_migrations.seed_files ADD COLUMN IF NOT EXISTS hash text NOT NULL",
}

type SchemaMigration struct {
	Version    string   `db:"version" json:"version"`       // primary key
	Statements []string `db:"statements" json:"statements"` // list of SQL statements
	Name       *string  `db:"name" json:"name,omitempty"`   // optional human-readable name
}

func LoadSchemaMigrations(ctx context.Context, conn *pgx.Conn) ([]SchemaMigration, error) {
	rows, err := conn.Query(ctx, `
		SELECT version, statements, name
		FROM supabase_migrations.schema_migrations
		ORDER BY version
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var migrations []SchemaMigration
	for rows.Next() {
		var m SchemaMigration
		if err := rows.Scan(&m.Version, &m.Statements, &m.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		migrations = append(migrations, m)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return migrations, nil
}
