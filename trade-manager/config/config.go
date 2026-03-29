package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"trade-manager/logging"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	BaseURL    string `json:"base_url"`
	Timeout    int    `json:"timeout"`
	APIToken   string `json:"api_token"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`       // 日志级别: debug, info, warn, error, fatal
	FilePath   string `json:"file_path"`  // 日志文件路径
	MaxSize    int64  `json:"max_size"`    // 最大文件大小 (MB)
	MaxBackups int    `json:"max_backups"` // 最大备份文件数
	MaxAge     int    `json:"max_age"`    // 最大保存天数
	Compress   bool   `json:"compress"`   // 是否压缩备份文件
}

// RuntimeConfig 运行时配置
type RuntimeConfig struct {
	MaxWorkers int    `json:"max_workers"`
	RetryCount int    `json:"retry_count"`
	LogLevel   string `json:"log_level"`
}

// AccountConfig 账户配置
type AccountConfig struct {
	XYAccount    string `json:"XyBusinAccount"`
	StockAccount string `json:"stockAccount"`
	ClientType   string `json:"clientType"`
	Enabled      bool   `json:"enabled"`
}

// SamePositionConfig 相同持仓配置
type SamePositionConfig struct {
	SourceAccount  AccountConfig   `json:"source_account"`
	TargetAccounts []AccountConfig `json:"target_accounts"`
	OrderTemplate  OrderTemplate   `json:"order_template"`
}

// OrderTemplate 订单模板配置
type OrderTemplate struct {
	Amount          int    `json:"amount"`
	OrderType       string `json:"order_type"`
	Direction       string `json:"direction"`
	ValidEndTimeStr string `json:"valid_end_time_str"`
	MinAmount       int    `json:"min_amount"`
	MaxAmount       int    `json:"max_amount"`
	Concurrency     int    `json:"concurrency"`
}

// Config 应用配置
type Config struct {
	Server          ServerConfig      `json:"server"`
	Runtime         RuntimeConfig     `json:"runtime"`
	Logging         LogConfig         `json:"logging"`
	SamePosition    SamePositionConfig `json:"same_position"`
	Accounts        []AccountConfig    `json:"accounts"` // 备用账户列表
}

var (
	instance *Config
	once     sync.Once
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			BaseURL:  "http://127.0.0.1:8000",
			Timeout:  30,
			APIToken: "",
		},
		Runtime: RuntimeConfig{
			MaxWorkers: 10,
			RetryCount: 3,
			LogLevel:   "info",
		},
		Logging: LogConfig{
			Level:      "info",
			FilePath:   "logs/trade-manager.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
		},
		SamePosition: SamePositionConfig{
			SourceAccount: AccountConfig{
				XYAccount:    "",
				StockAccount: "",
				ClientType:   "1",
				Enabled:      false,
			},
			TargetAccounts: []AccountConfig{},
			OrderTemplate: OrderTemplate{
				Amount:          100,
				OrderType:       "1",
				Direction:       "1",
				ValidEndTimeStr: "",
				MinAmount:       0,
				MaxAmount:       0,
				Concurrency:     5,
			},
		},
		Accounts: []AccountConfig{},
	}
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	once.Do(func() {
		instance = DefaultConfig()
		
		// 尝试从配置文件加载
		if err := LoadConfig("config.json", instance); err != nil {
			fmt.Printf("警告: 无法加载配置文件: %v\n", err)
		}
		
		// 从环境变量覆盖
		loadFromEnv(instance)
		
		// 初始化全局日志记录器
		initGlobalLogger(instance)
	})
	return instance
}

// initGlobalLogger 初始化全局日志记录器
func initGlobalLogger(cfg *Config) {
	logConfig := logging.Config{
		Level:      cfg.Logging.Level,
		FilePath:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}
	
	logger, err := logging.NewLogger(logConfig)
	if err != nil {
		fmt.Printf("警告: 日志记录器初始化失败: %v\n", err)
		return
	}
	
	logging.SetGlobalLogger(logger)
	logger.Info("全局日志记录器已初始化，日志级别: %s", cfg.Logging.Level)
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string, cfg *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(cfg *Config, filename string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	return nil
}

// GetSourceAccount 获取源账户
func (c *Config) GetSourceAccount() (*AccountConfig, error) {
	if c.SamePosition.SourceAccount.Enabled && c.SamePosition.SourceAccount.XYAccount != "" {
		return &c.SamePosition.SourceAccount, nil
	}
	
	// 回退到从账户列表中查找启用的源账户
	for _, account := range c.Accounts {
		if account.Enabled {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("未找到启用的源账户")
}

// GetTargetAccounts 获取目标账户列表
func (c *Config) GetTargetAccounts() []AccountConfig {
	if len(c.SamePosition.TargetAccounts) > 0 {
		return c.SamePosition.TargetAccounts
	}
	
	// 回退到从账户列表中查找目标账户
	targets := []AccountConfig{}
	for _, account := range c.Accounts {
		if !account.Enabled {
			targets = append(targets, account)
		}
	}
	return targets
}

// ValidateSamePositionConfig 验证相同持仓配置
func (c *Config) ValidateSamePositionConfig() error {
	if !c.SamePosition.SourceAccount.Enabled {
		return fmt.Errorf("源账户未启用")
	}
	
	if c.SamePosition.SourceAccount.XYAccount == "" {
		return fmt.Errorf("源账户ID不能为空")
	}
	
	if len(c.SamePosition.TargetAccounts) == 0 {
		return fmt.Errorf("至少需要一个目标账户")
	}
	
	for i, target := range c.SamePosition.TargetAccounts {
		if target.XYAccount == "" {
			return fmt.Errorf("目标账户 %d 的账户ID不能为空", i+1)
		}
	}
	
	return nil
}

// GetAccountByID 根据账户ID获取账户配置
func (c *Config) GetAccountByID(accountID string) (*AccountConfig, error) {
	for _, account := range c.Accounts {
		if account.XYAccount == accountID {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("账户 %s 未找到", accountID)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	if url := os.Getenv("TRADE_API_URL"); url != "" {
		cfg.Server.BaseURL = url
	}
	if token := os.Getenv("TRADE_API_TOKEN"); token != "" {
		cfg.Server.APIToken = token
	}
	if workers := os.Getenv("TRADE_MAX_WORKERS"); workers != "" {
		// 这里可以添加解析逻辑
	}
}