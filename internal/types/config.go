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
	Source string    `yaml:"source"`
	LLM    LLMConfig `yaml:"llm"`
}

type LLMConfig struct {
	Provider              string `yaml:"provider"`               // Provider: anthropic, openai, ollama
	Model                 string `yaml:"model"`                  // Model name (e.g., claude-3-5-haiku-latest, gpt-4, llama2)
	APIBase               string `yaml:"api_base,omitempty"`     // Optional: Custom API endpoint (e.g., for Ollama: http://localhost:11434)
	PregenerateThreshold  int    `yaml:"pregenerate_threshold"`
	FallbackToDummy       bool   `yaml:"fallback_to_dummy"`
	TimeoutSeconds        int    `yaml:"timeout_seconds"`
	MaxRetries            int    `yaml:"max_retries"`
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
