package manager

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"trade-manager/client"
	"trade-manager/logging"
	"trade-manager/models"
)

// TradeManager 交易管理器
type TradeManager struct {
	client     *client.RestClient
	maxWorkers int
	logger     *logging.Logger
}

// NewTradeManager 创建交易管理器
func NewTradeManager() *TradeManager {
	return &TradeManager{
		client:     client.NewRestClient(),
		maxWorkers: 10,
		logger:     logging.GetLogger().WithFields(map[string]interface{}{"component": "trade_manager"}),
	}
}

// QueryAccountPositions 查询账户持仓
func (m *TradeManager) QueryAccountPositions(account models.AccountInfo) ([]models.PositionInfo, error) {
	m.logger.Info("查询账户 %s 持仓...", account.XYAccount)
	
	positions, err := m.client.QueryAccountPositions(account)
	if err != nil {
		m.logger.Error("查询账户 %s 持仓失败: %v", account.XYAccount, err)
		return nil, fmt.Errorf("查询账户 %s 持仓失败: %v", account.XYAccount, err)
	}
	
	m.logger.Info("账户 %s 共有 %d 个持仓", account.XYAccount, len(positions))
	return positions, nil
}

// SamePositionOrderFlow 相同持仓订单流程
func (m *TradeManager) SamePositionOrderFlow(ctx context.Context, config models.SamePositionConfig) (*models.SamePositionResult, error) {
	result := &models.SamePositionResult{
		SourceAccount: config.SourceAccount.XYAccount,
		StartTime:     time.Now(),
	}
	
	// 1. 查询源账号持仓
	m.logger.Info("开始相同持仓订单流程，源账号: %s", config.SourceAccount.XYAccount)
	
	positions, err := m.QueryAccountPositions(config.SourceAccount)
	if err != nil {
		return result, err
	}
	
	result.Positions = positions
	
	// 2. 过滤持仓
	filteredPositions := m.filterPositions(positions, config.OrderTemplate.MinAmount, config.OrderTemplate.MaxAmount)
	if len(filteredPositions) == 0 {
		m.logger.Warn("没有符合条件的持仓（最小持仓量: %d）", config.OrderTemplate.MinAmount)
		m.finalizeResult(result)
		return result, nil
	}
	
	m.logger.Info("过滤后共有 %d 个持仓符合条件", len(filteredPositions))
	
	// 3. 生成订单
	generatedOrders := m.generateSamePositionOrders(filteredPositions, config.TargetAccounts, config.OrderTemplate)
	result.GeneratedOrders = generatedOrders
	result.TotalCount = len(generatedOrders)
	
	if len(generatedOrders) == 0 {
		m.logger.Warn("没有生成任何订单")
		m.finalizeResult(result)
		return result, nil
	}
	
	m.logger.Info("为 %d 个目标账号生成 %d 个订单", len(config.TargetAccounts), len(generatedOrders))
	
	// 4. 批量执行订单
	orderResults, err := m.batchPlaceOrders(ctx, generatedOrders, config.OrderTemplate.Concurrency)
	if err != nil {
		m.logger.Error("批量下单失败: %v", err)
		return result, fmt.Errorf("批量下单失败: %v", err)
	}
	
	result.OrderResults = orderResults
	
	// 5. 统计结果
	for _, res := range orderResults {
		if res.Success {
			result.SuccessCount++
		}
	}
	
	m.finalizeResult(result)
	m.logger.Info("相同持仓订单流程完成：成功 %d/%d，耗时 %v", 
		result.SuccessCount, result.TotalCount, result.Duration)
	
	return result, nil
}

