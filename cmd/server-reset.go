/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/fs"
	"log"
	"os"
	"path/filepath"
)

var resetCmd = &cobra.Command{
	Use:           "reset",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.NoArgs,
	Short:         "reset a ProjConf server instance",
	Long: `Delete all data associated with a ProjConf server.
This action cannot be undone.`,
	RunE: func(c *cobra.Command, args []string) error {
		root, err := fs.GetUserRoot()
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// remove local db folder
		if err := os.RemoveAll(filepath.Join(root, "db")); err != nil {
			return errors.New(fmt.Sprintf("failed to remove database: %s", err))
		} else {
			log.Printf("removed database directory")
		}

		// reset db config
		cfg.Supabase = config.GenDefaultConfig().Supabase
		if err := cfg.Flush(); err != nil {
			return errors.New(fmt.Sprintf("failed to reset database config: %s", err))
		} else {
			log.Printf("reset database config")
		}

		return nil
	},
}

func init() {
	serverCmd.AddCommand(resetCmd)
	resetCmd.Flags().Bool("confirm", false, "confirm deletion of the server")
	resetCmd.MarkFlagRequired("confirm")
}
