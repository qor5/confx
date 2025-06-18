package main

import (
	"log"

	"github.com/qor5/confx"
	"github.com/qor5/confx/examples/config"
	"github.com/spf13/cobra"
)

var serveCmd = func() *cobra.Command {
	var confPath string
	var confLoader confx.Loader[*config.Config]

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := confLoader(cmd.Context(), confPath)
			if err != nil {
				return err
			}
			conf.Print()
			return nil
		},
	}

	flagSet := cmd.Flags()
	flagSet.SortFlags = false
	flagSet.StringVar(&confPath, flagConfig, "", "Path to the configuration yaml file")

	var err error
	confLoader, err = config.Initialize(flagSet, envPrefix)
	if err != nil {
		log.Fatalf("Failed to initialize config: %+v", err)
	}

	return cmd
}()

func init() {
	rootCmd.AddCommand(serveCmd)
}
