package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Input      []string          `yaml:"input"`
	IgnoreKeys []string          `yaml:"ignore_keys"`
	Genders    map[string]string `yaml:"genders"`
}

func ReadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := &Config{
		Genders: make(map[string]string),
	}

	err = yaml.NewDecoder(f).Decode(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Config) Gender(in string) string {
	out, ok := c.Genders[in]
	if ok {
		return out
	}
	return ""
}

func (c *Config) SkipKey(key string) bool {
	for _, skip := range c.IgnoreKeys {
		if skip == key {
			return true
		}
	}

	return false
}
