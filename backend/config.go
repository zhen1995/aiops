package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构体
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

// LoadConfig 从 YAML 文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
