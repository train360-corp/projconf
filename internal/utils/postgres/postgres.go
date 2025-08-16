/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/train360-corp/projconf/internal/config"
	"log"
)

func GetConnection(ctx context.Context) (*pgx.Conn, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"supabase_admin",
		cfg.Supabase.Db.Password,
		"127.0.0.1",
		5432,
		"postgres",
	))
}

func ExecuteOnEmbeddedDatabase(ctx context.Context, commands []string) error {

	conn, err := GetConnection(ctx)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) // safe even if commit succeeds

	for _, command := range commands {
		if _, err := tx.Exec(ctx, command); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
