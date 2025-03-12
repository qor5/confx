package main

import (
	"log"

	"github.com/spf13/cobra"
)

const (
	flagConfig = "config"
	envPrefix  = "DEMO_"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "a server command",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
