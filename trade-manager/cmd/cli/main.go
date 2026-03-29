package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	
	"trade-manager/config"
	"trade-manager/manager"
	"trade-manager/models"
)

// CLI 命令行接口
type CLI struct {
	flags     *Flags
	manager   *manager.TradeManager
	config    *config.Config
}

// Flags 命令行参数
type Flags struct {
	ConfigFile    *string
	QueryPos      *bool
	SamePosition  *bool
	SourceAccount *string
	MinAmount     *int
	MaxAmount     *int
	Concurrency   *int
	Help          *bool
	InitConfig    *bool
	ListAccounts  *bool
}

// NewCLI 创建新的CLI实例
func NewCLI() *CLI {
	return &CLI{
		flags: &Flags{
			ConfigFile:    flag.String("config", "", "配置文件路径 (默认: config.json)"),
			QueryPos:      flag.Bool("query", false, "查询持仓"),
			SamePosition:  flag.Bool("same-position", false, "相同持仓订单流程"),
			SourceAccount: flag.String("source", "", "源账号（用于查询持仓）"),
			MinAmount:     flag.Int("min-amount", 0, "最小持仓数量"),
			MaxAmount:     flag.Int("max-amount", 0, "最大持仓数量"),
			Concurrency:   flag.Int("concurrency", 5, "并发数"),
			Help:          flag.Bool("help", false, "显示帮助"),
			InitConfig:    flag.Bool("init", false, "初始化配置文件"),
			ListAccounts:  flag.Bool("list-accounts", false, "列出配置中的账户"),
		},
		manager: manager.NewTradeManager(),
		config:  config.GetConfig(),
	}
}

// Run 运行CLI
func (c *CLI) Run() error {
	flag.Parse()
	
	if *c.flags.Help {
		c.printHelp()
		return nil
	}
	
	if *c.flags.InitConfig {
		return c.initConfig()
	}
	
	if *c.flags.ListAccounts {
		return c.listAccounts()
	}
	
	configFile := c.getConfigFile()
	
	if *c.flags.QueryPos {
		return c.handleQueryPositions(configFile)
	}
	
	if *c.flags.SamePosition {
		return c.handleSamePositionOrder(configFile)
	}
	
	c.printHelp()
	return nil
}

// getConfigFile 获取配置文件路径
func (c *CLI) getConfigFile() string {
	if *c.flags.ConfigFile != "" {
		return *c.flags.ConfigFile
	}
	return "config.json"
}

// handleQueryPositions 处理持仓查询
func (c *CLI) handleQueryPositions(configFile string) error {
	if *c.flags.SourceAccount != "" {
		// 查询指定账号持仓
		account := models.AccountInfo{XYAccount: *c.flags.SourceAccount}
		positions, err := c.manager.QueryAccountPositions(account)
		if err != nil {
			return fmt.Errorf("查询账号 %s 持仓失败: %v", *c.flags.SourceAccount, err)
		}
		
		data, _ := json.MarshalIndent(positions, "", "  ")
		fmt.Printf("账号 %s 持仓详情:\n", *c.flags.SourceAccount)
		fmt.Println(string(data))
	} else {
		// 查询所有账号持仓汇总
		accounts := c.getAccountsFromConfig(configFile)
		summary, err := c.manager.GetPositionSummary(accounts)
		if err != nil {
			return fmt.Errorf("查询持仓汇总失败: %v", err)
		}
		
		data, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println("持仓汇总:")
		fmt.Println(string(data))
	}
	return nil
}

// handleSamePositionOrder 处理相同持仓订单流程
func (c *CLI) handleSamePositionOrder(configFile string) error {
	sourceAccount := *c.flags.SourceAccount
	if sourceAccount == "" {
		// 尝试从配置文件获取源账户
		if account, err := c.config.GetSourceAccount(); err == nil {
			sourceAccount = account.XYAccount
		} else {
			return fmt.Errorf("相同持仓订单流程需要提供源账号 (使用 --source 或配置文件中启用源账户)")
		}
	}
	
	if configFile == "" {
		return fmt.Errorf("相同持仓订单流程需要提供配置文件")
	}
	
	var samePositionConfig models.SamePositionConfig
	if err := c.loadConfigFile(configFile, &samePositionConfig); err != nil {
		return fmt.Errorf("加载配置文件失败: %v", err)
	}
	
	// 设置源账号
	samePositionConfig.SourceAccount = models.AccountInfo{XYAccount: sourceAccount}
	
	// 设置过滤参数到订单模板
	samePositionConfig.OrderTemplate.MinAmount = *c.flags.MinAmount
	samePositionConfig.OrderTemplate.MaxAmount = *c.flags.MaxAmount
	samePositionConfig.OrderTemplate.Concurrency = *c.flags.Concurrency
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	result, err := c.manager.SamePositionOrderFlow(ctx, samePositionConfig)
	if err != nil {
		return fmt.Errorf("相同持仓订单流程执行失败: %v", err)
	}
	
	// 输出详细结果
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("相同持仓订单流程结果:")
	fmt.Println(string(data))
	
	return nil
}

