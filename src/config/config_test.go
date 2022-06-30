package config

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	config, err := ExampleConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Buckets) == 0 {
		t.Fatal("Got empty config")
	}
}

func TestConfig_Validate(t *testing.T) {
	config, err := ExampleConfig()
	if err != nil {
		t.Fatal(err)
	}
	err = config.Validate()
	if err == nil {
		t.Fatal(errors.New("should have failed validation for missing NextDNS config"))
	}
	exampleKey := "examplekey"
	exampleProfile := "exampleprofile"
	_ = os.Setenv(strings.Join([]string{EnvPrefix, "NEXTDNS", EnvKey}, "_"), exampleKey)
	_ = os.Setenv(strings.Join([]string{EnvPrefix, "NEXTDNS", EnvProfile}, "_"), exampleProfile)
	if config, err = ExampleConfig(); err != nil {
		t.Fatal(err)
	}
	if err = config.Validate(); err != nil {
		t.Fatal(err)
	}
}
