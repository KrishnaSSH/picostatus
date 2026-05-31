package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Check struct {
	Name     string        `toml:"name"`
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`
}

type Config struct {
	Checks []Check `toml:"checks"`
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
	}

	return &cfg, nil
}
