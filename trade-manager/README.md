# Trade Manager - Multi-Account Position Synchronization Tool

Trade Manager is a Go-based multi-account concurrent trading management tool specifically designed for implementing "place orders for target accounts based on source account positions" business scenarios.

## 🚀 Core Features

### 1. Multi-Account Position Query

- Query real-time position information for specified accounts
- Provide position summary statistics
- Support position filtering by quantity

### 2. Same Position Order Flow

- **Source Account Position Query**: Automatically query stocks held by source account
- **Position Filtering**: Filter positions by minimum/maximum quantity
- **Order Generation**: Generate orders with same stock codes for target accounts
- **Concurrent Execution**: Support multi-account concurrent order placement
- **Result Statistics**: Provide detailed execution results and performance metrics

### 3. Batch Order Management

- Support single and batch order operations
- Configurable concurrency and timeout settings
- Comprehensive error handling and retry mechanisms

## 📁 Project Structure

```
trade-manager/
├── client/           # HTTP API Client
│   └── http_client.go
├── cmd/              # Command Line Interface
│   └── cli/
│       └── main.go
├── config/           # Configuration Management
│   └── config.go
├── logging/          # Logging System
│   └── logger.go
├── manager/          # Core Trade Manager
│   └── trade_manager.go
├── models/           # Data Models
│   └── models.go
├── config.json       # Configuration File Example
├── go.mod           # Go Module File
├── integration_test.go # Integration Tests
└── TEST.md          # Test Documentation
```

## 🛠️ Quick Start

### Prerequisites

- Go 1.21 or higher
- Trading API service (supporting RESTful interfaces)

### Installation and Running

1. **Clone the Project**

```bash
git clone <repository-url>
cd trade-manager
```

1. **Build the Application**

```bash
# Download dependencies
go mod tidy

# Build the CLI tool
go build -o trade-cli cmd/cli/main.go

# Alternative: Build with optimizations
go build -ldflags="-s -w" -o trade-cli cmd/cli/main.go

# Verify the build
./trade-cli --version || ./trade-cli --help
```

1. **Initialize Configuration**

```bash
go run cmd/cli/main.go --init
```

1. **View Help**

```bash
go run cmd/cli/main.go --help
```

1. **Query Positions**

```bash
# Query specific account positions
go run cmd/cli/main.go --query --source 88060819

# Query all account position summary
go run cmd/cli/main.go --query --config config.json
```

1. **Execute Same Position Order Flow**

```bash
go run cmd/cli/main.go --same-position --config same-position.json --source 88060819
```

## 🔨 Build Instructions

### 1.Building from Source

```bash
# Navigate to project directory
cd trade-manager

# Download and verify dependencies
go mod download
go mod verify

# Build the CLI application
go build -o trade-cli cmd/cli/main.go

# Test the build
./trade-cli --help
```

### 2. Build with Script

```bash
# Standard build for current platform
./build.sh build

# Optimized build (smaller binary)
./build.sh build-optimized

# Cross-compile for multiple platforms
./build.sh cross

# Build Docker image
./build.sh docker

# Run tests only
./build.sh test

# Full build process (clean, deps, test, build, cross, docker)
./build.sh all
```

## ⚙️ Configuration Guide

### Configuration File Structure

```json
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
  "logging": {
    "level": "info",
    "file_path": "logs/trade-manager.log",
    "max_size": 100,
    "max_backups": 10,
    "max_age": 30,
    "compress": true
  },
  "same_position": {
    "source_account": {
      "XyBusinAccount": "88060819",
      "stockAccount": "0680048595",
      "clientType": "1",
      "enabled": true
    },
    "target_accounts": [
      {
        "XyBusinAccount": "88060820",
        "stockAccount": "0680048596",
        "clientType": "1",
        "enabled": false
      }
    ],
    "order_template": {
      "amount": 1000,
      "order_type": "1",
      "direction": "1",
      "valid_end_time_str": "2023-12-31",
      "min_amount": 100,
      "max_amount": 10000,
      "concurrency": 5
    }
  }
}
```

