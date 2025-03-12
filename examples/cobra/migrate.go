package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database",
	Run: func(cmd *cobra.Command, args []string) {
		dsn, err := cmd.Flags().GetString("database-dsn")
		if err != nil {
			log.Fatalf("Failed to get database dsn: %v", err)
		}
		if dsn == "" {
			dsn = os.Getenv(envPrefix + "DATABASE_DSN")
		}
		log.Printf("should migrate database with dsn: %s", dsn)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().String("database-dsn", "", "database dsn")
}
