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
		// QdrantURL 知识库 Qdrant 向量库地址，也是 system_configs 中 qdrant_url 的种子默认值
		QdrantURL string `mapstructure:"qdrant_url" yaml:"qdrant_url"`
	} `mapstructure:"knowledge" yaml:"knowledge"`
	// LogAnalysis 日志分析配置
	LogAnalysis struct {
		// Interval 周期分析任务间隔（如 5m / 30s / 1h）
		Interval string `mapstructure:"interval" yaml:"interval"`
	} `mapstructure:"log_analysis" yaml:"log_analysis"`
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
	_ = v.BindEnv("log_analysis.interval", "AIOPS_LOG_ANALYSIS_INTERVAL", "LOG_ANALYSIS_INTERVAL")
	_ = v.BindEnv("knowledge.qdrant_url", "AIOPS_KNOWLEDGE_QDRANT_URL", "KNOWLEDGE_QDRANT_URL")
	v.SetDefault("app.frontend_base_url", "http://localhost:5173")
	v.SetDefault("log_analysis.interval", "5m")
	v.SetDefault("knowledge.qdrant_url", "http://localhost:6333")

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
