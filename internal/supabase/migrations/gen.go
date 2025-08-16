/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

//go:generate bash -c "cp ../../../supabase/migrations/*.sql ./embedded"
package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"sort"
)

//go:embed "embedded/*.sql"
var sqlFiles embed.FS

type Migration struct {
	Name   string
	Data   []byte
	Number int
}

type MigrationHandler = func(Migration) error

func Get() ([]Migration, error) {

	migrations := make([]Migration, 0)

	entries, err := sqlFiles.ReadDir("embedded")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read embedded migrations: %v", err))
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
			return nil, errors.New(fmt.Sprintf("failed to read %s: %v", e.Name(), err))
		}
		migrations = append(migrations, Migration{
			Name:   e.Name(),
			Data:   data,
			Number: index + 1,
		})
	}

	return migrations, nil
}

func ProcessMigrations(handler MigrationHandler) error {
	entries, err := sqlFiles.ReadDir("embedded")
	if err != nil {
		log.Fatalf("failed to read embedded migrations: %v", err)
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
			return errors.New(fmt.Sprintf("failed to read %s: %v", e.Name(), err))
		}
		if err := handler(Migration{
			Name:   e.Name(),
			Data:   data,
			Number: index + 1,
		}); err != nil {
			return errors.New(fmt.Sprintf("failed to handle %s: %v", e.Name(), err))
		}
	}
	return nil
}
