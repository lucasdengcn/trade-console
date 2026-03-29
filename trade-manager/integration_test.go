package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"trade-manager/config"
	"trade-manager/manager"
	"trade-manager/models"
)

// TestMain 测试主函数，启动模拟服务器
func TestMain(m *testing.M) {
	// 启动模拟服务器在测试端口
	go startMockServer(8081)
	
	// 等待服务器启动
	time.Sleep(2 * time.Second)
	
	// 运行测试
	code := m.Run()
	
	os.Exit(code)
}

// startMockServer 启动模拟服务器
func startMockServer(port int) {
	// 这里可以调用模拟服务器的启动逻辑
	// 在实际实现中，需要导入模拟服务器包并启动
	fmt.Printf("模拟服务器启动在端口 %d\n", port)
}

// TestQueryPositions 测试持仓查询
func TestQueryPositions(t *testing.T) {
	// 设置测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "http://localhost:8081",
			Timeout: 10,
		},
	}
	
	// 创建交易管理器
	mgr := manager.NewTradeManager()
	
	// 测试账户
	testAccount := models.AccountInfo{
		XYAccount:    "88060819",
		StockAccount: "0680048595",
		ClientType:   "1",
	}
	
	// 查询持仓
	positions, err := mgr.QueryAccountPositions(testAccount)
	if err != nil {
		t.Fatalf("持仓查询失败: %v", err)
	}
	
	if len(positions) == 0 {
		t.Error("预期有持仓数据，但返回为空")
	}
	
	t.Logf("查询到 %d 个持仓", len(positions))
	for _, pos := range positions {
		t.Logf("股票: %s, 数量: %d, 市值: %.2f", pos.StockCode, pos.CurrentAmount, pos.MarketValue)
	}
}

// TestSamePositionOrderFlow 测试相同持仓订单流程
func TestSamePositionOrderFlow(t *testing.T) {
	mgr := manager.NewTradeManager()
	
	// 测试配置
	config := models.SamePositionConfig{
		SourceAccount: models.AccountInfo{
			XYAccount:    "88060819",
			StockAccount: "0680048595",
			ClientType:   "1",
		},
		TargetAccounts: []models.AccountInfo{
			{
				XYAccount:    "88060820",
				StockAccount: "0680048596",
				ClientType:   "1",
			},
			{
				XYAccount:    "88060821",
				StockAccount: "0680048597",
				ClientType:   "1",
			},
		},
		OrderTemplate: models.OrderTemplate{
			Amount:          100,
			OrderType:       "1",
			Direction:       "1",
			ValidEndTimeStr: "2023-12-31",
			MinAmount:       100,
			MaxAmount:       10000,
			Concurrency:     2,
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 执行相同持仓订单流程
	result, err := mgr.SamePositionOrderFlow(ctx, config)
	if err != nil {
		t.Fatalf("相同持仓订单流程失败: %v", err)
	}
	
	// 验证结果
	if result.SourceAccount != config.SourceAccount.XYAccount {
		t.Errorf("源账户不匹配: 期望 %s, 实际 %s", config.SourceAccount.XYAccount, result.SourceAccount)
	}
	
	if len(result.Positions) == 0 {
		t.Error("预期源账户有持仓数据")
	}
	
	if result.TotalCount == 0 {
		t.Error("预期生成了订单")
	}
	
	t.Logf("测试结果: 源账户 %s, 持仓 %d, 生成订单 %d, 成功 %d",
		result.SourceAccount, len(result.Positions), result.TotalCount, result.SuccessCount)
}

// TestBatchOrder 测试批量下单
func TestBatchOrder(t *testing.T) {
	mgr := manager.NewTradeManager()
	
	// 测试订单
	testOrders := []models.OrderRequest{
		{
			AccountInfo: models.AccountInfo{
				XYAccount:    "88060819",
				StockAccount: "0680048595",
				ClientType:   "1",
			},
			StockCode:       "000001",
			Amount:          100,
			OrderType:       "1",
			Direction:       "1",
			ValidEndTimeStr: "2023-12-31",
		},
		{
			AccountInfo: models.AccountInfo{
				XYAccount:    "88060820",
				StockAccount: "0680048596",
				ClientType:   "1",
			},
			StockCode:       "000002",
			Amount:          200,
			OrderType:       "1",
			Direction:       "1",
			ValidEndTimeStr: "2023-12-31",
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 执行批量下单
	results := mgr.BatchOrder(ctx, testOrders)
	
	if len(results) != len(testOrders) {
		t.Errorf("结果数量不匹配: 期望 %d, 实际 %d", len(testOrders), len(results))
	}
	
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
		t.Logf("订单结果: 账户 %s, 股票 %s, 成功: %v", result.Account, result.StockCode, result.Success)
	}
	
	t.Logf("批量下单结果: 总数 %d, 成功 %d", len(results), successCount)
}

// TestErrorScenarios 测试错误场景
func TestErrorScenarios(t *testing.T) {
	mgr := manager.NewTradeManager()
	
	// 测试不存在的账户
	nonexistentAccount := models.AccountInfo{
		XYAccount:    "99999999",
		StockAccount: "9999999999",
		ClientType:   "1",
	}
	
	_, err := mgr.QueryAccountPositions(nonexistentAccount)
	if err == nil {
		t.Error("预期查询不存在的账户应该返回错误")
	}
	
	// 测试无效的订单
	invalidOrder := models.OrderRequest{
		AccountInfo: models.AccountInfo{
			XYAccount:    "88060819",
			StockAccount: "0680048595",
			ClientType:   "1",
		},
		StockCode:       "", // 空的股票代码
		Amount:          0,  // 无效的数量
		OrderType:       "1",
		Direction:       "1",
		ValidEndTimeStr: "2023-12-31",
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	results := mgr.BatchOrder(ctx, []models.OrderRequest{invalidOrder})
	if len(results) > 0 && results[0].Success {
		t.Error("预期无效订单应该失败")
	}
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	mgr := manager.NewTradeManager()
	
	// 创建多个并发查询
	accounts := []models.AccountInfo{
		{XYAccount: "88060819", StockAccount: "0680048595", ClientType: "1"},
		{XYAccount: "88060820", StockAccount: "0680048596", ClientType: "1"},
		{XYAccount: "88060821", StockAccount: "0680048597", ClientType: "1"},
	}
	
	// 并发查询持仓
	results := make(chan []models.PositionInfo, len(accounts))
	errors := make(chan error, len(accounts))
	
	for _, account := range accounts {
		go func(acc models.AccountInfo) {
			positions, err := mgr.QueryAccountPositions(acc)
			if err != nil {
				errors <- err
				return
			}
			results <- positions
		}(account)
	}
	
	// 收集结果
	timeout := time.After(10 * time.Second)
	completed := 0
	
	for completed < len(accounts) {
		select {
		case positions := <-results:
			completed++
			t.Logf("账户 %s 查询到 %d 个持仓", accounts[completed-1].XYAccount, len(positions))
		case err := <-errors:
			completed++
			t.Errorf("并发查询失败: %v", err)
		case <-timeout:
			t.Fatal("并发查询超时")
		}
	}
}