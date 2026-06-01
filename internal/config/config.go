package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Embed    EmbedConfig    `yaml:"embed"`
	Logging  LoggingConfig  `yaml:"logging"`
	Projects ProjectsConfig `yaml:"projects"`
}

type ServerConfig struct {
	MCP MCPConfig `yaml:"mcp"`
}

type MCPConfig struct {
	Enabled        bool  `yaml:"enabled"`
	ToolTimeout    int   `yaml:"tool_timeout"`
	MaxRequestSize int64 `yaml:"max_request_size"`
}

type DatabaseConfig struct {
	Path            string `yaml:"path"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type EmbedConfig struct {
	ONNXPath  string `yaml:"onnx_path"`
	ONNXModel string `yaml:"onnx_model"`
	MaxSeqLen int    `yaml:"max_seq_len"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

type ProjectsConfig struct {
	AutoCreate        bool     `yaml:"auto_create"`
	DefaultCategories []string `yaml:"default_categories"`
	PathCache         bool     `yaml:"path_cache"`
	CacheTTL          int      `yaml:"cache_ttl"`
}

func Load(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	LoadFromEnv(cfg)

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func LoadFromEnv(cfg *Config) {
	if v := os.Getenv("MEMOOS_SERVER_MCP_ENABLED"); v != "" {
		cfg.Server.MCP.Enabled = parseBool(v, cfg.Server.MCP.Enabled)
	}
	if v := os.Getenv("MEMOOS_SERVER_MCP_TOOL_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Server.MCP.ToolTimeout = i
		}
	}
	if v := os.Getenv("MEMOOS_SERVER_MCP_MAX_REQUEST_SIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Server.MCP.MaxRequestSize = i
		}
	}

	if v := os.Getenv("MEMOOS_DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("MEMOOS_DATABASE_MAX_OPEN_CONNS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxOpenConns = i
		}
	}
	if v := os.Getenv("MEMOOS_DATABASE_MAX_IDLE_CONNS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxIdleConns = i
		}
	}
	if v := os.Getenv("MEMOOS_DATABASE_CONN_MAX_LIFETIME"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Database.ConnMaxLifetime = i
		}
	}

	if v := os.Getenv("MEMOOS_EMBED_ONNX_PATH"); v != "" {
		cfg.Embed.ONNXPath = v
	}
	if v := os.Getenv("MEMOOS_EMBED_ONNX_MODEL"); v != "" {
		cfg.Embed.ONNXModel = v
	}

	if v := os.Getenv("MEMOOS_LOGGING_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("MEMOOS_LOGGING_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
	if v := os.Getenv("MEMOOS_LOGGING_OUTPUT"); v != "" {
		cfg.Logging.Output = v
	}
	if v := os.Getenv("MEMOOS_LOGGING_FILE_PATH"); v != "" {
		cfg.Logging.FilePath = v
	}

	if v := os.Getenv("MEMOOS_PROJECTS_AUTO_CREATE"); v != "" {
		cfg.Projects.AutoCreate = parseBool(v, cfg.Projects.AutoCreate)
	}
	if v := os.Getenv("MEMOOS_PROJECTS_PATH_CACHE"); v != "" {
		cfg.Projects.PathCache = parseBool(v, cfg.Projects.PathCache)
	}
	if v := os.Getenv("MEMOOS_PROJECTS_CACHE_TTL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Projects.CacheTTL = i
		}
	}
	if v := os.Getenv("MEMOOS_PROJECTS_DEFAULT_CATEGORIES"); v != "" {
		cfg.Projects.DefaultCategories = strings.Split(v, ",")
	}
}

func parseBool(s string, defaultValue bool) bool {
	switch strings.ToLower(s) {
	case "true", "t", "yes", "y", "1":
		return true
	case "false", "f", "no", "n", "0":
		return false
	default:
		return defaultValue
	}
}

func Validate(cfg *Config) error {

	if cfg.Server.MCP.ToolTimeout <= 0 || cfg.Server.MCP.ToolTimeout > 300 {
		return fmt.Errorf("tool_timeout must be 1-300 seconds")
	}
	if cfg.Server.MCP.MaxRequestSize <= 0 || cfg.Server.MCP.MaxRequestSize > 100*1024*1024 {
		return fmt.Errorf("max_request_size must be 1-100MB")
	}

	if cfg.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive")
	}
	if cfg.Database.MaxIdleConns < 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return fmt.Errorf("max_idle_conns must be 0-max_open_conns")
	}
	if cfg.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("conn_max_lifetime must be positive")
	}

	if cfg.Embed.ONNXModel == "" {
		return fmt.Errorf("onnx_model is required")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("level must be debug, info, warn, or error")
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("format must be json or text")
	}

	if cfg.Logging.Output != "stdout" && cfg.Logging.Output != "stderr" && cfg.Logging.Output != "file" {
		return fmt.Errorf("output must be stdout, stderr, or file")
	}
	if cfg.Logging.Output == "file" && cfg.Logging.FilePath == "" {
		return fmt.Errorf("file_path required when output is 'file'")
	}

	if cfg.Projects.CacheTTL <= 0 {
		return fmt.Errorf("cache_ttl must be positive")
	}
	for _, cat := range cfg.Projects.DefaultCategories {
		if cat == "" {
			return fmt.Errorf("default_categories cannot contain empty strings")
		}
	}

	return nil
}

func DataDir() string {
	os.MkdirAll("./data", 0755)
	return "./data"
}

func (c *DatabaseConfig) GetDatabasePath() string {
	if c.Path != "" {
		return c.Path
	}
	return DataDir() + "/memoos.db"
}
