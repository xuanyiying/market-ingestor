# Quant-Trader - 专业级量化交易基础设施

`quant-trader` 是一个高性能量化交易引擎，专为**高并发、低延迟和机构级安全**而设计。它提供了从实时行情接入到隔离策略执行及模拟交易的完整流水线。

---

## 🚀 核心特性

### 1. 高性能行情流水线

- **多交易所接入**: 原生支持 Binance, OKX, Bybit 等 WebSocket 接入，具备自动容灾重连机制。
- **微秒级聚合**: 基于内存窗口管理实现毫秒级 K 线生成。
- **事件驱动架构**: 基于 NATS JetStream 实现可靠的异步数据分发。
- **TimescaleDB 存储**: 针对时序数据优化的存储方案，支持自动分区与压缩。

### 2. 商业级交易套件

- **模拟交易引擎 (Paper Trading)**: 低延迟撮合模拟，支持多资产余额追踪。
- **风险管理 (Risk Management)**: 事前风扣引擎（持仓限制、回撤保护等）。
- **策略市场 (Marketplace)**: 基于订阅制的策略分发，集成 Stripe 支付。
- **阶梯式频率限制**: 针对不同会员等级 (Free/Pro/Enterprise) 的多租户 API 限流。

### 3. 高级策略实现

- **WASM 沙箱隔离**: 使用 `wazero` 实现安全、隔离的策略执行环境。
- **通用指标库**: 内置 RSI, MACD, Bollinger Bands 等常用量化指标。
- **预警系统**: 灵活的基于规则的通知引擎，支持价格与技术指标触发。

---

## 🛠 技术栈

- **后端**: Golang (Go 1.24+), Gin, GORM
- **数据库**: TimescaleDB (PostgreSQL 16+)
- **消息队列**: NATS JetStream
- **安全**: WebAssembly (Wasm)
- **支付**: Stripe API
- **前端**: React, Vite, ECharts, TailwindCSS

---

## 🏗 系统架构

`quant-trader` 采用模块化、解耦的架构设计：

```mermaid
graph TD
    subgraph "外部交易所"
        Binance[Binance WS]
        OKX[OKX WS]
    end

    subgraph "接入层 (Go)"
        Connectors[行情接入器]
        Raw[NATS: market.raw]
    end

    subgraph "核心引擎"
        KP[K线处理器]
        AE[预警引擎]
        PE[模拟引擎]
        WR[Wasm 运行环境]
    end

    subgraph "存储与接口"
        TSDB[(TimescaleDB)]
        HTTP[Gin REST API]
        WS[推送网关]
    end

    Binance & OKX --> Connectors
    Connectors --> Raw
    Raw --> KP & AE
    KP --> PE & TSDB & WS
    PE --> WR
    HTTP --> TSDB & PE
```

---

## 🏁 快速开始

### 1. 环境要求

- Go 1.24+
- Docker & Docker Compose
- NATS Server

### 2. 安装

```bash
git clone https://github.com/your-repo/quant-trader.git
cd quant-trader/backend
go mod download
```

### 3. 配置

将 `config.yaml.example` 复制为 `config.yaml` 并配置相关凭据：

```yaml
database:
  url: "postgres://user:pass@localhost:5432/quant_trader"
nats:
  url: "nats://localhost:4222"
stripe:
  key: "sk_test_..."
```

### 4. 运行系统

```bash
# 启动基础组件
docker-compose up -d

# 启动后端
go run cmd/main.go
```

---

## 📊 性能基准

| 层级 | 延迟 (P99) | 吞吐量 |
| :--- | :--- | :--- |
| **行情接入** | < 2ms | 50,000 trades/s |
| **K线生成** | < 5ms | 100 symbols (1m period) |
| **撮合模拟** | < 10ms | 1,000 orders/s |
| **持久化** | < 20ms | 10,000 records/batch |

---

## ⚖️ 开源协议

基于 MIT 协议进行分发。详见 `LICENSE`。

---
*Quant-Trader - 用专业的基础设施赋能您的交易策略。*
