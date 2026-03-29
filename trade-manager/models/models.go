package models

import "time"

// AccountInfo 账户信息
type AccountInfo struct {
	XYAccount    string `json:"XyBusinAccount"`
	StockAccount string `json:"stockAccount"`
	ClientType   string `json:"clientType"`
}

// PositionInfo 持仓信息
type PositionInfo struct {
	StockCode     string  `json:"StockCode"`
	StockName     string  `json:"StockName"`
	CurrentAmount int     `json:"CurrentAmount"`
	MarketValue   float64 `json:"MarketValue"`
	CostPrice     float64 `json:"CostPrice"`
	Account       string  `json:"account"`
}

// OrderRequest 下单请求
type OrderRequest struct {
	AccountInfo
	StockCode       string `json:"stockCode"`
	Amount          int    `json:"Amount"`
	Price           string `json:"price,omitempty"`
	OrderType       string `json:"orderType"` // 1-限价 2-市价
	Direction       string `json:"direction"` // 1-买入 2-卖出
	ValidEndTimeStr string `json:"validEndTimeStr"`
}

// OrderResponse 下单响应
type OrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	OrderID string `json:"order_id,omitempty"`
	Error   string `json:"error,omitempty"`
	Account string `json:"account"`
	StockCode string `json:"stockCode"`
}

// PositionQueryRequest 持仓查询请求
type PositionQueryRequest struct {
	AccountInfo
	PosStr string `json:"PosStr"`
	QryNum int    `json:"QryNum"`
}

// SamePositionConfig 相同持仓配置
type SamePositionConfig struct {
	SourceAccount  AccountInfo   `json:"source_account"`
	TargetAccounts []AccountInfo `json:"target_accounts"`
	OrderTemplate  OrderTemplate  `json:"order_template"`
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

// SamePositionResult 相同持仓结果
type SamePositionResult struct {
	SourceAccount   string          `json:"source_account"`
	Positions       []PositionInfo  `json:"positions"`
	GeneratedOrders []OrderRequest  `json:"generated_orders"`
	OrderResults    []OrderResponse `json:"order_results"`
	SuccessCount    int             `json:"success_count"`
	TotalCount      int             `json:"total_count"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Duration        time.Duration   `json:"duration"`
}