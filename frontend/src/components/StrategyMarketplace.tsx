import React, { useState, useEffect } from 'react';
import { ShoppingCart, Star, TrendingUp, Info, CheckCircle2 } from 'lucide-react';
import axios from 'axios';

interface Strategy {
    id: number;
    name: string;
    description: string;
    price: number;
    author: string;
    is_subscribed: boolean;
}

const StrategyMarketplace: React.FC = () => {
    const [strategies, setStrategies] = useState<Strategy[]>([]);
    const [loading, setLoading] = useState(true);

    const loadStrategies = async () => {
        try {
            const response = await axios.get('/api/v1/market/strategies', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            setStrategies(response.data);
        } catch (error) {
            console.error('Failed to load strategies:', error);
        } finally {
            setLoading(false);
        }
    };

    const handlePurchase = async (id: number) => {
        try {
            await axios.post(`/api/v1/market/strategies/${id}/purchase`, {}, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            alert('Strategy purchased successfully!');
            loadStrategies();
        } catch (error) {
            alert('Purchase failed');
        }
    };

    useEffect(() => {
        loadStrategies();
    }, []);

    return (
        <div className="bg-card p-6 rounded-xl shadow-lg border border-gray-800">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                    <ShoppingCart size={20} className="text-yellow-500" />
                    <h2 className="font-bold">Strategy Marketplace</h2>
                </div>
                <div className="flex items-center gap-4 text-[10px] text-gray-400 font-bold uppercase">
                    <div className="flex items-center gap-1"><Star size={12} className="text-yellow-500 fill-yellow-500" /> TOP-RATED</div>
                    <div className="flex items-center gap-1"><TrendingUp size={12} className="text-blue-500" /> TRENDING</div>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {loading ? (
                    Array(2).fill(0).map((_, i) => (
                        <div key={i} className="h-40 bg-gray-900/40 rounded-xl animate-pulse border border-gray-800"></div>
                    ))
                ) : strategies.length === 0 ? (
                    <div className="col-span-2 py-12 text-center bg-gray-900/20 border border-dashed border-gray-800 rounded-xl">
                        <Info size={24} className="mx-auto text-gray-700 mb-2" />
                        <p className="text-gray-500 text-xs text-semibold">No strategies available currently.</p>
                    </div>
                ) : (
                    strategies.map((s) => (
                        <div key={s.id} className="bg-gray-900/50 p-5 rounded-xl border border-gray-800 hover:border-blue-500/50 transition-all hover:bg-gray-900/80 group">
                            <div className="flex justify-between items-start mb-3">
                                <div className="space-y-1">
                                    <h3 className="font-black text-gray-200 group-hover:text-blue-400 transition-colors uppercase tracking-tight">{s.name}</h3>
                                    <p className="text-[10px] text-gray-400 font-bold italic">by {s.author}</p>
                                </div>
                                <div className="bg-blue-600/10 text-blue-400 text-[10px] font-black px-2 py-1 rounded border border-blue-500/20">
                                    {s.price === 0 ? 'FREE' : `$${s.price}`}
                                </div>
                            </div>
                            <p className="text-xs text-gray-500 line-clamp-2 mb-4 leading-relaxed group-hover:text-gray-400 transition-colors">
                                {s.description || "A high-performance algorithmic strategy designed for volatile markets. Includes risk management and auto-stop loss."}
                            </p>
                            <div className="flex justify-between items-center">
                                <div className="flex gap-1">
                                    <span className="text-[10px] bg-gray-800 px-2 py-0.5 rounded text-gray-500 font-bold">RSI</span>
                                    <span className="text-[10px] bg-gray-800 px-2 py-0.5 rounded text-gray-500 font-bold">1H</span>
                                </div>
                                {s.is_subscribed ? (
                                    <div className="flex items-center gap-1 text-[10px] text-green-500 font-black">
                                        <CheckCircle2 size={12} /> SUBSCRIBED
                                    </div>
                                ) : (
                                    <button
                                        onClick={() => handlePurchase(s.id)}
                                        className="bg-blue-600 hover:bg-blue-700 text-white text-[10px] font-black px-4 py-2 rounded-md transition-all active:scale-95 shadow-lg shadow-blue-900/20"
                                    >
                                        SUBSCRIBE
                                    </button>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default StrategyMarketplace;