### Environment Variable Support

- `TRADE_API_URL` - Override API service URL
- `TRADE_API_TOKEN` - Set API authentication token
- `TRADE_MAX_WORKERS` - Set maximum worker threads

## 📊 Core Components

### 1. Trade Manager

**Main Methods:**

- `QueryAccountPositions()` - Query account positions
- `SamePositionOrderFlow()` - Same position order flow
- `BatchOrder()` - Batch order placement
- `GetPositionSummary()` - Get position summary

### 2. HTTP Client

**Supported Endpoints:**

- `POST /api/v1/positions` - Position query
- `POST /api/v1/orders` - Single order placement
- `POST /api/v1/orders/batch` - Batch order placement

### 3. Configuration Management

**Configuration Validation:**

- `ValidateSamePositionConfig()` - Validate same position configuration
- `GetSourceAccount()` - Get source account
- `GetTargetAccounts()` - Get target account list

### 4. Logging System

**Log Levels:** DEBUG, INFO, WARN, ERROR, FATAL

**Features:**

- File and console dual output
- Log rotation and backup management
- Structured logging
- Thread-safe writing

## 🔧 Command Line Usage

### Basic Commands

```bash
# Show help information
trade-cli --help

# Initialize configuration file
trade-cli --init

# List accounts in configuration
trade-cli --list-accounts

# Query positions
trade-cli --query --source 88060819

# Same position order flow
trade-cli --same-position --config config.json --source 88060819
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--config` | Configuration file path | `config.json` |
| `--query` | Query positions | `false` |
| `--same-position` | Same position order flow | `false` |
| `--source` | Source account ID | empty |
| `--min-amount` | Minimum position quantity | `0` |
| `--max-amount` | Maximum position quantity | `0` |
| `--concurrency` | Concurrency level | `5` |
| `--init` | Initialize configuration file | `false` |
| `--list-accounts` | List accounts in configuration | `false` |

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestQueryPositions
```

### Integration Tests

The project includes a complete integration testing framework supporting:

- Mock server testing
- Error scenario testing
- Concurrent performance testing
- Edge case testing

## 📈 Performance Features

### Concurrent Processing

- Configurable concurrent worker threads
- Intelligent task distribution mechanism
- Timeout and error handling

### Memory Optimization

- Stream data processing
- Reasonable caching strategy
- Resource cleanup mechanism

### Network Optimization

- Connection pool management
- Request retry mechanism
- Timeout control

## 🔒 Security Features

### Data Security

- Configuration file encryption support
- API token security management
- Sensitive information log filtering

### Operational Security

- Operation confirmation mechanism
- Error recovery capability
- Audit log recording

## 🐛 Troubleshooting

### Common Issues

1. **Connection Failure**
   - Check API service address and port
   - Verify network connection
   - Check firewall settings

2. **Authentication Failure**
   - Verify API token
   - Check permission settings
   - Confirm account status

3. **Order Placement Failure**
   - Check stock code format
   - Verify order quantity
   - Confirm trading hours

### Log Analysis

Log file location: `logs/trade-manager.log`

```bash
# View real-time logs
tail -f logs/trade-manager.log

# Search error logs
grep "ERROR" logs/trade-manager.log

# Filter logs by time
grep "2023-" logs/trade-manager.log
```

## 🤝 Contributing Guide

### Development Environment Setup

1. Fork the project repository
2. Create feature branch
3. Commit code changes
4. Create Pull Request

### Code Standards

- Follow Go language coding standards
- Add necessary comments and documentation
- Write unit tests
- Ensure code passes all tests

## 📄 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## 📞 Support

For questions or suggestions, please contact through:

- Create an Issue
- Send email
- Submit Pull Request

## 🔄 Changelog

### v1.0.0 (2023-12-01)

- Initial version release
- Support multi-account position query
- Implement same position order flow
- Add complete logging system
- Provide integration testing framework
