package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"time"
)

// TODO pipe this through
const configFile = "/Users/pnovotnak/src/ohm/example-config.yaml"

type Account struct {
	Key     string `yaml:"key"`
	Profile string `yaml:"profile"`
}

// BlockBucket TODO this just uses NextDNS built in tracker lists I think
type BlockBucket struct {
	// Allowance determines the time from first occurrence in logs to block
	Allowance *time.Duration `yaml:"allowance"`
	Cooldown  *time.Duration `yaml:"cooldown"`
	// Lockout determines the amount of time after allowance is exhausted to block for
	Lockout *time.Duration `yaml:"lockout"`

	// Regex is the internally rendered meaning of Name
	Regex *regexp.Regexp
	// FirstSessionLoad
	FirstSessionLoad *time.Time
	// LastSessionLoad is incremented every time NextDNS returns an answer that matches this bucket.
	// it is not incremented when queries are blocked.
	LastSessionLoad *time.Time
}

type Config struct {
	Account *Account `yaml:"account"`
	// Buckets is a map of FQDN fragments from the denylist to Ohm configurations
	Buckets map[string]*BlockBucket `yaml:"buckets"`
}

func Load() (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return config, err
	}
	for fqdnFragment, bucket := range config.Buckets {
		bucket.Regex, err = regexp.Compile(fmt.Sprintf(".*%s", fqdnFragment))
	}
	return config, err
}
