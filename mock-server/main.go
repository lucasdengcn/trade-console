package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"os"
	"github.com/gorilla/mux"
)

// MockPosition 模拟持仓数据
type MockPosition struct {
	StockCode     string  `json:"StockCode"`
	StockName     string  `json:"StockName"`
	CurrentAmount int     `json:"CurrentAmount"`
	MarketValue   float64 `json:"MarketValue"`
	CostPrice     float64 `json:"CostPrice"`
}

// MockOrderRequest 模拟下单请求
type MockOrderRequest struct {
	XyBusinAccount string `json:"XyBusinAccount"`
	StockAccount   string `json:"stockAccount"`
	ClientType     string `json:"clientType"`
	StockCode      string `json:"stockCode"`
	Amount         int    `json:"Amount"`
	OrderType      string `json:"orderType"`
	Direction      string `json:"direction"`
}

// MockOrderResponse 模拟下单响应
type MockOrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	OrderID string `json:"order_id"`
	Account string `json:"account"`
	StockCode string `json:"stockCode"`
}

// MockServer 模拟服务器
type MockServer struct {
	router *mux.Router
	port   int
	// 模拟数据存储
	positions map[string][]MockPosition // account -> positions
	orders    map[string][]MockOrderResponse // account -> orders
}

// NewMockServer 创建新的模拟服务器
func NewMockServer(port int) *MockServer {
	server := &MockServer{
		router:    mux.NewRouter(),
		port:      port,
		positions: make(map[string][]MockPosition),
		orders:    make(map[string][]MockOrderResponse),
	}
	
	// 初始化模拟数据
	server.initMockData()
	
	// 设置路由
	server.setupRoutes()
	
	return server
}

// initMockData 初始化模拟数据
func (s *MockServer) initMockData() {
	// 为几个测试账户初始化持仓数据
	testAccounts := []string{"88060819", "88060820", "88060821"}
	
	for _, account := range testAccounts {
		positions := []MockPosition{
			{
				StockCode:     "000001",
				StockName:     "平安银行",
				CurrentAmount: 1000 + rand.Intn(5000),
				MarketValue:   15000.0 + rand.Float64()*50000,
				CostPrice:     15.0 + rand.Float64()*10,
			},
			{
				StockCode:     "000002",
				StockName:     "万科A",
				CurrentAmount: 500 + rand.Intn(3000),
				MarketValue:   8000.0 + rand.Float64()*30000,
				CostPrice:     16.0 + rand.Float64()*8,
			},
			{
				StockCode:     "300152",
				StockName:     "科融环境",
				CurrentAmount: 2000 + rand.Intn(8000),
				MarketValue:   12000.0 + rand.Float64()*40000,
				CostPrice:     6.0 + rand.Float64()*4,
			},
		}
		s.positions[account] = positions
	}
}

// setupRoutes 设置路由
func (s *MockServer) setupRoutes() {
	// 持仓查询接口
	s.router.HandleFunc("/api/v1/positions", s.handleQueryPositions).Methods("POST")
	
	// 单个下单接口
	s.router.HandleFunc("/api/v1/orders", s.handlePlaceOrder).Methods("POST")
	
	// 批量下单接口
	s.router.HandleFunc("/api/v1/orders/batch", s.handleBatchOrders).Methods("POST")
	
	// 健康检查接口
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
	
	// 账户信息接口
	s.router.HandleFunc("/api/v1/accounts/{account}/positions", s.handleAccountPositions).Methods("GET")
	
	// 订单查询接口
	s.router.HandleFunc("/api/v1/accounts/{account}/orders", s.handleAccountOrders).Methods("GET")
}

