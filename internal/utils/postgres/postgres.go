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
)

func Execute(ctx context.Context, conn *pgx.Conn, commands []string) error {
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
