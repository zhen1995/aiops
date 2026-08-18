package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Database DatabaseConfig `mapstructure:"database" yaml:"database"`
	Server   ServerConfig   `mapstructure:"server" yaml:"server"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn" yaml:"dsn"`
}

type ServerConfig struct {
	Port string `mapstructure:"port" yaml:"port"`
}

// LoadConfig 使用 Viper 从配置文件、环境变量加载配置
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// 启用环境变量，前缀为 AIOPS，嵌套键通过下划线分隔
	v.SetEnvPrefix("AIOPS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 显式绑定关键环境变量，兼容无前缀形式
	_ = v.BindEnv("database.dsn", "AIOPS_DATABASE_DSN", "DATABASE_DSN")
	_ = v.BindEnv("server.port", "AIOPS_SERVER_PORT", "SERVER_PORT")

	// 指定配置文件路径
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}