// filterPositions 根据持仓数量过滤持仓
func (m *TradeManager) filterPositions(positions []models.PositionInfo, minAmount, maxAmount int) []models.PositionInfo {
	if minAmount <= 0 && maxAmount <= 0 {
		return positions
	}
	
	filtered := make([]models.PositionInfo, 0)
	for _, pos := range positions {
		if (minAmount <= 0 || pos.CurrentAmount >= minAmount) && 
		   (maxAmount <= 0 || pos.CurrentAmount <= maxAmount) {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// generateSamePositionOrders 生成相同持仓订单
func (m *TradeManager) generateSamePositionOrders(
	positions []models.PositionInfo, 
	targetAccounts []models.AccountInfo, 
	template models.OrderTemplate,
) []models.OrderRequest {
	orders := make([]models.OrderRequest, 0)
	
	for _, account := range targetAccounts {
		for _, position := range positions {
			order := models.OrderRequest{
				AccountInfo:     account,
				StockCode:       position.StockCode,
				Amount:          template.Amount,
				OrderType:       template.OrderType,
				Direction:       template.Direction,
				ValidEndTimeStr: template.ValidEndTimeStr,
			}
			
			// 设置默认值
			if order.Amount <= 0 {
				order.Amount = 100 // 默认下单数量
			}
			if order.OrderType == "" {
				order.OrderType = "1" // 默认限价单
			}
			if order.Direction == "" {
				order.Direction = "1" // 默认买入
			}
			
			orders = append(orders, order)
		}
	}
	
	return orders
}

// batchPlaceOrders 批量下单
func (m *TradeManager) batchPlaceOrders(ctx context.Context, orders []models.OrderRequest, concurrency int) ([]models.OrderResponse, error) {
	if concurrency <= 0 {
		concurrency = m.maxWorkers
	}
	
	var wg sync.WaitGroup
	results := make([]models.OrderResponse, 0, len(orders))
	resultChan := make(chan models.OrderResponse, len(orders))
	errorChan := make(chan error, 1)
	
	// 创建工作池
	taskChan := make(chan models.OrderRequest, len(orders))
	
	// 启动工作goroutine
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go m.orderWorker(ctx, &wg, taskChan, resultChan, errorChan)
	}
	
	// 分发任务
	go func() {
		defer close(taskChan)
		
		for _, order := range orders {
			select {
			case taskChan <- order:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	// 收集结果
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 处理结果
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				return results, nil
			}
			results = append(results, result)
			
		case err := <-errorChan:
			return results, err
			
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}
}

// orderWorker 订单工作goroutine
func (m *TradeManager) orderWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	tasks <-chan models.OrderRequest,
	results chan<- models.OrderResponse,
	errors chan<- error,
) {
	defer wg.Done()
	
	for task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
			response, err := m.client.PlaceOrder(task)
			if err != nil {
				response.Success = false
				response.Error = err.Error()
			}
			
			select {
			case results <- response:
			case <-ctx.Done():
				return
			}
		}
	}
}

// finalizeResult 完成结果统计
func (m *TradeManager) finalizeResult(result *models.SamePositionResult) {
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
}

// BatchOrder 批量下单
func (m *TradeManager) BatchOrder(ctx context.Context, orders []models.OrderRequest) []models.OrderResponse {
	m.logger.Info("开始批量下单，订单数量: %d", len(orders))
	
	if len(orders) == 0 {
		m.logger.Warn("批量下单: 订单列表为空")
		return []models.OrderResponse{}
	}
	
	results, err := m.batchPlaceOrders(ctx, orders, m.maxWorkers)
	if err != nil {
		m.logger.Error("批量下单失败: %v", err)
		// 返回错误结果
		errorResults := make([]models.OrderResponse, len(orders))
		for i, order := range orders {
			errorResults[i] = models.OrderResponse{
				Success:   false,
				Account:   order.XYAccount,
				StockCode: order.StockCode,
				Error:     err.Error(),
			}
		}
		return errorResults
	}
	
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	
	m.logger.Info("批量下单完成: 成功 %d/%d", successCount, len(orders))
	return results
}

// GetPositionSummary 获取持仓汇总
func (m *TradeManager) GetPositionSummary(accounts []models.AccountInfo) (map[string]interface{}, error) {
	m.logger.Info("开始获取 %d 个账户的持仓汇总", len(accounts))
	
	summary := make(map[string]interface{})
	totalPositions := 0
	totalValue := 0.0
	stockCount := make(map[string]int)
	
	for _, account := range accounts {
		positions, err := m.QueryAccountPositions(account)
		if err != nil {
			m.logger.Warn("查询账户 %s 持仓失败: %v", account.XYAccount, err)
			continue
		}
		
		totalPositions += len(positions)
		for _, pos := range positions {
			totalValue += pos.MarketValue
			stockCount[pos.StockCode]++
		}
	}
	
	summary["total_accounts"] = len(accounts)
	summary["total_positions"] = totalPositions
	summary["total_market_value"] = totalValue
	summary["distinct_stocks"] = len(stockCount)
	summary["positions_by_stock"] = stockCount
	
	m.logger.Info("持仓汇总完成: 账户数 %d, 持仓数 %d, 总市值 %.2f", 
		len(accounts), totalPositions, totalValue)
	
	return summary, nil
}