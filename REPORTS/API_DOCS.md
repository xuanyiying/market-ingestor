# Market Ingestor API 文档 (v1)

## 1. 认证接口

### 用户注册
- **POST** `/api/v1/register`
- **请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```
- **响应**: `201 Created`

### 用户登录
- **POST** `/api/v1/login`
- **请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```
- **响应**: `200 OK`, 返回 `token`。

---

## 2. 行情数据接口

### 获取历史 K 线
- **GET** `/api/v1/klines/:symbol`
- **参数**:
  - `period`: 周期 (1m, 5m, 1h等，默认 1m)
- **响应**: `200 OK`, K 线数组。

### 历史数据回填 (需认证)
- **POST** `/api/v1/backfill`
- **Header**: `Authorization: Bearer <token>`
- **请求体**:
```json
{
  "exchange": "binance",
  "symbol": "BTCUSDT",
  "start_time": "2025-12-01T00:00:00Z",
  "end_time": "2025-12-02T00:00:00Z"
}
```
- **响应**: `202 Accepted` (后台异步执行)

---

## 3. 策略与回测接口 (需认证)

### 运行策略回测
- **POST** `/api/v1/backtest`
- **Header**: `Authorization: Bearer <token>`
- **请求体**:
```json
{
  "symbol": "BTCUSDT",
  "strategy_type": "ma_cross_v2",
  "config": {
    "short_period": 5,
    "long_period": 20
  },
  "initial_balance": 10000,
  "start_time": "2025-12-01T00:00:00Z",
  "end_time": "2025-12-20T00:00:00Z"
}
```
- **响应**: `200 OK`, 返回详细回测报告 (收益率、回撤、交易日志等)。

---

## 4. WebSocket 推送

### 实时行情订阅
- **WS** `/ws`
- **协议**: 连接后无需特殊指令，系统会自动推送已配置交易对的实时 K 线数据。
- **消息格式**: `model.KLine` 的 JSON 字符串。
