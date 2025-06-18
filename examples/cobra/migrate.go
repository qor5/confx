package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var dsn string
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the database",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	flagSet.StringVar(&dsn, "database-dsn", "", "database dsn")
	return cmd
}()

func init() {
	rootCmd.AddCommand(migrateCmd)
}
