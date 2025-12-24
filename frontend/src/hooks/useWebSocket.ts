import { useEffect, useRef } from 'react';
import { useMarketStore } from '../store/useMarketStore';

export const useWebSocket = () => {
  const ws = useRef<WebSocket | null>(null);
  const { 
    symbol, 
    period, 
    setConnectionStatus, 
    updateKLine, 
    setLastTrade,
    addSignal 
  } = useMarketStore();

  const connect = () => {
    if (ws.current) ws.current.close();

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host === 'localhost:5173' ? 'localhost:8080' : window.location.host;
    const url = `${protocol}//${host}/ws`;

    ws.current = new WebSocket(url);

    ws.current.onopen = () => {
      setConnectionStatus('Connected');
      subscribe();
    };

    ws.current.onclose = () => {
      setConnectionStatus('Disconnected');
      setTimeout(connect, 5000);
    };

    ws.current.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        
        // Handle KLine
        if (msg.c && msg.t && msg.s) {
          updateKLine(msg);
        }
        
        // Handle Trade
        if (msg.price && msg.symbol) {
          setLastTrade(msg);
        }

        // Handle Signal (Custom message from strategy engine)
        if (msg.strategy && msg.action) {
          addSignal(msg);
        }
      } catch (e) {
        console.error('WS parse error:', e);
      }
    };
  };

  const subscribe = () => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      // Subscribe K-Line
      ws.current.send(JSON.stringify({
        action: 'subscribe',
        topic: `market.kline.${period}.${symbol}`
      }));
      // Subscribe Trade
      ws.current.send(JSON.stringify({
        action: 'subscribe',
        topic: `market.raw.*.${symbol}`
      }));
      // Subscribe Signals
      ws.current.send(JSON.stringify({
        action: 'subscribe',
        topic: `strategy.signal.*.${symbol}`
      }));
    }
  };

  useEffect(() => {
    connect();
    return () => {
      if (ws.current) ws.current.close();
    };
  }, []);

  useEffect(() => {
    subscribe();
  }, [symbol, period]);

  return ws.current;
};
