package config

import (
	"os"
	"path/filepath"
)

const ExampleConfigFile = "example-config.yaml"

func ExampleConfig() (*Config, error) {
	cwd, _ := os.Getwd()
	exampleConfigContents, err := os.ReadFile(filepath.Join(cwd, "..", "..", ExampleConfigFile))
	if err != nil {
		return nil, err
	}
	config, err := Parse(exampleConfigContents)
	if err != nil {
		return nil, err
	}
	return config, nil
}
