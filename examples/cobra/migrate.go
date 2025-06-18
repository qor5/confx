package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			dsn, err := cmd.Flags().GetString("database-dsn")
			if err != nil {
				return errors.Wrap(err, "failed to get database dsn")
			}
			if dsn == "" {
				dsn = os.Getenv(envPrefix + "DATABASE_DSN")
			}
			if dsn == "" {
				return errors.New("database dsn is required")
			}
			log.Printf("should migrate database with dsn: %s", dsn)
			return nil
		},
	}
	flagSet := cmd.Flags()
	flagSet.SortFlags = false
	flagSet.String("database-dsn", "", "database dsn")
	return cmd
}()

func init() {
	rootCmd.AddCommand(migrateCmd)
}
