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
	App      AppConfig      `mapstructure:"app" yaml:"app"`
	// Knowledge 知识库配置
	Knowledge struct {
		PythonBaseURL string `mapstructure:"python_base_url" yaml:"python_base_url"`
		UploadDir     string `mapstructure:"upload_dir" yaml:"upload_dir"`
	} `mapstructure:"knowledge" yaml:"knowledge"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn" yaml:"dsn"`
}

type ServerConfig struct {
	Port string `mapstructure:"port" yaml:"port"`
}

type AppConfig struct {
	// FrontendBaseURL 前端访问地址，用于巡检报告通知中的"完整报告"链接
	FrontendBaseURL string `mapstructure:"frontend_base_url" yaml:"frontend_base_url"`
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
	_ = v.BindEnv("app.frontend_base_url", "AIOPS_APP_FRONTEND_BASE_URL", "FRONTEND_BASE_URL")
	v.SetDefault("app.frontend_base_url", "http://localhost:5173")

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
