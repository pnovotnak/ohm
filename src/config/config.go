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

// NextDNS should be used for debugging only. It's better not to store this data on disk
type NextDNS struct {
	Key     string `mapstructure:"key"`
	Profile string `mapstructure:"profile"`
}

// BlockBucket TODO this just uses NextDNS built in tracker lists I think
type BlockBucket struct {
	// Allowance determines the time from first occurrence in logs to block
	Allowance time.Duration `yaml:"allowance"`
	Cooldown  time.Duration `yaml:"cooldown"`
	// Lockout determines the amount of time after allowance is exhausted to block for
	Lockout time.Duration `yaml:"lockout"`

	// Regex is the internally rendered meaning of Name
	Regex *regexp.Regexp
	// FirstSessionLoad
	FirstSessionLoad *time.Time
	// LastSessionLoad is incremented every time NextDNS returns an answer that matches this bucket.
	// it is not incremented when queries are blocked.
	LastSessionLoad *time.Time
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
		bucket.Regex, err = regexp.Compile(fmt.Sprintf(".*%s", fqdnFragment))
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
