# Quant Trader (量化交易系统)

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![NATS](https://img.shields.io/badge/NATS-JetStream-37A546?style=flat&logo=nats)](https://nats.io/)
[![TimescaleDB](https://img.shields.io/badge/TimescaleDB-PostgreSQL-FDB515?style=flat&logo=postgresql)](https://www.timescale.com/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)](https://www.docker.com/)

Quant Trader 是一个基于 **微服务架构 (Microservices)** 与 **事件驱动 (Event-Driven)** 模式构建的高性能加密货币量化交易系统。项目旨在满足高并发行情接入、低延迟数据处理、实时策略执行及大规模回测的需求。

---

## 🏗 系统架构

系统采用生产者-消费者模型，通过 **NATS JetStream** 进行解耦，主要包含以下核心服务：

1.  **Market Ingestor (行情接入服务)**
    *   负责连接各大交易所 (Binance, OKX, Bybit, Coinbase, Kraken) 的 WebSocket 接口。
    *   进行协议适配与数据清洗，将异构数据标准化为统一格式。
    *   支持断线自动重连 (Exponential Backoff) 与心跳保活。
2.  **Stream Processor (流处理服务)**
    *   订阅实时 Tick 数据，实时聚合生成各周期 K 线 (1m, 5m 等)。
    *   计算实时技术指标 (RSI, MA 等)。
3.  **Persistence Service (持久化服务)**
    *   消费 NATS 消息队列，采用 Batch Insert (批量插入) 策略写入 TimescaleDB。
    *   利用 Hypertable 自动分区技术，高效存储海量金融时序数据。
4.  **Push Gateway (推送网关)**
    *   维护客户端 WebSocket 长连接池。
    *   实现基于 Topic 的订阅/发布模型，将实时行情低延迟广播给前端或下游策略。
5.  **Backtest Engine (回测引擎)**
    *   纯 Go 实现的高性能回测核心。
    *   支持多策略配置、资金模拟、滑点/手续费计算。
    *   输出详细的绩效报告 (Win Rate, Max Drawdown, Sharpe Ratio)。
6.  **API Server**
    *   提供 RESTful API，用于历史数据查询、回测任务提交与结果检索。

---

## 🛠 技术选型

*   **编程语言**: Go (Golang)
*   **消息中间件**: NATS JetStream (低延迟、高性能、支持持久化)
*   **数据库**: TimescaleDB (基于 PostgreSQL 的时序数据库)
*   **缓存**: Redis (用于热数据、会话管理)
*   **精度处理**: `shopspring/decimal` (杜绝浮点数精度丢失)
*   **并发模型**: Go Worker Pool + Channels

---

## 🚀 功能特性

*   **多交易所支持**: 已接入 Binance, OKX, Bybit, Coinbase, Kraken。
*   **高精度计算**: 全链路采用 Decimal 类型，确保金额与价格零误差。
*   **实时聚合**: 基于时间窗口的流式 K 线生成算法。
*   **高性能存储**: 针对时序数据优化的数据库 Schema 设计。
*   **健壮性**: 完善的错误处理、重连机制与优雅关闭 (Graceful Shutdown)。
*   **可观测性**: 集成 Prometheus 监控指标 (连接数、处理速率、DB 延迟)。

---

## 📂 目录结构

```
quant-trader/
├── market-ingestor/
│   ├── cmd/                # 程序入口
│   ├── internal/
│   │   ├── api/            # HTTP API Handler
│   │   ├── config/         # 配置管理
│   │   ├── connector/      # 交易所连接器 (Binance, OKX...)
│   │   ├── engine/         # 回测引擎核心
│   │   ├── infrastructure/ # 基础设施 (DB, NATS, Logger)
│   │   ├── model/          # 数据模型定义
│   │   ├── processor/      # 流处理器 (K线聚合)
│   │   ├── push/           # WebSocket 推送网关
│   │   ├── storage/        # 数据持久化
│   │   └── strategy/       # 交易策略实现
│   └── scripts/            # 数据库初始化脚本
├── docker-compose.yml      # 容器编排
└── README.md               # 项目文档
```

---

## 💾 数据库设计 (Schema)

核心表结构设计如下 (TimescaleDB Hypertable)：

### 1. 原始成交记录 (market_trades)
```sql
CREATE TABLE market_trades (
    time        TIMESTAMPTZ NOT NULL,
    symbol      TEXT NOT NULL,
    exchange    TEXT NOT NULL,
    price       NUMERIC NOT NULL,
    amount      NUMERIC NOT NULL,
    side        TEXT,
    trade_id    TEXT
);
SELECT create_hypertable('market_trades', 'time');
```

### 2. K 线数据 (market_candles)
```sql
CREATE TABLE market_candles (
    time        TIMESTAMPTZ NOT NULL,
    symbol      TEXT NOT NULL,
    exchange    TEXT NOT NULL,
    period      TEXT NOT NULL,
    open        NUMERIC,
    high        NUMERIC,
    low         NUMERIC,
    close       NUMERIC,
    volume      NUMERIC
);
SELECT create_hypertable('market_candles', 'time');
```

---

## 🗓 开发路线图 (Roadmap)

### Phase 1: 基础设施与数据接入 (Completed) ✅
- [x] 项目初始化与 Docker 环境搭建
- [x] 定义核心数据模型 (Decimal 精度)
- [x] 开发 Market Ingestor (Binance, OKX, Bybit, Coinbase, Kraken)
- [x] 实现 TimescaleDB 批量写入

### Phase 2: 实时流处理与分发 (Completed) ✅
- [x] 集成 NATS JetStream
- [x] 实现 1m K 线实时聚合算法
- [x] 开发 WebSocket Push Gateway (订阅/广播)

### Phase 3: 回测引擎 (Completed) ✅
- [x] 定义策略接口 (Strategy Interface)
- [x] 实现简单移动平均 (SMA) 策略
- [x] 开发回测核心 (撮合、资金管理、绩效统计)

### Sprint 4: API & UI (Completed ✅)
- **Gin API Server**: Integrated Gin framework with JWT authentication.
- **Monitoring**: Prometheus metrics (latency, connections, insert rate) and Grafana dashboard.
- **Web UI**: Simple Vue.js + ECharts dashboard for real-time monitoring and history viewing.

---

## ⚡️ 快速开始

### 1. 启动基础设施
```bash
docker-compose up -d
```

### 2. 运行服务
```bash
cd market-ingestor
go run cmd/main.go
```

### 3. 测试 API
*   **获取历史 K 线**: `GET /api/v1/klines/BTCUSDT?period=1m`
*   **运行回测**: `POST /api/v1/backtest`
*   **WebSocket 订阅**: `ws://localhost:8080/ws`

---

## 🧪 测试

```bash
go test ./...
```
目前已覆盖 Connector, Processor, Storage, Engine 等核心模块的单元测试。
