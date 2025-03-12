// Package main demonstrates the usage of confx library with various features.
package main

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/qor5/confx/examples/config"
	"github.com/spf13/pflag"
)

const (
	EnvPrefix  = "APP_"
	FlagConfig = "config"
)

// setupConfig loads configuration from various sources.
func setupConfig(ctx context.Context) (*config.Config, error) {
	// Create a flag set for command-line arguments
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	flagSet.SortFlags = false

	// Add config file flag
	flagSet.StringP(FlagConfig, "c", "", "Path to configuration file")

	// Initialize the config loader with custom validator
	configLoader, err := config.Initialize(flagSet, EnvPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize config loader")
	}

	// Parse command-line arguments
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == pflag.ErrHelp {
			os.Exit(0)
		}
		return nil, errors.Wrap(err, "failed to parse command line flags")
	}

	// Load configuration
	confPath, err := flagSet.GetString(FlagConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config path")
	}
	conf, err := configLoader(ctx, confPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return conf, nil
}

func main() {
	ctx := context.Background()

	// Load configuration
	conf, err := setupConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	conf.Print()
}
