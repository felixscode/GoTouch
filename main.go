package main

import (
	"fmt"
	"go-touch/internal/sources"
	"go-touch/internal/types"
	"go-touch/internal/ui"
	"log"
)

var ConfigDir string = "config.yaml"

func getText(config types.Config) (string, error) {
	textSource, err := sources.NewTextSource(config.Text.Source)
	if err != nil {
		return "", err
	}
	text, err := textSource.GetText()
	return text, err
}

func main() {
	var config, err = types.LoadConfig(ConfigDir)
	if err != nil {
		log.Fatal(err)
	}
	text, err := getText(*config)
	if err != nil {
		log.Fatal(err)
	}
	sessionResult := ui.Run(*config, text)
	if sessionResult.Error != nil {
		log.Fatal(sessionResult.Error)
	}
	fmt.Println(sessionResult)
}
