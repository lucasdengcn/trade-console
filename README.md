# trade-console

## Overview

This project is a trade console application designed to manage and execute trades efficiently. It provides a user-friendly interface for traders to monitor market conditions, place orders, and manage their portfolios.

## Features

- Real-time market data updates
- Order placement and management
- Portfolio tracking and performance analysis
- Customizable alerts and notifications
- Integration with multiple trading platforms

## Installation

To install the trade console application, follow these steps:

1. Clone the repository:

   ```
   git clone https://github.com/yourusername/trade-console.git
   ```

2. Navigate to the project directory:

   ```
   cd trade-console
   ```

3. Build the application:

   ```
   go build -o trade-console
   ```

4. Run the application:

   ```
   ./trade-console
   ```

## Usage

```bash

# 初始化配置
trade-cli --init

# 列出账户
trade-cli --list-accounts

# 查询持仓（自动使用配置中的源账户）
trade-cli --query

# 相同持仓订单流程（自动使用配置中的源账户）
trade-cli --same-position --config same-position.json

# 使用特定配置文件
trade-cli --config my-config.json --query

# 指定源账户（覆盖配置文件）
trade-cli --source 88060819 --query

# 带过滤条件的相同持仓订单
trade-cli --same-position --config orders.json --min-amount 100 --concurrency 10

```

### Configuration

The application can be configured using a JSON file. Below is an example of the configuration file structure:

```json
{
  "target_accounts": [
    {
      "XyBusinAccount": "88060820",
      "stockAccount": "0680048596",
      "clientType": "1"
    }
  ],
  "order_template": {
    "Amount": 1000,
    "orderType": "1",
    "direction": "1",
    "validEndTimeStr": "2023-12-31"
  }
}
``
