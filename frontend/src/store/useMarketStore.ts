import { create } from 'zustand';
import type { KLine, Trade, StrategySignal } from '../types/market';

interface MarketStore {
  symbol: string;
  period: string;
  klines: KLine[];
  lastTrade: Trade | null;
  signals: StrategySignal[];
  connectionStatus: 'Connected' | 'Disconnected' | 'Connecting';
  alerts: any[];
  subscription: any | null;

  setSymbol: (symbol: string) => void;
  setPeriod: (period: string) => void;
  setKLines: (klines: KLine[]) => void;
  updateKLine: (kline: KLine) => void;
  setLastTrade: (trade: Trade) => void;
  addSignal: (signal: StrategySignal) => void;
  setConnectionStatus: (status: 'Connected' | 'Disconnected' | 'Connecting') => void;
  setAlerts: (alerts: any[]) => void;
  setSubscription: (sub: any) => void;
}

export const useMarketStore = create<MarketStore>((set) => ({
  symbol: 'BTCUSDT',
  period: '1m',
  klines: [],
  lastTrade: null,
  signals: [],
  connectionStatus: 'Disconnected',
  alerts: [],
  subscription: null,

  setSymbol: (symbol) => set({ symbol }),
  setPeriod: (period) => set({ period }),
  setKLines: (klines) => set({ klines }),
  updateKLine: (kline) => set((state) => {
    if (kline.s.toUpperCase() !== state.symbol.toUpperCase()) return state;
    if (kline.p !== state.period) return state;

    const newKLines = [...state.klines];
    const lastIdx = newKLines.length - 1;

    if (lastIdx >= 0 && new Date(newKLines[lastIdx].t).getTime() === new Date(kline.t).getTime()) {
      newKLines[lastIdx] = kline;
    } else {
      newKLines.push(kline);
    }

    if (newKLines.length > 1000) newKLines.shift();
    return { klines: newKLines };
  }),
  setLastTrade: (trade) => set((state) => {
    if (trade.symbol.toUpperCase() !== state.symbol.toUpperCase()) return state;
    return { lastTrade: trade };
  }),
  addSignal: (signal) => set((state) => ({
    signals: [signal, ...state.signals].slice(0, 100)
  })),
  setConnectionStatus: (status) => set({ connectionStatus: status }),
  setAlerts: (alerts) => set({ alerts }),
  setSubscription: (subscription) => set({ subscription }),
}));