// handleQueryPositions 处理持仓查询
func (s *MockServer) handleQueryPositions(w http.ResponseWriter, r *http.Request) {
	var request struct {
		XyBusinAccount string `json:"XyBusinAccount"`
		PosStr         string `json:"PosStr"`
		QryNum         int    `json:"QryNum"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, "解析请求失败", http.StatusBadRequest)
		return
	}
	
	account := request.XyBusinAccount
	if account == "" {
		s.sendError(w, "账户ID不能为空", http.StatusBadRequest)
		return
	}
	
	positions, exists := s.positions[account]
	if !exists {
		s.sendError(w, fmt.Sprintf("账户 %s 不存在", account), http.StatusNotFound)
		return
	}
	
	// 模拟查询延迟
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	
	s.sendJSON(w, positions, http.StatusOK)
}

// handlePlaceOrder 处理下单请求
func (s *MockServer) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var request MockOrderRequest
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, "解析下单请求失败", http.StatusBadRequest)
		return
	}
	
	// 验证必填字段
	if request.XyBusinAccount == "" || request.StockCode == "" || request.Amount <= 0 {
		s.sendError(w, "缺少必填字段", http.StatusBadRequest)
		return
	}
	
	// 模拟下单处理延迟
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	
	// 模拟成功率（90%成功）
	success := rand.Float32() < 0.9
	
	response := MockOrderResponse{
		Success:   success,
		Account:   request.XyBusinAccount,
		StockCode: request.StockCode,
	}
	
	if success {
		response.Message = "下单成功"
		response.OrderID = fmt.Sprintf("ORDER_%d_%d", time.Now().Unix(), rand.Intn(1000))
		
		// 记录订单
		s.orders[request.XyBusinAccount] = append(s.orders[request.XyBusinAccount], response)
	} else {
		response.Message = "下单失败：模拟错误"
	}
	
	s.sendJSON(w, response, http.StatusOK)
}

// handleBatchOrders 处理批量下单
func (s *MockServer) handleBatchOrders(w http.ResponseWriter, r *http.Request) {
	var requests []MockOrderRequest
	
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		s.sendError(w, "解析批量下单请求失败", http.StatusBadRequest)
		return
	}
	
	if len(requests) == 0 {
		s.sendError(w, "订单列表不能为空", http.StatusBadRequest)
		return
	}
	
	responses := make([]MockOrderResponse, 0, len(requests))
	
	// 模拟批量处理
	for _, req := range requests {
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		
		success := rand.Float32() < 0.9
		
		response := MockOrderResponse{
			Success:   success,
			Account:   req.XyBusinAccount,
			StockCode: req.StockCode,
		}
		
		if success {
			response.Message = "下单成功"
			response.OrderID = fmt.Sprintf("BATCH_ORDER_%d_%d", time.Now().Unix(), rand.Intn(1000))
			s.orders[req.XyBusinAccount] = append(s.orders[req.XyBusinAccount], response)
		} else {
			response.Message = "下单失败：模拟错误"
		}
		
		responses = append(responses, response)
	}
	
	s.sendJSON(w, responses, http.StatusOK)
}

// handleHealthCheck 健康检查
func (s *MockServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	s.sendJSON(w, response, http.StatusOK)
}

// handleAccountPositions 查询指定账户持仓
func (s *MockServer) handleAccountPositions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	account := vars["account"]
	
	positions, exists := s.positions[account]
	if !exists {
		s.sendError(w, fmt.Sprintf("账户 %s 不存在", account), http.StatusNotFound)
		return
	}
	
	s.sendJSON(w, positions, http.StatusOK)
}

// handleAccountOrders 查询指定账户订单
func (s *MockServer) handleAccountOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	account := vars["account"]
	
	orders, exists := s.orders[account]
	if !exists {
		orders = []MockOrderResponse{}
	}
	
	s.sendJSON(w, orders, http.StatusOK)
}

// sendJSON 发送JSON响应
func (s *MockServer) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON编码失败: %v", err)
	}
}

// sendError 发送错误响应
func (s *MockServer) sendError(w http.ResponseWriter, message string, statusCode int) {
	s.sendJSON(w, map[string]interface{}{
		"error":   message,
		"success": false,
	}, statusCode)
}

// Start 启动服务器
func (s *MockServer) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("模拟服务器启动在端口 %d", s.port)
	log.Printf("健康检查: http://localhost:%d/health", s.port)
	log.Printf("持仓查询: POST http://localhost:%d/api/v1/positions", s.port)
	log.Printf("下单接口: POST http://localhost:%d/api/v1/orders", s.port)
	
	return http.ListenAndServe(addr, s.router)
}

func main() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())
	
	// 默认端口8080，可以通过命令行参数修改
	port := 8080
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil && p > 0 {
			port = p
		}
	}
	
	server := NewMockServer(port)
	
	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}