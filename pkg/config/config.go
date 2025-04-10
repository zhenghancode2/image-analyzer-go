package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LogConfig 日志配置
type LogConfig struct {
	Dir    string `json:"dir"`
	File   string `json:"file"`
	Level  string `json:"level"`
	Format string `json:"format"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	MaxRequestSize int64  `json:"max_request_size"`
}

// AnalyzeConfig 分析配置
type AnalyzeConfig struct {
	TempDir             string   `json:"temp_dir"`
	CheckOSInfo         bool     `json:"check_os_info"`
	CheckPythonPackages bool     `json:"check_python_packages"`
	CheckCommonTools    bool     `json:"check_common_tools"`
	SpecificCommands    []string `json:"specific_commands"`
}

// Config 全局配置
type Config struct {
	Log     LogConfig     `json:"log"`
	Server  ServerConfig  `json:"server"`
	Analyze AnalyzeConfig `json:"analyze"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Log: LogConfig{
			Dir:    "logs",
			File:   "image-analyzer.log",
			Level:  "info",
			Format: "json",
		},
		Server: ServerConfig{
			Host:           "0.0.0.0",
			Port:           8080,
			ReadTimeout:    30,
			WriteTimeout:   30,
			MaxRequestSize: 10 * 1024 * 1024, // 10MB
		},
		Analyze: AnalyzeConfig{
			TempDir:             "temp",
			CheckOSInfo:         true,
			CheckPythonPackages: true,
			CheckCommonTools:    true,
			SpecificCommands:    []string{},
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 解析 YAML
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetLogPath 获取完整的日志文件路径
func (c *Config) GetLogPath() string {
	return filepath.Join(c.Log.Dir, c.Log.File)
}

// GetTempDir 获取临时目录路径
func (c *Config) GetTempDir() string {
	return c.Analyze.TempDir
}

// EnsureDirs 确保所需的目录存在
func (c *Config) EnsureDirs() error {
	dirs := []string{c.Log.Dir, c.Analyze.TempDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
