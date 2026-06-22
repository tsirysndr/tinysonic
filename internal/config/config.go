package config

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	MusicDir     string `toml:"music_dir"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	Port         int    `toml:"port"`
	Host         string `toml:"host"`
	DatabasePath string `toml:"database_path"`
	CoversDir    string `toml:"covers_dir"`
}

func Load(path string) (*Config, error) {
	c := &Config{
		Port:         4533,
		Host:         "0.0.0.0",
		DatabasePath: "smolsonic.db",
		CoversDir:    "covers",
	}
	if _, err := toml.DecodeFile(path, c); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}
	if c.Username == "" {
		return nil, errors.New("config: username must not be empty")
	}
	if c.Password == "" {
		return nil, errors.New("config: password must not be empty")
	}
	if c.MusicDir == "" {
		return nil, errors.New("config: music_dir must not be empty")
	}
	return c, nil
}
