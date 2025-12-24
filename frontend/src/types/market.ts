export interface KLine {
  s: string; // symbol
  e: string; // exchange
  p: string; // period
  o: string; // open
  h: string; // high
  l: string; // low
  c: string; // close
  v: string; // volume
  t: string; // timestamp (RFC3339 or ISO)
}

export interface Trade {
  symbol: string;
  exchange: string;
  price: string;
  amount: string;
  timestamp: string;
  side: string;
}

export interface StrategySignal {
  strategy: string;
  symbol: string;
  period: string;
  action: 'buy' | 'sell' | 'hold';
  price: string;
  time: string;
}
