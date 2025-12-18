package main

import (
	"context"
	"log"

	"github.com/qor5/confx"
	"github.com/qor5/confx/examples/config"
)

func main() {
	loader, err := config.Initialize(confx.WithEnvPrefix("APP_"))
	if err != nil {
		log.Fatalf("Failed to initialize config loader: %v", err)
	}
	conf, err := loader(context.Background(), "")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	conf.Print()
}
