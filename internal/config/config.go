package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	DefaultRetainResults = 1000
	DefaultTimeout       = 10 * time.Second
)

type Check struct {
	Name          string        `toml:"name"`
	URL           string        `toml:"url"`
	Interval      time.Duration `toml:"interval"`
	Timeout       time.Duration `toml:"timeout"`
	RetainResults int           `toml:"retain_results"`
}

type Config struct {
	RetainResults int           `toml:"retain_results"`
	Timeout       time.Duration `toml:"timeout"`
	Checks        []Check       `toml:"checks"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if len(cfg.Checks) == 0 {
		return nil, fmt.Errorf("no checks defined in config")
	}

	if cfg.RetainResults <= 0 {
		cfg.RetainResults = DefaultRetainResults
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}

	for i, c := range cfg.Checks {
		if c.Name == "" {
			return nil, fmt.Errorf("check %d: missing name", i+1)
		}
		if c.URL == "" {
			return nil, fmt.Errorf("check %q: missing url", c.Name)
		}
		if c.Interval <= 0 {
			return nil, fmt.Errorf("check %q: interval must be positive", c.Name)
		}
		if cfg.Checks[i].Timeout <= 0 {
			cfg.Checks[i].Timeout = cfg.Timeout
		}
		if cfg.Checks[i].RetainResults <= 0 {
			cfg.Checks[i].RetainResults = cfg.RetainResults
		}
	}

	return &cfg, nil
}
