package types

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Text  TextConfig  `yaml:"text"`
	Ui    UiConfig    `yaml:"ui"`
	Stats StatsConfig `yaml:"stats"`
}

type UiConfig struct {
	Theme string `yaml:"theme"`
}

type TextConfig struct {
	Source string `yaml:"source"`
}
type StatsConfig struct {
	FileDir string `yaml:"file_dir"`
}

func LoadConfig(dir string) (*Config, error) {
	data, err := ioutil.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
