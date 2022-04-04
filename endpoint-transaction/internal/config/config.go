package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	Redis    Redis    `yaml:"redis"`
	JWT      JWT      `yaml:"jwt"`
}

func New() (*Config, error) {
	filename := fmt.Sprintf("/app/conf/app.%s.yaml", os.Getenv("BACKEND_STAGE"))
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if loadGsmErr := cfg.loadFromGsm(); loadGsmErr != nil {
		return nil, loadGsmErr
	}

	if parseDurationErr := cfg.parseDuration(); parseDurationErr != nil {
		return nil, parseDurationErr
	}

	if parseFileErr := cfg.parseFile(); parseFileErr != nil {
		return nil, parseFileErr
	}

	return &cfg, nil
}

func (cfg *Config) parseDuration() error {
	accessExpired, err := time.ParseDuration(cfg.JWT.AccessExpired)
	if err != nil {
		return err
	}

	cfg.JWT.AccessExpires = accessExpired

	refreshExpired, err := time.ParseDuration(cfg.JWT.RefreshExpired)
	if err != nil {
		return err
	}

	cfg.JWT.RefreshExpires = refreshExpired

	return nil
}

func (cfg *Config) parseFile() error {
	publicKey, err := os.ReadFile(cfg.JWT.PublicKey)
	if err != nil {
		return err
	}

	cfg.JWT.PublicKey = string(publicKey)

	privateKey, err := os.ReadFile(cfg.JWT.PrivateKey)
	if err != nil {
		return err
	}

	cfg.JWT.PrivateKey = string(privateKey)

	return nil
}
