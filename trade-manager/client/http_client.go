package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"trade-manager/config"
	"trade-manager/models"
)

// RestClient REST API客户端
type RestClient struct {
	baseURL    string
	httpClient *http.Client
	apiToken   string
}

// NewRestClient 创建新的REST客户端
func NewRestClient() *RestClient {
	cfg := config.GetConfig()
	return &RestClient{
		baseURL: cfg.Server.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Server.Timeout) * time.Second,
		},
		apiToken: cfg.Server.APIToken,
	}
}

// doRequest 执行HTTP请求
func (c *RestClient) doRequest(method, endpoint string, payload interface{}) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload failed: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Trade-Manager/1.0")
	req.Header.Set("Content-Type", "application/json")
	
	if c.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	return responseBody, nil
}

// QueryAccountPositions 查询账户持仓
func (c *RestClient) QueryAccountPositions(account models.AccountInfo) ([]models.PositionInfo, error) {
	endpoint := "/api/v1/positions"
	
	request := models.PositionQueryRequest{
		AccountInfo: account,
		PosStr:      "",
		QryNum:      1000,
	}
	
	response, err := c.doRequest("POST", endpoint, request)
	if err != nil {
		return nil, err
	}
	
	var positions []models.PositionInfo
	if err := json.Unmarshal(response, &positions); err != nil {
		return nil, fmt.Errorf("parse positions failed: %v", err)
	}
	
	// 设置账户信息
	for i := range positions {
		positions[i].Account = account.XYAccount
	}
	
	return positions, nil
}

// PlaceOrder 下单
func (c *RestClient) PlaceOrder(order models.OrderRequest) (models.OrderResponse, error) {
	endpoint := "/api/v1/orders"
	
	response, err := c.doRequest("POST", endpoint, order)
	if err != nil {
		return models.OrderResponse{
			Success:   false,
			Account:   order.XYAccount,
			StockCode: order.StockCode,
			Error:     err.Error(),
		}, err
	}
	
	var result models.OrderResponse
	if err := json.Unmarshal(response, &result); err != nil {
		return models.OrderResponse{
			Success:   false,
			Account:   order.XYAccount,
			StockCode: order.StockCode,
			Error:     fmt.Sprintf("parse response failed: %v", err),
		}, err
	}
	
	result.Success = true
	result.Account = order.XYAccount
	result.StockCode = order.StockCode
	
	return result, nil
}

// BatchPlaceOrders 批量下单
func (c *RestClient) BatchPlaceOrders(orders []models.OrderRequest) ([]models.OrderResponse, error) {
	endpoint := "/api/v1/orders/batch"
	
	response, err := c.doRequest("POST", endpoint, orders)
	if err != nil {
		return nil, err
	}
	
	var results []models.OrderResponse
	if err := json.Unmarshal(response, &results); err != nil {
		return nil, fmt.Errorf("parse batch response failed: %v", err)
	}
	
	return results, nil
}