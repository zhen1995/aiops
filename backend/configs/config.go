package configs

import (
	"fmt"
	"os"
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
	// FrontendBaseURL 前端访问地址：仅作为首启种子写入 system_configs（frontend_base_url），
	// 之后以「系统配置」页为准（巡检报告链接、通知模板 $.domain 变量即时生效）
	FrontendBaseURL string `mapstructure:"frontend_base_url" yaml:"frontend_base_url"`
}

// LoadConfig 使用 Viper 加载配置：依次尝试候选配置文件（第一个存在的生效），
// 全部不存在时退化为纯环境变量 + 默认值运行（如容器内仅注入环境变量的场景）。
func LoadConfig(paths ...string) (*Config, error) {
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
	_ = v.BindEnv("knowledge.python_base_url", "AIOPS_KNOWLEDGE_PYTHON_BASE_URL", "KNOWLEDGE_PYTHON_BASE_URL")
	_ = v.BindEnv("knowledge.upload_dir", "AIOPS_KNOWLEDGE_UPLOAD_DIR", "KNOWLEDGE_UPLOAD_DIR")
	_ = v.BindEnv("knowledge.qdrant_url", "AIOPS_KNOWLEDGE_QDRANT_URL", "KNOWLEDGE_QDRANT_URL")
	v.SetDefault("server.port", ":8080")
	v.SetDefault("app.frontend_base_url", "http://localhost:5173")
	v.SetDefault("log_analysis.interval", "5m")
	v.SetDefault("knowledge.python_base_url", "http://localhost:9000")
	v.SetDefault("knowledge.upload_dir", "./uploads/knowledge")
	v.SetDefault("knowledge.qdrant_url", "http://localhost:6333")

	// 依次尝试候选配置文件，第一个存在的生效；全部不存在时纯依赖环境变量与默认值
	var loaded string
	for _, path := range paths {
		if path == "" {
			continue
		}
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err == nil {
			loaded = path
			break
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
		}
	}
	if loaded == "" {
		fmt.Println("未找到配置文件，使用环境变量与默认值（候选路径: " + strings.Join(paths, ", ") + "）")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}
