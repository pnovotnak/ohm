package config

import (
	_ "embed"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	EnvPrefix  = "OHM"
	EnvKey     = "KEY"
	EnvProfile = "PROFILE"
)

type NextDNS struct {
	Key     string `mapstructure:"key"`
	Profile string `mapstructure:"profile"`
}

type BlockBucket struct {
	// Allowance determines the time from first occurrence in logs to block
	Allowance time.Duration `yaml:"allowance"`
	Cooldown  time.Duration `yaml:"cooldown"`
	// Lockout determines the amount of time after allowance is exhausted to block for
	Lockout time.Duration `yaml:"lockout"`

	// Regex is the internally rendered meaning of Name
	Regex *regexp.Regexp
}

func (bb *BlockBucket) Init(fqdnFragment string) error {
	var err error
	bb.Regex, err = regexp.Compile(fmt.Sprintf(".*%s", fqdnFragment))
	return err
}

type Config struct {
	Test    string  `yaml:"test"`
	NextDNS NextDNS `yaml:"nextdns"`
	// Buckets is a map of FQDN fragments from the denylist to Ohm configurations
	Buckets map[string]*BlockBucket `yaml:"buckets"`
}

func Parse(config []byte) (*Config, error) {
	var err error

	parsed := &Config{}
	err = yaml.Unmarshal(config, parsed)

	parsed.NextDNS.Key = getEnv(parsed.NextDNS.Key, "nextdns", EnvKey)
	parsed.NextDNS.Profile = getEnv(parsed.NextDNS.Profile, "nextdns", EnvProfile)
	if err != nil {
		return parsed, err
	}
	for fqdnFragment, bucket := range parsed.Buckets {
		err = bucket.Init(fqdnFragment)
	}
	return parsed, err
}

func (c *Config) Validate() error {
	if c.NextDNS.Key == "" || c.NextDNS.Profile == "" {
		return errors.New("nextdns key and profile must be provided")
	}
	return nil
}

func getEnv(defaultValue string, keyparts ...string) string {
	key := strings.ToUpper(strings.Join(append([]string{EnvPrefix}, keyparts...), "_"))
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		return defaultValue
	}
}