// initConfig 初始化配置文件
func (c *CLI) initConfig() error {
	defaultConfig := config.DefaultConfig()
	
	// 添加示例账户配置
	defaultConfig.Accounts = []config.AccountConfig{
		{
			XYAccount:    "88060819",
			StockAccount: "0680048595",
			ClientType:   "1",
			Enabled:      true,
		},
		{
			XYAccount:    "88060820", 
			StockAccount: "0680048596",
			ClientType:   "1",
			Enabled:      false,
		},
	}
	
	filename := c.getConfigFile()
	if err := config.SaveConfig(defaultConfig, filename); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}
	
	fmt.Printf("配置文件已初始化: %s\n", filename)
	return nil
}

// listAccounts 列出配置中的账户
func (c *CLI) listAccounts() error {
	configFile := c.getConfigFile()
	
	var cfg struct {
		Accounts []config.AccountConfig `json:"accounts"`
	}
	
	if err := c.loadConfigFile(configFile, &cfg); err != nil {
		return fmt.Errorf("加载配置文件失败: %v", err)
	}
	
	if len(cfg.Accounts) == 0 {
		fmt.Println("配置文件中未找到账户配置")
		return nil
	}
	
	fmt.Printf("配置文件中找到 %d 个账户:\n", len(cfg.Accounts))
	for i, account := range cfg.Accounts {
		status := "禁用"
		if account.Enabled {
			status = "启用"
		}
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, account.XYAccount, account.StockAccount, status)
	}
	
	return nil
}

// loadConfigFile 加载配置文件
func (c *CLI) loadConfigFile(filename string, target interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	return nil
}

// getAccountsFromConfig 从配置文件获取账户列表
func (c *CLI) getAccountsFromConfig(configFile string) []models.AccountInfo {
	if configFile == "" {
		return []models.AccountInfo{}
	}
	
	var cfg struct {
		Accounts []config.AccountConfig `json:"accounts"`
	}
	
	if err := c.loadConfigFile(configFile, &cfg); err != nil {
		log.Printf("加载配置文件失败: %v", err)
		return []models.AccountInfo{}
	}
	
	accounts := make([]models.AccountInfo, 0, len(cfg.Accounts))
	for _, acc := range cfg.Accounts {
		accounts = append(accounts, models.AccountInfo{
			XYAccount:    acc.XYAccount,
			StockAccount: acc.StockAccount,
			ClientType:   acc.ClientType,
		})
	}
	
	return accounts
}
// printHelp 显示帮助信息
func (c *CLI) printHelp() {
	helpText := `交易管理器 - 多账号持仓同步下单工具
 
用法:
  trade-cli [命令] [选项]
 
命令:
  --query                查询持仓
  --same-position        相同持仓订单流程
  --init                 初始化配置文件
  --list-accounts        列出配置中的账户
 
选项:
  --config string        配置文件路径 (默认: config.json)
  --source string        源账号（用于查询持仓）
  --min-amount int       最小持仓数量（默认: 0）
  --max-amount int       最大持仓数量（默认: 0）
  --concurrency int      并发数（默认: 5）
  --help                 显示帮助
 
示例:
  # 初始化配置文件
  trade-cli --init
  
  # 列出账户
  trade-cli --list-accounts
  
  # 查询指定账号持仓
  trade-cli --query --source 88060819
  
  # 查询所有账号持仓汇总
  trade-cli --query --config my-config.json
  
  # 执行相同持仓订单流程
  trade-cli --same-position --config same-position.json --source 88060819 --min-amount 100
 
配置文件结构:
{
  "server": {
    "base_url": "http://127.0.0.1:8000",
    "timeout": 30,
    "api_token": ""
  },
  "runtime": {
    "max_workers": 10,
    "retry_count": 3,
    "log_level": "info"
  },
  "accounts": [
    {
      "XyBusinAccount": "88060819",
      "stockAccount": "0680048595",
      "clientType": "1",
      "enabled": true
    }
  ]
}`
	
	fmt.Println(helpText)
}

func main() {
	cli := NewCLI()
	if err := cli.Run(); err != nil {
		log.Fatalf("错误: %v", err)
	}
}