package config

import (
	"os"
	"path/filepath"
	"testing"
)

const ExampleConfig = "example-config.yaml"

func TestParse(t *testing.T) {
	cwd, _ := os.Getwd()
	exampleConfigContents, err := os.ReadFile(filepath.Join(cwd, "..", "..", ExampleConfig))
	if err != nil {
		t.Fatal(err)
	}
	config, err := Parse(exampleConfigContents)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Buckets) == 0 {
		t.Fatal("Got empty config")
	}
}
