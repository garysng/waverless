package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var GlobalConfig *Config

// Config global configuration
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Redis        RedisConfig        `yaml:"redis"`
	MySQL        MySQLConfig        `yaml:"mysql"`
	Queue        QueueConfig        `yaml:"queue"`
	Worker       WorkerConfig       `yaml:"worker"`
	Logger       LoggerConfig       `yaml:"logger"`
	K8s          K8sConfig          `yaml:"k8s"`
	AutoScaler   AutoScalerConfig   `yaml:"autoscaler"`
	Docker       DockerConfig       `yaml:"docker"`              // Docker registry authentication
	Notification NotificationConfig `yaml:"notification"`        // Notification configuration
	Providers    *ProvidersConfig   `yaml:"providers,omitempty"` // Providers configuration (optional)
	Novita       NovitaConfig       `yaml:"novita"`              // Novita serverless configuration
}

// ServerConfig server configuration
type ServerConfig struct {
	Port    int    `yaml:"port"`
	Mode    string `yaml:"mode"`     // debug, release
	APIKey  string `yaml:"api_key"`  // API key for worker authentication (optional, if empty, auth is disabled)
	BaseURL string `yaml:"base_url"` // Base URL for the server
}

// RedisConfig Redis configuration
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MySQLConfig MySQL configuration
type MySQLConfig struct {
	Host     string       `yaml:"host"`
	Port     int          `yaml:"port"`
	User     string       `yaml:"user"`
	Password string       `yaml:"password"`
	Database string       `yaml:"database"`
	Proxy    *ProxyConfig `yaml:"proxy,omitempty"` // Proxy configuration (optional)
}

// ProxyConfig Proxy configuration for network connections
type ProxyConfig struct {
	Enabled bool   `yaml:"enabled"` // Enable proxy
	Type    string `yaml:"type"`    // Proxy type: http, https, socks5
	Host    string `yaml:"host"`    // Proxy host
	Port    int    `yaml:"port"`    // Proxy port
}

// QueueConfig queue configuration
type QueueConfig struct {
	Concurrency int `yaml:"concurrency"`  // queue processing concurrency
	MaxRetry    int `yaml:"max_retry"`    // maximum retry count
	TaskTimeout int `yaml:"task_timeout"` // task timeout (seconds)
	// Note: Task data is persisted permanently in Redis (no TTL)
}

// WorkerConfig Worker configuration
type WorkerConfig struct {
	HeartbeatInterval  int `yaml:"heartbeat_interval"`  // Heartbeat interval (seconds)
	HeartbeatTimeout   int `yaml:"heartbeat_timeout"`   // Heartbeat timeout (seconds)
	DefaultConcurrency int `yaml:"default_concurrency"` // default concurrency
}

// LoggerConfig logger configuration
type LoggerConfig struct {
	Level  string           `yaml:"level"`  // debug, info, warn, error
	Output string           `yaml:"output"` // console, file, both
	File   LoggerFileConfig `yaml:"file"`
}

// LoggerFileConfig logger file configuration
type LoggerFileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`    // MB per file (default: 100)
	MaxBackups int    `yaml:"max_backups"` // max backup files (default: 3)
	MaxAge     int    `yaml:"max_age"`     // days to keep (default: 7)
	Compress   bool   `yaml:"compress"`    // compress rotated files (default: false)
}

// K8sConfig K8s configuration
type K8sConfig struct {
	Enabled   bool       `yaml:"enabled"`       // whether to enable K8s features
	Namespace string     `yaml:"namespace"`     // K8s namespace
	Platform  string     `yaml:"platform"`      // Platform type: generic, aliyun-ack, aws-eks
	ConfigDir string     `yaml:"config_dir"`    // Configuration directory (specs.yaml and templates)
	AWS       *AWSConfig `yaml:"aws,omitempty"` // AWS configuration (for aws-eks platform)
}

// AWSConfig AWS configuration
type AWSConfig struct {
	Region          string `yaml:"region"`            // AWS region (optional, auto-detect if empty)
	AccessKeyID     string `yaml:"access_key_id"`     // AWS Access Key ID (optional, use IAM role if empty)
	SecretAccessKey string `yaml:"secret_access_key"` // AWS Secret Access Key (optional)
}

// ProvidersConfig providers configuration
type ProvidersConfig struct {
	Deployment string `yaml:"deployment"` // Deployment provider: k8s, docker, custom
	Queue      string `yaml:"queue"`      // Queue provider: redis, mysql, postgres
	Metadata   string `yaml:"metadata"`   // Metadata storage: redis, mysql, postgres
}

// AutoScalerConfig autoscaler configuration
type AutoScalerConfig struct {
	Enabled        bool `yaml:"enabled"`         // Whether to enable autoscaling
	Interval       int  `yaml:"interval"`        // Control loop interval (seconds)
	MaxGPUCount    int  `yaml:"max_gpu_count"`   // Total cluster GPU count
	MaxCPUCores    int  `yaml:"max_cpu_cores"`   // Total cluster CPU cores
	MaxMemoryGB    int  `yaml:"max_memory_gb"`   // Total cluster memory (GB)
	StarvationTime int  `yaml:"starvation_time"` // Starvation time threshold (seconds)
}

// DockerConfig Docker registry authentication configuration
type DockerConfig struct {
	ProxyURL   string                        `yaml:"proxy_url"`  // HTTP proxy URL (e.g., "http://127.0.0.1:7890")
	Registries map[string]DockerRegistryAuth `yaml:"registries"` // Registry authentication (key: registry URL)
}

// DockerRegistryAuth Docker registry authentication info
type DockerRegistryAuth struct {
	Username string `yaml:"username"` // Registry username
	Password string `yaml:"password"` // Registry password or token
	Auth     string `yaml:"auth"`     // Base64 encoded username:password (optional)
}

// NotificationConfig Notification configuration
type NotificationConfig struct {
	FeishuWebhookURL string `yaml:"feishu_webhook_url"` // Feishu (Lark) webhook URL
}

// NovitaConfig Novita serverless configuration
type NovitaConfig struct {
	Enabled      bool   `yaml:"enabled"`       // Whether to enable Novita provider
	APIKey       string `yaml:"api_key"`       // Novita API key (Bearer token)
	BaseURL      string `yaml:"base_url"`      // API base URL, default: https://api.novita.ai
	ConfigDir    string `yaml:"config_dir"`    // Configuration directory (specs.yaml and templates)
	PollInterval int    `yaml:"poll_interval"` // Poll interval for status updates (seconds, default: 10)
}

// Init initializes configuration
func Init() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	GlobalConfig = &cfg
	return nil
}
