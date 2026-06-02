package config

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CA    CAConfig    `yaml:"ca"`
	Certs CertsConfig `yaml:"certs"`
}

type CAConfig struct {
	CommonName   string `yaml:"common_name"`
	Organization string `yaml:"organization"`
	Country      string `yaml:"country"`
	ValidYears   int    `yaml:"valid_years"`
}

type CertsConfig struct {
	ValidDays int `yaml:"valid_days"`
}

func DefaultConfig() *Config {
	return &Config{
		CA: CAConfig{
			CommonName:   "Ore2CA Local Root CA",
			Organization: "Ore2CA",
			Country:      "JP",
			ValidYears:   10,
		},
		Certs: CertsConfig{
			ValidDays: 825,
		},
	}
}

func HomeDir() (string, error) {
	// sudo 実行時は元ユーザーのホームを使う
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if u, err := user.Lookup(sudoUser); err == nil {
			return filepath.Join(u.HomeDir, ".ore2ca"), nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ore2ca"), nil
}

func Load() (*Config, error) {
	dir, err := HomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	cfg := DefaultConfig()
	return cfg, yaml.Unmarshal(data, cfg)
}

func Save(cfg *Config) error {
	dir, err := HomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0600)
}
