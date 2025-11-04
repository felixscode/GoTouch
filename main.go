package main

import (
	"flag"
	"fmt"
	"go-touch/internal/config"
	"go-touch/internal/sources"
	"go-touch/internal/types"
	"go-touch/internal/ui"
	"log"
)

func getText(cfg types.Config) (string, sources.TextSource, error) {
	textSource, err := sources.NewTextSource(cfg.Text.Source, cfg.Text)
	if err != nil {
		return "", nil, err
	}
	text, err := textSource.GetText()
	return text, textSource, err
}

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Load or create config
	cfg, cfgPath, err := config.LoadOrCreateConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfgPath != "" {
		fmt.Printf("Using config: %s\n", cfgPath)
	}

	// Ensure data directory exists for stats
	dataDir, err := config.GetDataDir()
	if err == nil {
		config.EnsureDir(dataDir)
	}

	text, textSource, err := getText(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Debug: Check if we got text
	if text == "" {
		log.Fatal("Error: No text generated")
	}

	sessionResult := ui.Run(*cfg, text, textSource)
	if sessionResult.Error != nil {
		log.Fatal(sessionResult.Error)
	}
	fmt.Println(sessionResult)
}
