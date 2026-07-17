package main

import (
	"context"
	"log"
	"os"

	"github.com/qor5/confx"
	"github.com/qor5/confx/examples/config"
	"github.com/spf13/pflag"
)

func main() {
	var confPath string

	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	flagSet.SortFlags = false
	flagSet.StringVarP(&confPath, "config", "c", "", "Path to the configuration yaml file")

	loader, err := config.Initialize(confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
	if err != nil {
		log.Fatalf("Failed to initialize config loader: %v", err)
	}

	// When using custom flagSet with user-defined flags (like --config),
	// you must parse flags before calling the loader to populate those variables.
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	conf, err := loader(context.Background(), confPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	conf.Print()
}
