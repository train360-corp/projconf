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
	"log"
	"reflect"
	"strings"
)

// Insert inserts struct `row` into `table` using *pgx.Conn.
// Rules:
//   - Column names come from `db:"col"`; fallback to `json:"name"`, else field name.
//   - Pointer fields that are nil are OMITTED (allowing DB DEFAULTs to fire).
//   - `db:"-"` skips a field; `db:"col,omitempty"` omits zero values.
//   - Pass a RETURNING clause (e.g., "id") and matching scan targets via dest.
func Insert[T any](ctx context.Context, table string, row T, returning string, dest ...any) error {

	conn, err := GetConnection(ctx)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	cols, args, err := structToInsertColsArgs(row)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return fmt.Errorf("no columns to insert for table %s", table)
	}
	ph := make([]string, len(cols))
	for i := range cols {
		ph[i] = fmt.Sprintf("$%d", i+1)
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(cols, ", "))
	sb.WriteString(") VALUES (")
	sb.WriteString(strings.Join(ph, ", "))
	sb.WriteString(")")

	if returning != "" {
		sb.WriteString(" RETURNING ")
		sb.WriteString(returning)
		return conn.QueryRow(ctx, sb.String(), args...).Scan(dest...)
	}
	// No RETURNING: use Exec instead of a dummy Scan.
	_, err = conn.Exec(ctx, sb.String(), args...)
	return err
}

func structToInsertColsArgs[T any](row T) ([]string, []any, error) {
	v := reflect.ValueOf(row)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("row must be struct or *struct, got %s", v.Kind())
	}
	t := v.Type()

	var cols []string
	var args []any

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() || sf.Anonymous {
			continue
		}

		col, omitempty, skip := parseDBTag(sf)
		if skip {
			continue
		}
		if col == "" {
			if jt, ok := sf.Tag.Lookup("json"); ok {
				name := strings.Split(jt, ",")[0]
				if name != "" && name != "-" {
					col = name
				}
			}
			if col == "" {
				col = sf.Name // keep as-is; add snake_case if you want
			}
		}

		fv := v.Field(i)

		// If pointer is nil -> omit column (let DEFAULT fire)
		if fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				continue
			}
			// Keep the original pointer in args to preserve pgx behavior for []byte/pgtype, etc.
		} else if omitempty && fv.IsZero() {
			continue
		}

		cols = append(cols, col)
		args = append(args, v.Field(i).Interface())
	}
	return cols, args, nil
}

func parseDBTag(sf reflect.StructField) (col string, omitempty bool, skip bool) {
	tag, ok := sf.Tag.Lookup("db")
	if !ok {
		return "", false, false
	}
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	col = parts[0]
	for _, p := range parts[1:] {
		if p == "omitempty" {
			omitempty = true
		}
	}
	return col, omitempty, false
}
