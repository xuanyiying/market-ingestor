import React, { useState, useEffect } from 'react';
import { Activity, RefreshCw, TrendingUp, Settings, Database, Play } from 'lucide-react';
import Chart from './components/Chart';
import AlertsManager from './components/AlertsManager';
import StrategyMarketplace from './components/StrategyMarketplace';
import PortfolioReport from './components/PortfolioReport';
import { useMarketStore } from './store/useMarketStore';
import { useWebSocket } from './hooks/useWebSocket';
import axios from 'axios';

const App: React.FC = () => {
  const {
    symbol,
    period,
    setSymbol,
    setPeriod,
    setKLines,
    lastTrade,
    connectionStatus,
    signals,
    alerts,
    setAlerts,
    subscription,
    setSubscription
  } = useMarketStore();

  const [inputSymbol, setInputSymbol] = useState(symbol);
  const [paperAccount, setPaperAccount] = useState<any>(null);
  const [positions, setPositions] = useState<any[]>([]);
  const [orderQty, setOrderQty] = useState('1');
  const [loading, setLoading] = useState(true);

  useWebSocket();

  const loadHistory = async () => {
    try {
      const response = await axios.get(`/api/v1/klines/${symbol}?period=${period}`);
      const data = response.data;
      if (Array.isArray(data)) {
        const mappedData = data.map((k: any) => ({
          s: k.symbol,
          e: k.exchange,
          p: k.period,
          o: k.open,
          h: k.high,
          l: k.low,
          c: k.close,
          v: k.volume,
          t: k.time || k.timestamp
        })).sort((a: any, b: any) => new Date(a.t).getTime() - new Date(b.t).getTime());
        setKLines(mappedData);
      }
    } catch (error) {
      console.error('Failed to load history:', error);
    }
  };

  const loadSubscription = async () => {
    try {
      const response = await axios.get('/api/v1/subscription', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      setSubscription(response.data);
    } catch (error) {
      console.error('Failed to load subscription:', error);
    }
  };

  const loadAlerts = async () => {
    try {
      const response = await axios.get('/api/v1/alerts', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      setAlerts(response.data);
    } catch (error) {
      console.error('Failed to load alerts:', error);
    }
  };

  const loadPaperAccount = async () => {
    try {
      const response = await axios.get('/api/v1/paper/account', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      setPaperAccount(response.data);
    } catch (error) {
      console.error('Failed to load paper account:', error);
    }
  };

  const loadPositions = async () => {
    try {
      const response = await axios.get('/api/v1/paper/positions', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      setPositions(response.data);
    } catch (error) {
      console.error('Failed to load positions:', error);
    }
  };

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      await Promise.all([
        loadHistory(),
        loadSubscription(),
        loadAlerts(),
        loadPaperAccount(),
        loadPositions()
      ]);
      setLoading(false);
    };
    fetchData();
  }, [symbol, period]);

  const handleUpdateSymbol = () => {
    setSymbol(inputSymbol.toUpperCase());
  };

  const handleCreateOrder = async (side: 'buy' | 'sell') => {
    try {
      await axios.post('/api/v1/paper/orders', {
        symbol: symbol,
        side: side,
        type: 'market',
        qty: parseFloat(orderQty)
      }, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      alert('Order placed successfully');
      loadPaperAccount();
      loadPositions();
    } catch (error) {
      alert('Order failed: ' + ((error as any).response?.data?.error || error));
    }
  };

  const handleUpgrade = async () => {
    try {
      const response = await axios.post('/api/v1/subscription/checkout', {
        price_id: 'price_pro_default'
      }, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      window.location.href = response.data.url;
    } catch (error) {
      alert('Failed to start checkout');
    }
  };

  const handleTriggerBackfill = async () => {
    try {
      await axios.post('/api/v1/backfill', {
        exchange: 'binance',
        symbol: symbol,
        start_time: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        end_time: new Date().toISOString()
      }, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
      });
      alert('Backfill task started');
    } catch (error) {
      console.error('Backfill failed:', error);
      alert('Failed to start backfill');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <Activity size={48} className="text-blue-600 animate-spin" />
          <p className="text-gray-500 font-bold uppercase tracking-widest animate-pulse">Initializing Terminal...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-gray-200 selection:bg-blue-500/30">
      <div className="max-w-[1600px] mx-auto p-4 lg:p-8 space-y-8">
        {/* Header */}
        <header className="flex flex-col md:flex-row justify-between items-center gap-6 bg-card p-6 rounded-2xl shadow-xl border border-gray-800/50 backdrop-blur-sm">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-blue-600 rounded-xl shadow-lg shadow-blue-900/20">
              <TrendingUp size={28} className="text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-black tracking-tight text-white">QUANT<span className="text-blue-500">TRADER</span></h1>
              <div className="flex items-center gap-2">
                <p className="text-[10px] text-gray-500 uppercase tracking-widest font-black">
                  {subscription ? `${subscription.tier_name} MEMBER` : 'LOADING...'}
                </p>
                {subscription?.tier_name !== 'Free' ? (
                  <span className="bg-yellow-500/10 text-yellow-500 text-[8px] px-1.5 py-0.5 rounded font-black border border-yellow-500/20 uppercase">Pro Access</span>
                ) : (
                  <button onClick={handleUpgrade} className="bg-blue-600/10 text-blue-400 text-[8px] px-1.5 py-0.5 rounded font-black border border-blue-500/20 hover:bg-blue-600 hover:text-white transition-all uppercase">Upgrade Now</button>
                )}
              </div>
            </div>
          </div>

          <div className="flex flex-wrap justify-center gap-4">
            <div className="bg-gray-900/50 px-5 py-3 rounded-xl border border-gray-800 flex flex-col items-center min-w-[140px]">
              <span className="text-[10px] text-gray-500 uppercase font-black tracking-tighter mb-1">Paper Balance</span>
              <span className="text-lg font-black text-up flex items-center gap-2">
                ${paperAccount ? parseFloat(paperAccount.balance).toLocaleString() : '0,00'}
              </span>
            </div>
            <div className="bg-gray-900/50 px-5 py-3 rounded-xl border border-gray-800 flex flex-col items-center min-w-[140px]">
              <span className="text-[10px] text-gray-500 uppercase font-black tracking-tighter mb-1">Status</span>
              <span className={`text-sm font-black flex items-center gap-2 ${connectionStatus === 'Connected' ? 'text-up' : 'text-down'}`}>
                <Activity size={16} />
                {connectionStatus.toUpperCase()}
              </span>
            </div>
            <div className="bg-gray-900/50 px-5 py-3 rounded-xl border border-gray-800 flex flex-col items-center min-w-[140px]">
              <span className="text-[10px] text-gray-500 uppercase font-black tracking-tighter mb-1">Market Price</span>
              <span className="text-lg font-black text-blue-400">
                {lastTrade ? parseFloat(lastTrade.price).toFixed(2) : '--.--'} <span className="text-[10px] ml-1 text-gray-600 font-bold">USDT</span>
              </span>
            </div>
          </div>
        </header>

        <div className="grid grid-cols-12 gap-8">
          {/* Main Content Area */}
          <div className="col-span-12 lg:col-span-9 space-y-8">
            {/* Chart Section */}
            <div className="bg-card p-6 rounded-2xl shadow-xl border border-gray-800/50">
              <div className="flex flex-wrap justify-between items-center gap-4 mb-8">
                <div className="flex items-center gap-2 bg-gray-900 px-2 py-1.5 rounded-xl border border-gray-800">
                  <input
                    value={inputSymbol}
                    onChange={(e) => setInputSymbol(e.target.value)}
                    className="bg-transparent border-none outline-none px-3 py-1 text-sm font-black w-28 text-white placeholder-gray-700"
                    placeholder="BTCUSDT"
                  />
                  <button
                    onClick={handleUpdateSymbol}
                    className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded-lg text-[10px] font-black transition-all active:scale-95 shadow-lg shadow-blue-900/20"
                  >
                    SYNC
                  </button>
                </div>

                <div className="flex gap-1.5 bg-gray-900 p-1 rounded-xl border border-gray-800">
                  {['1m', '5m', '15m', '1h', '4h', '1d'].map(p => (
                    <button
                      key={p}
                      onClick={() => setPeriod(p)}
                      className={`px-4 py-1.5 rounded-lg text-[10px] font-black transition-all ${period === p
                        ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20'
                        : 'text-gray-500 hover:text-gray-300 hover:bg-gray-800'
                        }`}
                    >
                      {p.toUpperCase()}
                    </button>
                  ))}
                </div>

                <div className="flex gap-2">
                  <button onClick={loadHistory} className="bg-gray-900 hover:bg-gray-800 text-gray-400 p-2.5 rounded-xl border border-gray-800 transition-all font-bold" title="Refresh">
                    <RefreshCw size={18} />
                  </button>
                  <button onClick={handleTriggerBackfill} className="bg-gray-900 hover:bg-gray-800 text-gray-400 p-2.5 rounded-xl border border-gray-800 transition-all font-bold" title="Backfill">
                    <Database size={18} />
                  </button>
                  <button className="bg-gray-900 hover:bg-gray-800 text-gray-400 p-2.5 rounded-xl border border-gray-800 transition-all font-bold" title="Settings">
                    <Settings size={18} />
                  </button>
                </div>
              </div>

              <div className="h-[500px] w-full">
                <Chart />
              </div>
            </div>

            {/* Trading Panel */}
            <div className="bg-card p-6 rounded-2xl shadow-xl border border-gray-800/50">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                  <div className="p-2 bg-blue-500/10 rounded-lg">
                    <Play size={20} className="text-blue-500" />
                  </div>
                  <h2 className="text-lg font-black uppercase tracking-tight">Simulator: {symbol}</h2>
                </div>
                <span className="text-[10px] font-mono text-gray-500 bg-gray-900 px-3 py-1 rounded-lg border border-gray-800">
                  MARKET PRICE: {lastTrade ? parseFloat(lastTrade.price).toFixed(2) : '0.00'}
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-12 gap-8">
                <div className="md:col-span-4 space-y-6 pr-6 border-r border-gray-800/50">
                  <div className="space-y-2">
                    <label className="text-[10px] text-gray-500 uppercase font-black tracking-widest">Order Amount ({symbol.replace('USDT', '')})</label>
                    <input
                      type="number"
                      value={orderQty}
                      onChange={(e) => setOrderQty(e.target.value)}
                      className="w-full bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-sm font-black outline-none focus:ring-2 focus:ring-blue-600/50 transition-all"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <button
                      onClick={() => handleCreateOrder('buy')}
                      className="bg-up hover:bg-green-600 text-white font-black py-4 rounded-xl transition-all shadow-xl shadow-green-900/20 active:scale-95 text-xs tracking-widest"
                    >
                      BUY / LONG
                    </button>
                    <button
                      onClick={() => handleCreateOrder('sell')}
                      className="bg-down hover:bg-red-600 text-white font-black py-4 rounded-xl transition-all shadow-xl shadow-red-900/20 active:scale-95 text-xs tracking-widest"
                    >
                      SELL / SHORT
                    </button>
                  </div>
                  <div className="bg-gray-900/30 p-3 rounded-xl border border-dashed border-gray-800">
                    <p className="text-[10px] text-gray-600 italic leading-relaxed text-center">Market orders are subject to platform risk engine validation.</p>
                  </div>
                </div>

                <div className="md:col-span-8 space-y-4">
                  <div className="flex justify-between items-center">
                    <label className="text-[10px] text-gray-500 uppercase font-black tracking-widest">Active Inventory</label>
                    <span className="text-[10px] bg-blue-600/10 text-blue-400 px-2 py-0.5 rounded-full font-bold">{positions.length} Open</span>
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 max-h-[160px] overflow-y-auto pr-2 custom-scrollbar">
                    {positions.length === 0 ? (
                      <div className="col-span-2 py-10 text-center bg-gray-900/20 border border-dashed border-gray-800 rounded-xl">
                        <span className="text-xs text-gray-600 font-bold italic">No active positions for this asset</span>
                      </div>
                    ) : (
                      positions.map((pos, i) => (
                        <div key={i} className="bg-gray-900/40 p-4 rounded-xl border border-gray-800 flex justify-between items-center group hover:border-blue-500/30 transition-all">
                          <div>
                            <span className="text-xs font-black text-gray-200 uppercase">{pos.symbol}</span>
                            <div className="text-[10px] text-gray-500 font-bold mt-1">VOL: {parseFloat(pos.qty).toFixed(4)}</div>
                          </div>
                          <div className="text-right">
                            <div className="text-xs font-mono text-blue-400">@{parseFloat(pos.avg_price).toFixed(2)}</div>
                            <div className={`text-[10px] font-black mt-1 ${lastTrade && parseFloat(lastTrade.price) > parseFloat(pos.avg_price) ? 'text-up' : 'text-down'
                              }`}>
                              {lastTrade ? (
                                ((parseFloat(lastTrade.price) / parseFloat(pos.avg_price) - 1) * 100).toFixed(2) + '%'
                              ) : '-%'}
                            </div>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="col-span-12 lg:col-span-3 space-y-8">
            <StrategyMarketplace />
            <PortfolioReport />
            <AlertsManager
              alerts={alerts}
              symbol={symbol}
              onRefresh={loadAlerts}
            />

            <div className="bg-card p-6 rounded-2xl shadow-xl border border-gray-800/50">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                  <Activity size={18} className="text-blue-400" />
                  <h2 className="font-black uppercase tracking-tight">Signal Feed</h2>
                </div>
                <div className="flex items-center gap-1.5">
                  <div className="w-1.5 h-1.5 bg-up rounded-full animate-ping"></div>
                  <span className="text-[8px] font-black text-up uppercase">LIVE</span>
                </div>
              </div>
              <div className="space-y-4 overflow-y-auto max-h-[400px] pr-2 custom-scrollbar">
                {signals.length === 0 ? (
                  <div className="text-center py-12 text-gray-700 bg-gray-900/20 border border-dashed border-gray-800 rounded-xl">
                    <p className="text-xs font-bold italic">Listening for market triggers...</p>
                  </div>
                ) : (
                  signals.map((sig, i) => (
                    <div key={i} className="bg-gray-900/50 p-4 rounded-xl border border-gray-800 space-y-3 relative overflow-hidden group">
                      <div className={`absolute left-0 top-0 bottom-0 w-1 ${sig.action === 'buy' ? 'bg-up' : 'bg-down'}`}></div>
                      <div className="flex justify-between items-start">
                        <span className="text-[10px] font-black text-gray-400 uppercase tracking-tighter">{sig.strategy}</span>
                        <span className="text-[8px] text-gray-600 font-mono tracking-tighter">{new Date(sig.time).toLocaleTimeString()}</span>
                      </div>
                      <div className="flex justify-between items-center">
                        <span className="text-xs font-black text-white">{sig.symbol}</span>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest ${sig.action === 'buy' ? 'bg-up/10 text-up' : 'bg-down/10 text-down'
                          }`}>
                          {sig.action}
                        </span>
                      </div>
                      <div className="text-right text-[10px] font-mono text-gray-500 font-bold">
                        PRC: {parseFloat(sig.price).toFixed(2)}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;
