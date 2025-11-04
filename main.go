package main

import (
	"go-touch/internal/sources"
	"go-touch/internal/types"
	"go-touch/internal/ui"
	"log"
)

var ConfigDir string = "config.yaml"

func getText(config types.Config) (string, sources.TextSource, error) {
	textSource, err := sources.NewTextSource(config.Text.Source, config.Text)
	if err != nil {
		return "", nil, err
	}
	text, err := textSource.GetText()
	return text, textSource, err
}

func main() {
	var config, err = types.LoadConfig(ConfigDir)
	if err != nil {
		log.Fatal(err)
	}
	text, textSource, err := getText(*config)
	if err != nil {
		log.Fatal(err)
	}

	// Debug: Check if we got text
	if text == "" {
		log.Fatal("Error: No text generated")
	}

	sessionResult := ui.Run(*config, text, textSource)
	if sessionResult.Error != nil {
		log.Fatal(sessionResult.Error)
	}
	// fmt.Println(sessionResult)
}
