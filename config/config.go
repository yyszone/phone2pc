package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config 是保存在 config.json 中的运行时配置
type Config struct {
	Port      string `json:"port"`
	UploadDir string `json:"uploadDir"`
}

const configFile = "config.json"

var (
	mu  sync.RWMutex
	cur *Config
)

// Load 读取 config.json；不存在则生成默认配置
func Load() error {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			cur = defaults()
			return saveLocked()
		}
		return err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("解析 %s 失败: %w", configFile, err)
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	if c.UploadDir == "" {
		c.UploadDir = "./uploadFile"
	}
	cur = &c
	return nil
}

// Get 返回当前配置的快照（线程安全）
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return *cur
}

// SetUploadDir 更新上传路径并落盘
func SetUploadDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("路径不能为空")
	}
	// 尝试创建目录做一次校验
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("无法创建目录: %w", err)
	}
	mu.Lock()
	defer mu.Unlock()
	cur.UploadDir = dir
	return saveLocked()
}

// SetPort 更新端口并落盘（重启后生效）
func SetPort(port string) error {
	if port == "" {
		return fmt.Errorf("端口不能为空")
	}
	mu.Lock()
	defer mu.Unlock()
	cur.Port = port
	return saveLocked()
}

func defaults() *Config {
	return &Config{Port: "8080", UploadDir: "./uploadFile"}
}

func saveLocked() error {
	data, err := json.MarshalIndent(cur, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0o644)
}

// AbsUploadDir 返回上传目录的绝对路径，便于展示
func AbsUploadDir() string {
	c := Get()
	abs, err := filepath.Abs(c.UploadDir)
	if err != nil {
		return c.UploadDir
	}
	return abs
}