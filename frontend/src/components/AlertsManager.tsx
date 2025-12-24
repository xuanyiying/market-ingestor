import React, { useState } from 'react';
import { Bell, Plus, Trash2, ShieldAlert } from 'lucide-react';
import axios from 'axios';

interface Alert {
    id: number;
    symbol: string;
    condition_type: string;
    target_value: number;
}

interface Props {
    alerts: Alert[];
    symbol: string;
    onRefresh: () => void;
}

const AlertsManager: React.FC<Props> = ({ alerts, symbol, onRefresh }) => {
    const [alertPrice, setAlertPrice] = useState('');
    const [alertType, setAlertType] = useState('price_above');
    const [loading, setLoading] = useState(false);

    const handleCreateAlert = async () => {
        if (!alertPrice) return;
        setLoading(true);
        try {
            await axios.post('/api/v1/alerts', {
                symbol: symbol,
                condition_type: alertType,
                target_value: parseFloat(alertPrice)
            }, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            setAlertPrice('');
            onRefresh();
        } catch (error) {
            alert('Failed to create alert');
        } finally {
            setLoading(false);
        }
    };

    const handleDeleteAlert = async (id: number) => {
        try {
            await axios.delete(`/api/v1/alerts/${id}`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            onRefresh();
        } catch (error) {
            alert('Failed to delete alert');
        }
    };

    return (
        <div className="bg-card p-5 rounded-xl shadow-lg border border-gray-800">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                    <Bell size={18} className="text-blue-400" />
                    <h2 className="font-bold">Alert Management</h2>
                </div>
                <span className="text-[10px] text-gray-500 font-bold uppercase">{symbol}</span>
            </div>

            {/* Quick Alert Form */}
            <div className="mb-6 p-4 bg-gray-900/50 rounded-lg border border-gray-800 space-y-3">
                <div className="flex gap-2">
                    <div className="flex-1 space-y-1">
                        <label className="text-[8px] text-gray-500 uppercase font-black">Type</label>
                        <select
                            value={alertType}
                            onChange={(e) => setAlertType(e.target.value)}
                            className="w-full bg-gray-800 border-none rounded-md px-2 py-2 text-xs outline-none focus:ring-1 focus:ring-blue-500 transition-all cursor-pointer"
                        >
                            <option value="price_above">Price Above (&ge;)</option>
                            <option value="price_below">Price Below (&le;)</option>
                            <option value="strategy_signal">Strategy Signal</option>
                        </select>
                    </div>
                    <div className="w-24 space-y-1">
                        <label className="text-[8px] text-gray-500 uppercase font-black">Value</label>
                        <input
                            type="number"
                            value={alertPrice}
                            onChange={(e) => setAlertPrice(e.target.value)}
                            placeholder="Price"
                            className="w-full bg-gray-800 border-none rounded-md px-2 py-2 text-xs outline-none focus:ring-1 focus:ring-blue-500 transition-all"
                        />
                    </div>
                </div>
                <button
                    onClick={handleCreateAlert}
                    disabled={loading}
                    className="w-full bg-blue-600 hover:bg-blue-700 text-white text-xs font-black py-2 rounded-md transition-all shadow-lg shadow-blue-900/20 active:scale-95 flex items-center justify-center gap-2"
                >
                    {loading ? 'SETTING...' : <><Plus size={14} /> SET PRICE ALERT</>}
                </button>
            </div>

            <div className="space-y-2 overflow-y-auto max-h-[250px] pr-2 custom-scrollbar">
                <label className="text-[8px] text-gray-500 uppercase font-black block mb-2">Active Rules ({alerts.length})</label>
                {alerts.length === 0 ? (
                    <div className="text-center py-8 bg-gray-900/20 border border-dashed border-gray-800 rounded-lg">
                        <ShieldAlert size={20} className="mx-auto text-gray-700 mb-2" />
                        <p className="text-gray-600 text-[10px] italic">No active alerts for {symbol}</p>
                    </div>
                ) : (
                    alerts.map((a) => (
                        <div key={a.id} className="flex justify-between items-center p-3 bg-gray-900/30 rounded-lg border border-gray-800/50 group hover:border-blue-500/30 transition-all">
                            <div className="flex items-center gap-3">
                                <div className={`p-1.5 rounded ${a.condition_type === 'price_above' ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'}`}>
                                    <Bell size={12} />
                                </div>
                                <div className="flex flex-col">
                                    <span className="text-xs font-black text-gray-300">{a.symbol}</span>
                                    <span className="text-[10px] text-gray-500 font-mono">
                                        {a.condition_type === 'price_above' ? 'UPPERBOUND' : 'LOWERBOUND'} · {a.target_value}
                                    </span>
                                </div>
                            </div>
                            <button
                                onClick={() => handleDeleteAlert(a.id)}
                                className="p-1.5 text-gray-600 hover:text-red-400 hover:bg-red-500/10 rounded transition-all opacity-0 group-hover:opacity-100"
                            >
                                <Trash2 size={14} />
                            </button>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default AlertsManager;
