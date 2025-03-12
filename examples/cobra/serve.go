package main

import (
	"log"

	"github.com/qor5/confx"
	"github.com/qor5/confx/examples/config"
	"github.com/spf13/cobra"
)

var serveConfLoader confx.Loader[*config.Config]

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := cmd.Flags().GetString(flagConfig)
		if err != nil {
			log.Fatalf("Failed to get config path: %v", err)
		}

		conf, err := serveConfLoader(cmd.Context(), confPath)
		if err != nil {
			log.Fatalf("Failed to load config: %+v", err)
		}

		conf.Print()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().SortFlags = false
	serveCmd.Flags().String(flagConfig, "", "Path to the configuration yaml file")

	loader, err := config.Initialize(serveCmd.Flags(), envPrefix)
	if err != nil {
		log.Fatalf("Failed to initialize config: %+v", err)
	}
	serveConfLoader = loader
}
