package config

import (
	"github.com/spf13/viper"
	"strings"
)

// Здесь будет конфигурация микросервиса

type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	GRPC struct {
		Port int
	}
}

func Init(path string) (*Config, error) {
	if err := parseConfigFile(path); err != nil {
		return nil, err
	}
	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseConfigFile(filepath string) error {
	path := strings.Split(filepath, "/")

	viper.AddConfigPath(path[0]) // folder
	viper.SetConfigName(path[1]) // config file name

	return viper.ReadInConfig()
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("db", &cfg.DB); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("grpc", &cfg.GRPC); err != nil {
		return err
	}
	return nil
}
